package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/haydary1986/archiving-qa/internal/models"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// ==================== Audit Logs ====================

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
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

	userFilter := c.Query("user_id")
	actionFilter := c.Query("action")
	resourceFilter := c.Query("resource")

	query := `SELECT a.id, a.user_id, u.full_name, u.email, a.action, a.resource,
		a.resource_id, a.details, a.ip_address, a.user_agent, a.created_at
		FROM audit_logs a JOIN users u ON a.user_id = u.id WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_logs a WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if userFilter != "" {
		clause := fmt.Sprintf(" AND a.user_id = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, userFilter)
		argIdx++
	}
	if actionFilter != "" {
		clause := fmt.Sprintf(" AND a.action = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, actionFilter)
		argIdx++
	}
	if resourceFilter != "" {
		clause := fmt.Sprintf(" AND a.resource = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, resourceFilter)
		argIdx++
	}

	var total int64
	h.db.QueryRow(countQuery, args...).Scan(&total)

	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب سجل التدقيق"})
		return
	}
	defer rows.Close()

	type AuditLogEntry struct {
		models.AuditLog
		UserName  string `json:"user_name"`
		UserEmail string `json:"user_email"`
	}

	var logs []AuditLogEntry
	for rows.Next() {
		var log AuditLogEntry
		rows.Scan(
			&log.ID, &log.UserID, &log.UserName, &log.UserEmail,
			&log.Action, &log.Resource, &log.ResourceID,
			&log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		logs = append(logs, log)
	}
	if logs == nil {
		logs = []AuditLogEntry{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ==================== System Settings ====================

func (h *AdminHandler) GetSettings(c *gin.Context) {
	rows, err := h.db.Query("SELECT key, value, updated_at FROM system_settings ORDER BY key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الإعدادات"})
		return
	}
	defer rows.Close()

	settings := map[string]string{}
	for rows.Next() {
		var s models.SystemSetting
		rows.Scan(&s.Key, &s.Value, &s.UpdatedAt)
		settings[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSetting(c *gin.Context) {
	var req models.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	h.db.Exec(`
		INSERT INTO system_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, req.Key, req.Value)

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الإعدادات"})
}

// ==================== Custom Field Definitions ====================

func (h *AdminHandler) ListCustomFields(c *gin.Context) {
	categoryID := c.Query("category_id")

	query := `SELECT id, name, label, field_type, options, required, category_id, sort_order, created_at
		FROM custom_field_defs WHERE deleted_at IS NULL`
	args := []interface{}{}

	if categoryID != "" {
		query += " AND (category_id = $1 OR category_id IS NULL)"
		args = append(args, categoryID)
	}
	query += " ORDER BY sort_order ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الحقول المخصصة"})
		return
	}
	defer rows.Close()

	var fields []models.CustomFieldDef
	for rows.Next() {
		var f models.CustomFieldDef
		rows.Scan(&f.ID, &f.Name, &f.Label, &f.FieldType, &f.Options,
			&f.Required, &f.CategoryID, &f.SortOrder, &f.CreatedAt)
		fields = append(fields, f)
	}
	if fields == nil {
		fields = []models.CustomFieldDef{}
	}

	c.JSON(http.StatusOK, fields)
}

func (h *AdminHandler) CreateCustomField(c *gin.Context) {
	var req models.CreateCustomFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	options := []byte("[]")
	if req.Options != nil {
		options = req.Options
	}

	var field models.CustomFieldDef
	err := h.db.QueryRow(`
		INSERT INTO custom_field_defs (name, label, field_type, options, required, category_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, label, field_type, options, required, category_id, sort_order, created_at
	`, req.Name, req.Label, req.FieldType, options, req.Required, req.CategoryID, req.SortOrder).Scan(
		&field.ID, &field.Name, &field.Label, &field.FieldType, &field.Options,
		&field.Required, &field.CategoryID, &field.SortOrder, &field.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الحقل المخصص"})
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *AdminHandler) DeleteCustomField(c *gin.Context) {
	id := c.Param("id")
	h.db.Exec("UPDATE custom_field_defs SET deleted_at = NOW() WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الحقل المخصص"})
}

