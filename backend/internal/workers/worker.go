package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/services"
)

const (
	TaskOCRProcess  = "ocr:process"
	TaskAIAnalyze   = "ai:analyze"
	TaskCompressFile = "file:compress"
)

type OCRPayload struct {
	FileID     string `json:"file_id"`
	DocumentID string `json:"document_id"`
	FilePath   string `json:"file_path"`
	MimeType   string `json:"mime_type"`
}

type AIPayload struct {
	DocumentID string `json:"document_id"`
	Text       string `json:"text"`
}

type CompressPayload struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
	MimeType string `json:"mime_type"`
}

type WorkerServer struct {
	db          *sql.DB
	cfg         *config.Config
	ocrService  *services.OCRService
	aiService   *services.AIService
	driveService *services.DriveService
}

func NewWorkerServer(db *sql.DB, cfg *config.Config) *WorkerServer {
	ws := &WorkerServer{
		db:         db,
		cfg:        cfg,
		ocrService: services.NewOCRService(),
		aiService:  services.NewAIService(&cfg.AI),
	}
	ds, err := services.NewDriveService(&cfg.Google)
	if err == nil {
		ws.driveService = ds
	}
	return ws
}

func (w *WorkerServer) Start() error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     w.cfg.Redis.Addr(),
			Password: w.cfg.Redis.Password,
			DB:       w.cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskOCRProcess, w.HandleOCR)
	mux.HandleFunc(TaskAIAnalyze, w.HandleAIAnalyze)
	mux.HandleFunc(TaskCompressFile, w.HandleCompress)

	log.Println("Worker server started")
	return srv.Run(mux)
}

func (w *WorkerServer) createJob(taskType, status string, documentID, fileID *string) string {
	var jobID string
	w.db.QueryRow(`
		INSERT INTO jobs (task_type, status, document_id, file_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, taskType, status, documentID, fileID).Scan(&jobID)
	return jobID
}

func (w *WorkerServer) updateJobStatus(jobID, status, errorMsg string) {
	if jobID == "" {
		return
	}
	switch status {
	case "processing":
		w.db.Exec("UPDATE jobs SET status = $1, started_at = NOW(), attempts = attempts + 1 WHERE id = $2", status, jobID)
	case "completed":
		w.db.Exec("UPDATE jobs SET status = $1, completed_at = NOW() WHERE id = $2", status, jobID)
	case "failed":
		w.db.Exec("UPDATE jobs SET status = $1, error_message = $2, completed_at = NOW() WHERE id = $3", status, errorMsg, jobID)
	}
}

func (w *WorkerServer) HandleOCR(ctx context.Context, t *asynq.Task) error {
	var payload OCRPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal OCR payload: %w", err)
	}

	jobID := w.createJob(TaskOCRProcess, "processing", &payload.DocumentID, &payload.FileID)
	log.Printf("Processing OCR for file: %s (job: %s)", payload.FileID, jobID)

	// Update status to processing
	w.db.Exec("UPDATE files SET ocr_status = 'processing' WHERE id = $1", payload.FileID)

	// Download file from Drive if needed
	filePath := payload.FilePath
	if filePath == "" {
		// Download from Drive
		var driveFileID string
		w.db.QueryRow("SELECT drive_file_id FROM files WHERE id = $1", payload.FileID).Scan(&driveFileID)

		reader, err := w.driveService.DownloadFile(ctx, driveFileID)
		if err != nil {
			w.db.Exec("UPDATE files SET ocr_status = 'failed' WHERE id = $1", payload.FileID)
			w.updateJobStatus(jobID, "failed", err.Error())
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer reader.Close()

		tmpFile := filepath.Join(os.TempDir(), payload.FileID)
		f, err := os.Create(tmpFile)
		if err != nil {
			w.updateJobStatus(jobID, "failed", err.Error())
			return err
		}
		io.Copy(f, reader)
		f.Close()
		filePath = tmpFile
		defer os.Remove(tmpFile)
	}

	// Extract text
	var text string
	var err error
	if payload.MimeType == "application/pdf" {
		text, err = w.ocrService.ExtractTextFromPDF(ctx, filePath)
	} else {
		text, err = w.ocrService.ExtractText(ctx, filePath)
	}

	if err != nil {
		w.db.Exec("UPDATE files SET ocr_status = 'failed' WHERE id = $1", payload.FileID)
		w.updateJobStatus(jobID, "failed", err.Error())
		return fmt.Errorf("OCR failed: %w", err)
	}

	// Save OCR text
	w.db.Exec("UPDATE files SET ocr_text = $1, ocr_status = 'completed' WHERE id = $2", text, payload.FileID)

	// Also update document ocr_text (append)
	w.db.Exec(`UPDATE documents SET ocr_text = CONCAT(ocr_text, E'\n', $1), updated_at = NOW() WHERE id = $2`,
		text, payload.DocumentID)

	log.Printf("OCR completed for file: %s", payload.FileID)
	w.updateJobStatus(jobID, "completed", "")

	// Enqueue AI analysis
	aiPayload, _ := json.Marshal(AIPayload{DocumentID: payload.DocumentID, Text: text})
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     w.cfg.Redis.Addr(),
		Password: w.cfg.Redis.Password,
		DB:       w.cfg.Redis.DB,
	})
	defer client.Close()
	client.Enqueue(asynq.NewTask(TaskAIAnalyze, aiPayload), asynq.Queue("default"))

	return nil
}

func (w *WorkerServer) HandleAIAnalyze(ctx context.Context, t *asynq.Task) error {
	var payload AIPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal AI payload: %w", err)
	}

	jobID := w.createJob(TaskAIAnalyze, "processing", &payload.DocumentID, nil)
	log.Printf("AI analyzing document: %s (job: %s)", payload.DocumentID, jobID)

	result, err := w.aiService.AnalyzeDocument(ctx, payload.Text)
	if err != nil {
		log.Printf("AI analysis failed for document %s: %v", payload.DocumentID, err)
		w.updateJobStatus(jobID, "failed", err.Error())
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	resultJSON, _ := json.Marshal(result)
	w.db.Exec("UPDATE documents SET ai_extracted = $1, status = 'completed', updated_at = NOW() WHERE id = $2",
		resultJSON, payload.DocumentID)

	log.Printf("AI analysis completed for document: %s", payload.DocumentID)
	w.updateJobStatus(jobID, "completed", "")
	return nil
}

func (w *WorkerServer) HandleCompress(ctx context.Context, t *asynq.Task) error {
	var payload CompressPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal compress payload: %w", err)
	}

	log.Printf("Compressing file: %s", payload.FileID)
	// Compression is handled during upload in the document handler
	return nil
}

// EnqueueOCR creates an OCR task for a file
func EnqueueOCR(client *asynq.Client, fileID, documentID, mimeType string) error {
	payload, _ := json.Marshal(OCRPayload{
		FileID:     fileID,
		DocumentID: documentID,
		MimeType:   mimeType,
	})
	_, err := client.Enqueue(
		asynq.NewTask(TaskOCRProcess, payload),
		asynq.Queue("default"),
		asynq.MaxRetry(3),
	)
	return err
}
