package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/models"
)

type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var user models.User
	var roleID uuid.UUID
	var roleName string
	err := h.db.QueryRow(`
		SELECT u.id, u.email, u.password_hash, u.full_name, u.avatar_url, u.provider,
		       u.is_active, u.role_id, r.name as role_name
		FROM users u JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 AND u.deleted_at IS NULL
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.AvatarURL, &user.Provider, &user.IsActive, &roleID, &roleName,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "البريد الإلكتروني أو كلمة المرور غير صحيحة"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في الخادم"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "الحساب معطل"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "البريد الإلكتروني أو كلمة المرور غير صحيحة"})
		return
	}

	user.RoleID = roleID
	user.Role = &models.Role{Name: roleName}

	accessToken, err := h.generateToken(user.ID, user.Email, roleID, roleName, h.cfg.JWT.Secret, time.Duration(h.cfg.JWT.ExpirationHours)*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الرمز"})
		return
	}

	refreshToken, err := h.generateToken(user.ID, user.Email, roleID, roleName, h.cfg.JWT.RefreshSecret, 7*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الرمز"})
		return
	}

	// Update last login
	h.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.cfg.JWT.ExpirationHours * 3600,
		User:         user,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	// Check if local auth is enabled
	var localAuthEnabled string
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'local_auth_enabled'").Scan(&localAuthEnabled)
	if localAuthEnabled == "false" {
		c.JSON(http.StatusForbidden, gin.H{"error": "التسجيل المحلي معطل"})
		return
	}

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	// Check if email exists
	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)", req.Email).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "البريد الإلكتروني مسجل مسبقاً"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في الخادم"})
		return
	}

	// Default role: viewer
	viewerRoleID := "a0000000-0000-0000-0000-000000000004"

	var user models.User
	err = h.db.QueryRow(`
		INSERT INTO users (email, password_hash, full_name, provider, role_id)
		VALUES ($1, $2, $3, 'local', $4)
		RETURNING id, email, full_name, provider, is_active, role_id, created_at, updated_at
	`, req.Email, string(hashedPassword), req.FullName, viewerRoleID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Provider,
		&user.IsActive, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطأ في إنشاء الحساب"})
		return
	}

	accessToken, _ := h.generateToken(user.ID, user.Email, user.RoleID, "viewer", h.cfg.JWT.Secret, time.Duration(h.cfg.JWT.ExpirationHours)*time.Hour)
	refreshToken, _ := h.generateToken(user.ID, user.Email, user.RoleID, "viewer", h.cfg.JWT.RefreshSecret, 7*24*time.Hour)

	c.JSON(http.StatusCreated, models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.cfg.JWT.ExpirationHours * 3600,
		User:         user,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWT.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "رمز التحديث غير صالح"})
		return
	}

	userID, _ := uuid.Parse(claims["user_id"].(string))
	email := claims["email"].(string)
	roleID, _ := uuid.Parse(claims["role_id"].(string))
	roleName := claims["role_name"].(string)

	accessToken, _ := h.generateToken(userID, email, roleID, roleName, h.cfg.JWT.Secret, time.Duration(h.cfg.JWT.ExpirationHours)*time.Hour)
	newRefreshToken, _ := h.generateToken(userID, email, roleID, roleName, h.cfg.JWT.RefreshSecret, 7*24*time.Hour)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    h.cfg.JWT.ExpirationHours * 3600,
	})
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز المصادقة مطلوب"})
		return
	}

	// Exchange code for token - implementation depends on oauth2 setup
	// This is a placeholder for the Google OAuth flow
	c.JSON(http.StatusOK, gin.H{"message": "Google OAuth callback - implement with your Google credentials"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	var roleName string
	err := h.db.QueryRow(`
		SELECT u.id, u.email, u.full_name, u.avatar_url, u.provider, u.is_active,
		       u.role_id, r.name as role_name, u.last_login_at, u.created_at, u.updated_at
		FROM users u JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, userID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.AvatarURL,
		&user.Provider, &user.IsActive, &user.RoleID, &roleName,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	user.Role = &models.Role{Name: roleName}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) generateToken(userID uuid.UUID, email string, roleID uuid.UUID, roleName string, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID.String(),
		"email":     email,
		"role_id":   roleID.String(),
		"role_name": roleName,
		"exp":       time.Now().Add(expiry).Unix(),
		"iat":       time.Now().Unix(),
		"jti":       generateJTI(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