// ==================== Tags ====================

type TagHandler struct {
	db *sql.DB
}

func NewTagHandler(db *sql.DB) *TagHandler {
	return &TagHandler{db: db}
}

func (h *TagHandler) List(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, name, color, created_at FROM tags WHERE deleted_at IS NULL ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الوسوم"})
		return
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		rows.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt)
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []models.Tag{}
	}

	c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}
	if req.Color == "" {
		req.Color = "#3B82F6"
	}

	var tag models.Tag
	err := h.db.QueryRow("INSERT INTO tags (name, color) VALUES ($1, $2) RETURNING id, name, color, created_at",
		req.Name, req.Color).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الوسم"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.db.Exec("UPDATE tags SET deleted_at = NOW() WHERE id = $1", id)
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الوسم"})
}

// ==================== Share Links ====================

type ShareHandler struct {
	db *sql.DB
}

func NewShareHandler(db *sql.DB) *ShareHandler {
	return &ShareHandler{db: db}
}

func (h *ShareHandler) Create(c *gin.Context) {
	var req models.CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	userID := c.GetString("user_id")
	token := generateJTI() + generateJTI()

	var passwordHash string
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		passwordHash = string(hash)
	}

	maxViews := req.MaxViews
	if maxViews <= 0 {
		maxViews = 0 // unlimited
	}

	var link models.ShareLink
	err := h.db.QueryRow(`
		INSERT INTO share_links (document_id, token, password_hash, expires_at, max_views, created_by_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, document_id, token, expires_at, max_views, view_count, is_active, created_at
	`, req.DocumentID, token, passwordHash, req.ExpiresAt, maxViews, userID).Scan(
		&link.ID, &link.DocumentID, &link.Token, &link.ExpiresAt,
		&link.MaxViews, &link.ViewCount, &link.IsActive, &link.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء رابط المشاركة"})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *ShareHandler) Access(c *gin.Context) {
	token := c.Param("token")
	password := c.Query("password")

	var link models.ShareLink
	err := h.db.QueryRow(`
		SELECT id, document_id, token, password_hash, expires_at, max_views, view_count, is_active
		FROM share_links WHERE token = $1 AND deleted_at IS NULL
	`, token).Scan(
		&link.ID, &link.DocumentID, &link.Token, &link.PasswordHash,
		&link.ExpiresAt, &link.MaxViews, &link.ViewCount, &link.IsActive,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "رابط غير صالح"})
		return
	}

	if !link.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "الرابط معطل"})
		return
	}

	if link.ExpiresAt != nil && link.ExpiresAt.Before(timeNow()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "الرابط منتهي الصلاحية"})
		return
	}

	if link.MaxViews > 0 && link.ViewCount >= link.MaxViews {
		c.JSON(http.StatusForbidden, gin.H{"error": "تم تجاوز عدد المشاهدات المسموح"})
		return
	}

	if link.PasswordHash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(link.PasswordHash), []byte(password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "كلمة المرور غير صحيحة", "requires_password": true})
			return
		}
	}

	// Increment view count
	h.db.Exec("UPDATE share_links SET view_count = view_count + 1 WHERE id = $1", link.ID)

	// Return document
	var doc models.Document
	h.db.QueryRow(`
		SELECT id, title, description, document_number, document_date, document_type,
		       classification, source_entity, dest_entity, status, created_at
		FROM documents WHERE id = $1 AND deleted_at IS NULL
	`, link.DocumentID).Scan(
		&doc.ID, &doc.Title, &doc.Description, &doc.DocumentNumber,
		&doc.DocumentDate, &doc.DocumentType, &doc.Classification,
		&doc.SourceEntity, &doc.DestEntity, &doc.Status, &doc.CreatedAt,
	)

	c.JSON(http.StatusOK, doc)
}

