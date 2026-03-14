package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/haydary1986/archiving-qa/internal/models"
)

type PersonHandler struct {
	db *sql.DB
}

func NewPersonHandler(db *sql.DB) *PersonHandler {
	return &PersonHandler{db: db}
}

func (h *PersonHandler) Create(c *gin.Context) {
	var req models.CreatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var person models.Person
	err := h.db.QueryRow(`
		INSERT INTO persons (full_name, title, department, email, phone, person_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, full_name, title, department, email, phone, person_type, created_at, updated_at
	`, req.FullName, req.Title, req.Department, req.Email, req.Phone, req.PersonType).Scan(
		&person.ID, &person.FullName, &person.Title, &person.Department,
		&person.Email, &person.Phone, &person.PersonType, &person.CreatedAt, &person.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء سجل الشخص"})
		return
	}

	c.JSON(http.StatusCreated, person)
}

func (h *PersonHandler) List(c *gin.Context) {
	search := c.Query("search")
	personType := c.Query("type")

	query := `SELECT id, full_name, title, department, email, phone, person_type, created_at
		FROM persons WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		query += fmt.Sprintf(` AND (full_name ILIKE $%d OR email ILIKE $%d OR department ILIKE $%d)`, argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	if personType != "" {
		query += fmt.Sprintf(` AND person_type = $%d`, argIdx)
		args = append(args, personType)
		argIdx++
	}

	query += " ORDER BY full_name ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الأشخاص"})
		return
	}
	defer rows.Close()

	var persons []models.Person
	for rows.Next() {
		var p models.Person
		rows.Scan(&p.ID, &p.FullName, &p.Title, &p.Department,
			&p.Email, &p.Phone, &p.PersonType, &p.CreatedAt)
		persons = append(persons, p)
	}
	if persons == nil {
		persons = []models.Person{}
	}

	c.JSON(http.StatusOK, persons)
}

func (h *PersonHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var person models.Person
	err := h.db.QueryRow(`
		SELECT id, full_name, title, department, email, phone, person_type, created_at, updated_at
		FROM persons WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&person.ID, &person.FullName, &person.Title, &person.Department,
		&person.Email, &person.Phone, &person.PersonType, &person.CreatedAt, &person.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "الشخص غير موجود"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في الخادم"})
		return
	}

	// Get related documents
	rows, _ := h.db.Query(`
		SELECT d.id, d.title, d.document_number, d.document_type, d.document_date, dp.relation
		FROM documents d JOIN document_persons dp ON d.id = dp.document_id
		WHERE dp.person_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.document_date DESC
	`, id)
	if rows != nil {
		defer rows.Close()
		type PersonDoc struct {
			models.Document
			Relation string `json:"relation"`
		}
		var docs []PersonDoc
		for rows.Next() {
			var pd PersonDoc
			rows.Scan(&pd.ID, &pd.Title, &pd.DocumentNumber, &pd.DocumentType, &pd.DocumentDate, &pd.Relation)
			docs = append(docs, pd)
		}
		c.JSON(http.StatusOK, gin.H{"person": person, "documents": docs})
		return
	}

	c.JSON(http.StatusOK, gin.H{"person": person, "documents": []interface{}{}})
}

func (h *PersonHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.CreatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	result, err := h.db.Exec(`
		UPDATE persons SET full_name=$1, title=$2, department=$3, email=$4, phone=$5,
		    person_type=$6, updated_at=NOW()
		WHERE id = $7 AND deleted_at IS NULL
	`, req.FullName, req.Title, req.Department, req.Email, req.Phone, req.PersonType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في تحديث البيانات"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "الشخص غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث البيانات بنجاح"})
}

func (h *PersonHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.db.Exec("UPDATE persons SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	c.JSON(http.StatusOK, gin.H{"message": "تم الحذف بنجاح"})
}
