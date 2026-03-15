package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/haydary1986/archiving-qa/internal/models"
)

type QueueHandler struct {
	db *sql.DB
}

func NewQueueHandler(db *sql.DB) *QueueHandler {
	return &QueueHandler{db: db}
}

// ListJobs returns all background jobs with filtering
func (h *QueueHandler) ListJobs(c *gin.Context) {
	page := 1
	pageSize := 50
	fmt.Sscanf(c.DefaultQuery("page", "1"), "%d", &page)
	fmt.Sscanf(c.DefaultQuery("page_size", "50"), "%d", &pageSize)
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	statusFilter := c.Query("status")
	taskTypeFilter := c.Query("task_type")

	query := `SELECT j.id, j.task_type, j.status, j.document_id, j.file_id,
		j.error_message, j.attempts, j.max_retries,
		j.created_at, j.started_at, j.completed_at,
		COALESCE(d.title, '') as doc_title,
		COALESCE(f.original_name, '') as file_name
		FROM jobs j
		LEFT JOIN documents d ON j.document_id = d.id
		LEFT JOIN files f ON j.file_id = f.id
		WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM jobs j WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if statusFilter != "" {
		clause := fmt.Sprintf(" AND j.status = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, statusFilter)
		argIdx++
	}
	if taskTypeFilter != "" {
		clause := fmt.Sprintf(" AND j.task_type = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, taskTypeFilter)
		argIdx++
	}

	var total int64
	h.db.QueryRow(countQuery, args...).Scan(&total)

	query += fmt.Sprintf(" ORDER BY j.created_at DESC LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب المهام"})
		return
	}
	defer rows.Close()

	type JobEntry struct {
		ID           uuid.UUID  `json:"id"`
		TaskType     string     `json:"task_type"`
		Status       string     `json:"status"`
		DocumentID   *uuid.UUID `json:"document_id,omitempty"`
		FileID       *uuid.UUID `json:"file_id,omitempty"`
		ErrorMessage string     `json:"error_message,omitempty"`
		Attempts     int        `json:"attempts"`
		MaxRetries   int        `json:"max_retries"`
		CreatedAt    interface{} `json:"created_at"`
		StartedAt    interface{} `json:"started_at,omitempty"`
		CompletedAt  interface{} `json:"completed_at,omitempty"`
		DocTitle     string     `json:"doc_title,omitempty"`
		FileName     string     `json:"file_name,omitempty"`
	}

	var jobs []JobEntry
	for rows.Next() {
		var j JobEntry
		rows.Scan(
			&j.ID, &j.TaskType, &j.Status, &j.DocumentID, &j.FileID,
			&j.ErrorMessage, &j.Attempts, &j.MaxRetries,
			&j.CreatedAt, &j.StartedAt, &j.CompletedAt,
			&j.DocTitle, &j.FileName,
		)
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []JobEntry{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:     jobs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// QueueStats returns summary counts by status and type
func (h *QueueHandler) QueueStats(c *gin.Context) {
	stats := map[string]interface{}{}

	// By status
	statusRows, _ := h.db.Query("SELECT status, COUNT(*) FROM jobs GROUP BY status")
	if statusRows != nil {
		defer statusRows.Close()
		byStatus := map[string]int64{}
		for statusRows.Next() {
			var s string
			var count int64
			statusRows.Scan(&s, &count)
			byStatus[s] = count
		}
		stats["by_status"] = byStatus
	}

	// By type
	typeRows, _ := h.db.Query("SELECT task_type, COUNT(*) FROM jobs GROUP BY task_type")
	if typeRows != nil {
		defer typeRows.Close()
		byType := map[string]int64{}
		for typeRows.Next() {
			var t string
			var count int64
			typeRows.Scan(&t, &count)
			byType[t] = count
		}
		stats["by_type"] = byType
	}

	// Files OCR status
	ocrRows, _ := h.db.Query("SELECT ocr_status, COUNT(*) FROM files WHERE deleted_at IS NULL GROUP BY ocr_status")
	if ocrRows != nil {
		defer ocrRows.Close()
		ocrStatus := map[string]int64{}
		for ocrRows.Next() {
			var s string
			var count int64
			ocrRows.Scan(&s, &count)
			ocrStatus[s] = count
		}
		stats["ocr_status"] = ocrStatus
	}

	// Recent failed
	var failedCount int64
	h.db.QueryRow("SELECT COUNT(*) FROM jobs WHERE status = 'failed'").Scan(&failedCount)
	stats["failed_count"] = failedCount

	// Active processing
	var activeCount int64
	h.db.QueryRow("SELECT COUNT(*) FROM jobs WHERE status = 'processing'").Scan(&activeCount)
	stats["active_count"] = activeCount

	c.JSON(http.StatusOK, stats)
}

// RetryJob resets a failed job to pending
func (h *QueueHandler) RetryJob(c *gin.Context) {
	id := c.Param("id")

	result, err := h.db.Exec(`
		UPDATE jobs SET status = 'pending', error_message = '', started_at = NULL, completed_at = NULL
		WHERE id = $1 AND status = 'failed'
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إعادة المهمة"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "المهمة غير موجودة أو ليست فاشلة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم إعادة المهمة إلى الطابور"})
}

// ClearCompleted removes all completed jobs
func (h *QueueHandler) ClearCompleted(c *gin.Context) {
	result, _ := h.db.Exec("DELETE FROM jobs WHERE status = 'completed'")
	rows, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم حذف %d مهمة مكتملة", rows)})
}
