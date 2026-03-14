package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/haydary1986/archiving-qa/internal/models"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT u.id, u.email, u.full_name, u.avatar_url, u.provider, u.is_active,
		       u.role_id, r.name as role_name, u.last_login_at, u.created_at
		FROM users u JOIN roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب المستخدمين"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var roleName string
		rows.Scan(&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.Provider,
			&u.IsActive, &u.RoleID, &roleName, &u.LastLoginAt, &u.CreatedAt)
		u.Role = &models.Role{Name: roleName}
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var u models.User
	var roleName string
	err := h.db.QueryRow(`
		SELECT u.id, u.email, u.full_name, u.avatar_url, u.provider, u.is_active,
		       u.role_id, r.name as role_name, u.last_login_at, u.created_at, u.updated_at
		FROM users u JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.Provider,
		&u.IsActive, &u.RoleID, &roleName, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في الخادم"})
		return
	}

	u.Role = &models.Role{Name: roleName}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Email    string    `json:"email" binding:"required,email"`
		Password string    `json:"password" binding:"required,min=8"`
		FullName string    `json:"full_name" binding:"required"`
		RoleID   uuid.UUID `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)", req.Email).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "البريد الإلكتروني مسجل مسبقاً"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	var user models.User
	err := h.db.QueryRow(`
		INSERT INTO users (email, password_hash, full_name, provider, role_id)
		VALUES ($1, $2, $3, 'local', $4)
		RETURNING id, email, full_name, provider, is_active, role_id, created_at
	`, req.Email, string(hash), req.FullName, req.RoleID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Provider,
		&user.IsActive, &user.RoleID, &user.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء المستخدم"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if req.FullName != "" {
		h.db.Exec("UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2", req.FullName, id)
	}
	if req.IsActive != nil {
		h.db.Exec("UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2", *req.IsActive, id)
	}
	if req.RoleID != nil {
		h.db.Exec("UPDATE users SET role_id = $1, updated_at = NOW() WHERE id = $2", *req.RoleID, id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث المستخدم بنجاح"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	// Prevent self-deletion
	if id == c.GetString("user_id") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لا يمكنك حذف حسابك الخاص"})
		return
	}

	h.db.Exec("UPDATE users SET deleted_at = NOW(), is_active = false WHERE id = $1 AND deleted_at IS NULL", id)
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المستخدم بنجاح"})
}

// ==================== Roles ====================

type RoleHandler struct {
	db *sql.DB
}

func NewRoleHandler(db *sql.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

func (h *RoleHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, description, created_at FROM roles WHERE deleted_at IS NULL ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الأدوار"})
		return
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt)

		// Load permissions
		pRows, _ := h.db.Query(`
			SELECT p.id, p.name, p.resource, p.action, p.description
			FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
		`, r.ID)
		if pRows != nil {
			for pRows.Next() {
				var p models.Permission
				pRows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description)
				r.Permissions = append(r.Permissions, p)
			}
			pRows.Close()
		}
		if r.Permissions == nil {
			r.Permissions = []models.Permission{}
		}
		roles = append(roles, r)
	}
	if roles == nil {
		roles = []models.Role{}
	}

	c.JSON(http.StatusOK, roles)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	var role models.Role
	err := h.db.QueryRow(`
		INSERT INTO roles (name, description) VALUES ($1, $2)
		RETURNING id, name, description, created_at
	`, req.Name, req.Description).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الدور"})
		return
	}

	// Assign permissions
	for _, pid := range req.PermissionIDs {
		h.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", role.ID, pid)
	}

	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	h.db.Exec("UPDATE roles SET name = $1, description = $2, updated_at = NOW() WHERE id = $3", req.Name, req.Description, id)

	// Update permissions
	h.db.Exec("DELETE FROM role_permissions WHERE role_id = $1", id)
	for _, pid := range req.PermissionIDs {
		h.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)", id, pid)
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الدور بنجاح"})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, name, resource, action, description FROM permissions ORDER BY resource, action")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في جلب الصلاحيات"})
		return
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description)
		perms = append(perms, p)
	}
	if perms == nil {
		perms = []models.Permission{}
	}

	c.JSON(http.StatusOK, perms)
}
