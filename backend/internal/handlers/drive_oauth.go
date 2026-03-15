package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveOAuthHandler struct {
	db *sql.DB
}

func NewDriveOAuthHandler(db *sql.DB) *DriveOAuthHandler {
	return &DriveOAuthHandler{db: db}
}

func (h *DriveOAuthHandler) getOAuthConfig(c *gin.Context) *oauth2.Config {
	var clientID, clientSecret, redirectURL string
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_client_id'").Scan(&clientID)
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_client_secret'").Scan(&clientSecret)
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_redirect_url'").Scan(&redirectURL)

	if redirectURL == "" {
		// Auto-detect from request
		scheme := "https"
		if c != nil {
			if fwd := c.GetHeader("X-Forwarded-Proto"); fwd != "" {
				scheme = fwd
			}
		}
		host := ""
		if c != nil {
			host = c.GetHeader("X-Forwarded-Host")
			if host == "" {
				host = c.Request.Host
			}
		}
		if host != "" {
			redirectURL = scheme + "://" + host + "/api/v1/admin/drive/callback"
		}
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{drive.DriveScope},
		Endpoint:     google.Endpoint,
	}
}

// SaveCredentials saves OAuth client ID and secret
func (h *DriveOAuthHandler) SaveCredentials(c *gin.Context) {
	var req struct {
		ClientID     string `json:"client_id" binding:"required"`
		ClientSecret string `json:"client_secret" binding:"required"`
		FolderID     string `json:"folder_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	h.upsertSetting("drive_client_id", req.ClientID)
	h.upsertSetting("drive_client_secret", req.ClientSecret)
	if req.FolderID != "" {
		h.upsertSetting("drive_folder_id", req.FolderID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حفظ بيانات الاعتماد"})
}

// GetAuthURL returns the Google OAuth authorization URL
func (h *DriveOAuthHandler) GetAuthURL(c *gin.Context) {
	cfg := h.getOAuthConfig(c)
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال Client ID و Client Secret أولاً"})
		return
	}

	url := cfg.AuthCodeURL("archiving-drive-auth", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// Callback handles the OAuth2 callback from Google
func (h *DriveOAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز التفويض مفقود"})
		return
	}

	cfg := h.getOAuthConfig(c)
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Drive OAuth exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تبادل رمز التفويض", "details": err.Error()})
		return
	}

	// Save tokens
	tokenJSON, _ := json.Marshal(token)
	h.upsertSetting("drive_oauth_token", string(tokenJSON))

	// Get user email for display
	srv, err := drive.NewService(context.Background(), option.WithTokenSource(cfg.TokenSource(context.Background(), token)))
	if err == nil {
		about, err := srv.About.Get().Fields("user").Do()
		if err == nil && about.User != nil {
			h.upsertSetting("drive_connected_email", about.User.EmailAddress)
		}
	}

	// Redirect to settings page with success
	frontendURL := "https://qa.uoturath.edu.iq"
	var feURL string
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'frontend_url'").Scan(&feURL)
	if feURL != "" {
		frontendURL = feURL
	}
	c.Redirect(http.StatusFound, frontendURL+"/admin/settings?drive=connected")
}

// Status returns the current Drive connection status
func (h *DriveOAuthHandler) Status(c *gin.Context) {
	var tokenStr, email, folderID string
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_oauth_token'").Scan(&tokenStr)
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_connected_email'").Scan(&email)
	h.db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_folder_id'").Scan(&folderID)

	connected := false
	if tokenStr != "" {
		var token oauth2.Token
		if err := json.Unmarshal([]byte(tokenStr), &token); err == nil {
			connected = token.RefreshToken != ""
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":  connected,
		"email":      email,
		"folder_id":  folderID,
	})
}

// Disconnect removes stored Drive credentials
func (h *DriveOAuthHandler) Disconnect(c *gin.Context) {
	h.db.Exec("DELETE FROM system_settings WHERE key IN ('drive_oauth_token', 'drive_connected_email')")
	c.JSON(http.StatusOK, gin.H{"message": "تم قطع الاتصال بـ Google Drive"})
}

// UpdateFolder updates the target Drive folder
func (h *DriveOAuthHandler) UpdateFolder(c *gin.Context) {
	var req struct {
		FolderID string `json:"folder_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المجلد مطلوب"})
		return
	}
	h.upsertSetting("drive_folder_id", req.FolderID)
	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث مجلد Drive"})
}

func (h *DriveOAuthHandler) upsertSetting(key, value string) {
	h.db.Exec(`
		INSERT INTO system_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, key, value)
}
