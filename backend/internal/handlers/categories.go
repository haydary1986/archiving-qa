package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/haydary1986/archiving-qa/internal/models"
)

type CategoryHandler struct {
	db *sql.DB
}

func NewCategoryHandler(db *sql.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var cat models.Category
	err := h.db.QueryRow(`
		INSERT INTO categories (name, parent_id, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, name, parent_id, sort_order, created_at, updated_at
	`, req.Name, req.ParentID, req.SortOrder).Scan(
		&cat.ID, &cat.Name, &cat.ParentID, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء التصنيف"})
		return
	}

	c.JSON(http.StatusCreated, cat)
}

func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE deleted_at IS NULL
		ORDER BY sort_order ASC, name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب التصنيفات"})
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		rows.Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt)
		categories = append(categories, cat)
	}
	if categories == nil {
		categories = []models.Category{}
	}

	// Build tree structure
	tree := buildCategoryTree(categories, nil)
	c.JSON(http.StatusOK, tree)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	result, err := h.db.Exec(`
		UPDATE categories SET name = $1, parent_id = $2, sort_order = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`, req.Name, req.ParentID, req.SortOrder, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في تحديث التصنيف"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "التصنيف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث التصنيف بنجاح"})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.db.Exec("UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف التصنيف بنجاح"})
}

func buildCategoryTree(categories []models.Category, parentID *uuid.UUID) []models.Category {
	var tree []models.Category
	for _, cat := range categories {
		if (parentID == nil && cat.ParentID == nil) || (parentID != nil && cat.ParentID != nil && *parentID == *cat.ParentID) {
			cat.Children = buildCategoryTree(categories, &cat.ID)
			tree = append(tree, cat)
		}
	}
	if tree == nil {
		tree = []models.Category{}
	}
	return tree
}
