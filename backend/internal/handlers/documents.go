package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/models"
	"github.com/haydary1986/archiving-qa/internal/services"
	"github.com/haydary1986/archiving-qa/internal/workers"
)

type DocumentHandler struct {
	db             *sql.DB
	cfg            *config.Config
	driveService   *services.DriveService
	compressService *services.CompressService
}

func NewDocumentHandler(db *sql.DB, cfg *config.Config, ds *services.DriveService, cs *services.CompressService) *DocumentHandler {
	return &DocumentHandler{db: db, cfg: cfg, driveService: ds, compressService: cs}
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var req models.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)

	customFields := []byte("{}")
	if req.CustomFields != nil {
		customFields = req.CustomFields
	}

	var doc models.Document
	err := h.db.QueryRow(`
		INSERT INTO documents (title, description, document_number, document_date, document_type,
		    category_id, classification, source_entity, dest_entity, custom_fields, created_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, title, description, document_number, document_date, document_type,
		    category_id, classification, source_entity, dest_entity, custom_fields,
		    status, created_by_id, created_at, updated_at
	`, req.Title, req.Description, req.DocumentNumber, req.DocumentDate,
		req.DocumentType, req.CategoryID, req.Classification,
		req.SourceEntity, req.DestEntity, customFields, uid,
	).Scan(
		&doc.ID, &doc.Title, &doc.Description, &doc.DocumentNumber,
		&doc.DocumentDate, &doc.DocumentType, &doc.CategoryID,
		&doc.Classification, &doc.SourceEntity, &doc.DestEntity,
		&doc.CustomFields, &doc.Status, &doc.CreatedByID,
		&doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الوثيقة", "details": err.Error()})
		return
	}

	// Link persons
	if len(req.PersonIDs) > 0 {
		for _, p := range req.PersonIDs {
			h.db.Exec("INSERT INTO document_persons (document_id, person_id, relation) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
				doc.ID, p.PersonID, p.Relation)
		}
	}

	// Link tags
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			h.db.Exec("INSERT INTO document_tags (document_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				doc.ID, tagID)
		}
	}

	// Log audit
	h.logAudit(c, uid, "create", "document", &doc.ID, nil)

	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var doc models.Document
	err := h.db.QueryRow(`
		SELECT d.id, d.title, d.description, d.document_number, d.document_date,
		       d.document_type, d.category_id, d.classification, d.source_entity,
		       d.dest_entity, d.custom_fields, d.ocr_text, d.ai_extracted,
		       d.status, d.created_by_id, d.created_at, d.updated_at
		FROM documents d
		WHERE d.id = $1 AND d.deleted_at IS NULL
	`, id).Scan(
		&doc.ID, &doc.Title, &doc.Description, &doc.DocumentNumber,
		&doc.DocumentDate, &doc.DocumentType, &doc.CategoryID,
		&doc.Classification, &doc.SourceEntity, &doc.DestEntity,
		&doc.CustomFields, &doc.OCRText, &doc.AIExtracted,
		&doc.Status, &doc.CreatedByID, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الوثيقة"})
		return
	}

	// Load files
	rows, _ := h.db.Query(`
		SELECT id, document_id, file_name, original_name, mime_type, file_size,
		       drive_file_id, drive_url, thumbnail_url, ocr_status, created_at
		FROM files WHERE document_id = $1 AND deleted_at IS NULL
	`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var f models.File
			rows.Scan(&f.ID, &f.DocumentID, &f.FileName, &f.OriginalName,
				&f.MimeType, &f.FileSize, &f.DriveFileID, &f.DriveURL,
				&f.ThumbnailURL, &f.OCRStatus, &f.CreatedAt)
			doc.Files = append(doc.Files, f)
		}
	}

	// Load persons
	pRows, _ := h.db.Query(`
		SELECT p.id, p.full_name, p.title, p.department, p.person_type, dp.relation
		FROM persons p JOIN document_persons dp ON p.id = dp.person_id
		WHERE dp.document_id = $1 AND p.deleted_at IS NULL
	`, id)
	if pRows != nil {
		defer pRows.Close()
		for pRows.Next() {
			var p models.Person
			var relation string
			pRows.Scan(&p.ID, &p.FullName, &p.Title, &p.Department, &p.PersonType, &relation)
			doc.Persons = append(doc.Persons, p)
		}
	}

	// Load routings
	rRows, _ := h.db.Query(`
		SELECT id, document_id, from_entity, to_entity, action, notes, action_date, action_by_id, created_at
		FROM routings WHERE document_id = $1 AND deleted_at IS NULL ORDER BY action_date ASC
	`, id)
	if rRows != nil {
		defer rRows.Close()
		for rRows.Next() {
			var r models.Routing
			rRows.Scan(&r.ID, &r.DocumentID, &r.FromEntity, &r.ToEntity,
				&r.Action, &r.Notes, &r.ActionDate, &r.ActionByID, &r.CreatedAt)
			doc.Routings = append(doc.Routings, r)
		}
	}

	// Load tags
	tRows, _ := h.db.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t JOIN document_tags dt ON t.id = dt.tag_id
		WHERE dt.document_id = $1 AND t.deleted_at IS NULL
	`, id)
	if tRows != nil {
		defer tRows.Close()
		for tRows.Next() {
			var t models.Tag
			tRows.Scan(&t.ID, &t.Name, &t.Color)
			doc.Tags = append(doc.Tags, t)
		}
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	var filter models.DocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات غير صالحة"})
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	query := `SELECT d.id, d.title, d.description, d.document_number, d.document_date,
		d.document_type, d.category_id, d.classification, d.source_entity,
		d.dest_entity, d.status, d.created_by_id, d.created_at, d.updated_at
		FROM documents d WHERE d.deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM documents d WHERE d.deleted_at IS NULL`

	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		searchClause := fmt.Sprintf(` AND (d.title ILIKE $%d OR d.document_number ILIKE $%d OR d.ocr_text ILIKE $%d OR d.description ILIKE $%d)`, argIdx, argIdx, argIdx, argIdx)
		query += searchClause
		countQuery += searchClause
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.DocumentType != "" {
		clause := fmt.Sprintf(` AND d.document_type = $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, filter.DocumentType)
		argIdx++
	}

	if filter.CategoryID != nil {
		clause := fmt.Sprintf(` AND d.category_id = $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.Classification != "" {
		clause := fmt.Sprintf(` AND d.classification = $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, filter.Classification)
		argIdx++
	}

	if filter.Status != "" {
		clause := fmt.Sprintf(` AND d.status = $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.DateFrom != nil {
		clause := fmt.Sprintf(` AND d.document_date >= $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.DateFrom)
		argIdx++
	}

	if filter.DateTo != nil {
		clause := fmt.Sprintf(` AND d.document_date <= $%d`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.DateTo)
		argIdx++
	}

	if filter.PersonID != nil {
		clause := fmt.Sprintf(` AND d.id IN (SELECT document_id FROM document_persons WHERE person_id = $%d)`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.PersonID)
		argIdx++
	}

	if filter.TagID != nil {
		clause := fmt.Sprintf(` AND d.id IN (SELECT document_id FROM document_tags WHERE tag_id = $%d)`, argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.TagID)
		argIdx++
	}

	// Count total
	var total int64
	h.db.QueryRow(countQuery, args...).Scan(&total)

	// Sort
	allowedSorts := map[string]bool{"created_at": true, "updated_at": true, "title": true, "document_date": true, "document_number": true}
	sortBy := "created_at"
	if allowedSorts[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if strings.ToUpper(filter.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY d.%s %s", sortBy, sortOrder)

	// Pagination
	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الوثائق"})
		return
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		rows.Scan(
			&doc.ID, &doc.Title, &doc.Description, &doc.DocumentNumber,
			&doc.DocumentDate, &doc.DocumentType, &doc.CategoryID,
			&doc.Classification, &doc.SourceEntity, &doc.DestEntity,
			&doc.Status, &doc.CreatedByID, &doc.CreatedAt, &doc.UpdatedAt,
		)
		documents = append(documents, doc)
	}

	if documents == nil {
		documents = []models.Document{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       documents,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.PageSize))),
	})
}