// ==================== Dashboard Stats ====================

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats := map[string]interface{}{}

	var totalDocs, totalFiles, totalUsers, totalPersons int64
	h.db.QueryRow("SELECT COUNT(*) FROM documents WHERE deleted_at IS NULL").Scan(&totalDocs)
	h.db.QueryRow("SELECT COUNT(*) FROM files WHERE deleted_at IS NULL").Scan(&totalFiles)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&totalUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM persons WHERE deleted_at IS NULL").Scan(&totalPersons)

	stats["total_documents"] = totalDocs
	stats["total_files"] = totalFiles
	stats["total_users"] = totalUsers
	stats["total_persons"] = totalPersons

	// Documents by type
	typeRows, _ := h.db.Query("SELECT document_type, COUNT(*) FROM documents WHERE deleted_at IS NULL GROUP BY document_type")
	if typeRows != nil {
		defer typeRows.Close()
		byType := map[string]int64{}
		for typeRows.Next() {
			var t string
			var count int64
			typeRows.Scan(&t, &count)
			byType[t] = count
		}
		stats["documents_by_type"] = byType
	}

	// Documents by classification
	classRows, _ := h.db.Query("SELECT classification, COUNT(*) FROM documents WHERE deleted_at IS NULL GROUP BY classification")
	if classRows != nil {
		defer classRows.Close()
		byClass := map[string]int64{}
		for classRows.Next() {
			var cl string
			var count int64
			classRows.Scan(&cl, &count)
			byClass[cl] = count
		}
		stats["documents_by_classification"] = byClass
	}

	// Recent activity
	recentRows, _ := h.db.Query(`
		SELECT a.action, a.resource, u.full_name, a.created_at
		FROM audit_logs a JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC LIMIT 10
	`)
	if recentRows != nil {
		defer recentRows.Close()
		var recent []map[string]interface{}
		for recentRows.Next() {
			entry := map[string]interface{}{}
			var action, resource, userName string
			var createdAt interface{}
			recentRows.Scan(&action, &resource, &userName, &createdAt)
			entry["action"] = action
			entry["resource"] = resource
			entry["user_name"] = userName
			entry["created_at"] = createdAt
			recent = append(recent, entry)
		}
		stats["recent_activity"] = recent
	}

	// Storage stats
	var totalSize int64
	h.db.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM files WHERE deleted_at IS NULL").Scan(&totalSize)
	stats["total_storage_bytes"] = totalSize
	stats["total_storage_mb"] = totalSize / (1024 * 1024)

	c.JSON(http.StatusOK, stats)
}

// ==================== Export Handler ====================

type ExportHandler struct {
	db            *sql.DB
	exportService interface{ ExportToExcel(ctx interface{}, docs []models.Document) ([]byte, error) }
}

func NewExportHandler(db *sql.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

func (h *ExportHandler) Export(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	// Fetch documents
	var docs []models.Document
	for _, docID := range req.DocumentIDs {
		var doc models.Document
		err := h.db.QueryRow(`
			SELECT id, title, description, document_number, document_date, document_type,
			       classification, source_entity, dest_entity, custom_fields, status, created_at
			FROM documents WHERE id = $1 AND deleted_at IS NULL
		`, docID).Scan(
			&doc.ID, &doc.Title, &doc.Description, &doc.DocumentNumber,
			&doc.DocumentDate, &doc.DocumentType, &doc.Classification,
			&doc.SourceEntity, &doc.DestEntity, &doc.CustomFields,
			&doc.Status, &doc.CreatedAt,
		)
		if err == nil {
			docs = append(docs, doc)
		}
	}

	if len(docs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "لم يتم العثور على وثائق"})
		return
	}

	// For now, return as JSON (Excel/ZIP export via service layer)
	c.JSON(http.StatusOK, gin.H{
		"message":   "تم تجهيز البيانات للتصدير",
		"count":     len(docs),
		"documents": docs,
	})
}

func timeNow() time.Time {
	return time.Now()
}