func (h *DocumentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	docID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req models.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	// Build dynamic update query
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.DocumentNumber != nil {
		sets = append(sets, fmt.Sprintf("document_number = $%d", argIdx))
		args = append(args, *req.DocumentNumber)
		argIdx++
	}
	if req.DocumentDate != nil {
		sets = append(sets, fmt.Sprintf("document_date = $%d", argIdx))
		args = append(args, *req.DocumentDate)
		argIdx++
	}
	if req.DocumentType != nil {
		sets = append(sets, fmt.Sprintf("document_type = $%d", argIdx))
		args = append(args, *req.DocumentType)
		argIdx++
	}
	if req.CategoryID != nil {
		sets = append(sets, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *req.CategoryID)
		argIdx++
	}
	if req.Classification != nil {
		sets = append(sets, fmt.Sprintf("classification = $%d", argIdx))
		args = append(args, *req.Classification)
		argIdx++
	}
	if req.SourceEntity != nil {
		sets = append(sets, fmt.Sprintf("source_entity = $%d", argIdx))
		args = append(args, *req.SourceEntity)
		argIdx++
	}
	if req.DestEntity != nil {
		sets = append(sets, fmt.Sprintf("dest_entity = $%d", argIdx))
		args = append(args, *req.DestEntity)
		argIdx++
	}
	if req.CustomFields != nil {
		sets = append(sets, fmt.Sprintf("custom_fields = $%d", argIdx))
		args = append(args, req.CustomFields)
		argIdx++
	}

	args = append(args, docID)
	query := fmt.Sprintf("UPDATE documents SET %s WHERE id = $%d AND deleted_at IS NULL", strings.Join(sets, ", "), argIdx)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في تحديث الوثيقة"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}

	// Update persons
	if req.PersonIDs != nil {
		h.db.Exec("DELETE FROM document_persons WHERE document_id = $1", docID)
		for _, p := range req.PersonIDs {
			h.db.Exec("INSERT INTO document_persons (document_id, person_id, relation) VALUES ($1, $2, $3)",
				docID, p.PersonID, p.Relation)
		}
	}

	// Update tags
	if req.TagIDs != nil {
		h.db.Exec("DELETE FROM document_tags WHERE document_id = $1", docID)
		for _, tagID := range req.TagIDs {
			h.db.Exec("INSERT INTO document_tags (document_id, tag_id) VALUES ($1, $2)", docID, tagID)
		}
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	h.logAudit(c, userID, "update", "document", &docID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الوثيقة بنجاح"})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	docID, _ := uuid.Parse(id)

	// Soft delete
	result, err := h.db.Exec("UPDATE documents SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في حذف الوثيقة"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	h.logAudit(c, userID, "delete", "document", &docID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الوثيقة بنجاح"})
}

func (h *DocumentHandler) Restore(c *gin.Context) {
	id := c.Param("id")
	docID, _ := uuid.Parse(id)

	result, err := h.db.Exec("UPDATE documents SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في استعادة الوثيقة"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة في سلة المهملات"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	h.logAudit(c, userID, "restore", "document", &docID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "تم استعادة الوثيقة بنجاح"})
}

func (h *DocumentHandler) ListTrash(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, title, document_number, document_type, classification, status, deleted_at, created_at
		FROM documents WHERE deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب المحذوفات"})
		return
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var doc models.Document
		rows.Scan(&doc.ID, &doc.Title, &doc.DocumentNumber, &doc.DocumentType,
			&doc.Classification, &doc.Status, &doc.DeletedAt, &doc.CreatedAt)
		docs = append(docs, doc)
	}
	if docs == nil {
		docs = []models.Document{}
	}

	c.JSON(http.StatusOK, docs)
}

func (h *DocumentHandler) UploadFile(c *gin.Context) {
	docID := c.Param("id")

	// Verify document exists
	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1 AND deleted_at IS NULL)", docID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الملف مطلوب"})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.cfg.Storage.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "حجم الملف يتجاوز الحد المسموح"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := false
	for _, allowed := range h.cfg.Storage.AllowedTypes {
		if ext == allowed {
			validExt = true
			break
		}
	}
	if !validExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "نوع الملف غير مسموح"})
		return
	}

	// Save temp file for compression
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, uuid.New().String()+ext)
	out, err := os.Create(tmpFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في معالجة الملف"})
		return
	}
	io.Copy(out, file)
	out.Close()
	defer os.Remove(tmpFile)

	// Compress if applicable
	processedFile := tmpFile
	if h.cfg.Storage.CompressImages && isImageFile(ext) {
		compressed := tmpFile + ".compressed" + ext
		if err := h.compressService.CompressImage(tmpFile, compressed); err == nil {
			processedFile = compressed
			defer os.Remove(compressed)
		}
	} else if h.cfg.Storage.CompressPDFs && ext == ".pdf" {
		compressed := tmpFile + ".compressed.pdf"
		if err := h.compressService.CompressPDF(tmpFile, compressed); err == nil {
			processedFile = compressed
			defer os.Remove(compressed)
		}
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	fileName := fmt.Sprintf("%s_%s", docID, header.Filename)

	// Upload to Google Drive
	var driveFileID, driveURL string
	if h.driveService != nil {
		f, err := os.Open(processedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في قراءة الملف"})
			return
		}
		defer f.Close()

		driveFileID, driveURL, err = h.driveService.UploadFile(c, fileName, f, mimeType, h.cfg.Google.DriveFolderID)
		if err != nil {
			log.Printf("Drive upload failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في رفع الملف إلى Drive", "details": err.Error()})
			return
		}
		log.Printf("File uploaded to Drive: %s (URL: %s)", driveFileID, driveURL)
	} else {
		log.Println("WARNING: Drive service not configured, skipping Drive upload")
	}

	// Get file size
	fi, _ := os.Stat(processedFile)
	fileSize := fi.Size()

	userID := c.GetString("user_id")

	var fileRecord models.File
	err = h.db.QueryRow(`
		INSERT INTO files (document_id, file_name, original_name, mime_type, file_size,
		    drive_file_id, drive_url, ocr_status, uploaded_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8)
		RETURNING id, document_id, file_name, original_name, mime_type, file_size,
		    drive_file_id, drive_url, ocr_status, uploaded_by_id, created_at
	`, docID, fileName, header.Filename, mimeType, fileSize, driveFileID, driveURL, userID,
	).Scan(
		&fileRecord.ID, &fileRecord.DocumentID, &fileRecord.FileName,
		&fileRecord.OriginalName, &fileRecord.MimeType, &fileRecord.FileSize,
		&fileRecord.DriveFileID, &fileRecord.DriveURL, &fileRecord.OCRStatus,
		&fileRecord.UploadedByID, &fileRecord.CreatedAt,
	)
	if err != nil {
		log.Printf("Failed to save file record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في حفظ معلومات الملف"})
		return
	}

	// Enqueue OCR task
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     h.cfg.Redis.Addr(),
		Password: h.cfg.Redis.Password,
		DB:       h.cfg.Redis.DB,
	})
	defer asynqClient.Close()

	if err := workers.EnqueueOCR(asynqClient, fileRecord.ID.String(), docID, mimeType); err != nil {
		log.Printf("Failed to enqueue OCR task: %v", err)
	} else {
		log.Printf("OCR task enqueued for file: %s", fileRecord.ID)
	}

	uid, _ := uuid.Parse(userID)
	h.logAudit(c, uid, "upload", "file", &fileRecord.ID, nil)

	c.JSON(http.StatusCreated, fileRecord)
}

func (h *DocumentHandler) AddRouting(c *gin.Context) {
	var req models.CreateRoutingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	var routing models.Routing
	err := h.db.QueryRow(`
		INSERT INTO routings (document_id, from_entity, to_entity, action, notes, action_date, action_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, document_id, from_entity, to_entity, action, notes, action_date, action_by_id, created_at
	`, req.DocumentID, req.FromEntity, req.ToEntity, req.Action, req.Notes, req.ActionDate, userID,
	).Scan(
		&routing.ID, &routing.DocumentID, &routing.FromEntity, &routing.ToEntity,
		&routing.Action, &routing.Notes, &routing.ActionDate, &routing.ActionByID, &routing.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إضافة مسار التوجيه"})
		return
	}

	c.JSON(http.StatusCreated, routing)
}

func (h *DocumentHandler) logAudit(c *gin.Context, userID uuid.UUID, action, resource string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	if details == nil {
		detailsJSON = []byte("{}")
	}
	h.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, action, resource, resourceID, detailsJSON, c.ClientIP(), c.Request.UserAgent())
}

func (h *DocumentHandler) ListEntities(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT DISTINCT entity FROM (
			SELECT source_entity AS entity FROM documents WHERE source_entity != '' AND deleted_at IS NULL
			UNION
			SELECT dest_entity AS entity FROM documents WHERE dest_entity != '' AND deleted_at IS NULL
		) sub ORDER BY entity
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الجهات"})
		return
	}
	defer rows.Close()

	var entities []string
	for rows.Next() {
		var e string
		rows.Scan(&e)
		entities = append(entities, e)
	}
	if entities == nil {
		entities = []string{}
	}

	c.JSON(http.StatusOK, entities)
}

func isImageFile(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".tiff", ".bmp", ".webp":
		return true
	}
	return false
}
