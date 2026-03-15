package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveService handles interactions with Google Drive API.
type DriveService struct {
	client       *drive.Service
	rootFolderID string
}

// NewDriveServiceFromOAuth creates a DriveService using OAuth2 tokens stored in the database.
// This is the preferred method — triggered by the "Connect Drive" button in settings.
func NewDriveServiceFromOAuth(db *sql.DB) (*DriveService, error) {
	var clientID, clientSecret, tokenStr, folderID string
	db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_client_id'").Scan(&clientID)
	db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_client_secret'").Scan(&clientSecret)
	db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_oauth_token'").Scan(&tokenStr)
	db.QueryRow("SELECT value FROM system_settings WHERE key = 'drive_folder_id'").Scan(&folderID)

	if clientID == "" || clientSecret == "" || tokenStr == "" {
		return nil, fmt.Errorf("Drive OAuth not configured")
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		return nil, fmt.Errorf("invalid stored token: %w", err)
	}

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{drive.DriveScope},
		Endpoint:     google.Endpoint,
	}

	tokenSource := cfg.TokenSource(context.Background(), &token)

	// Check if token was refreshed and save the new one
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	if newToken.AccessToken != token.AccessToken {
		newTokenJSON, _ := json.Marshal(newToken)
		db.Exec(`UPDATE system_settings SET value = $1, updated_at = NOW() WHERE key = 'drive_oauth_token'`, string(newTokenJSON))
	}

	srv, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	log.Println("Drive: connected via OAuth2")
	return &DriveService{
		client:       srv,
		rootFolderID: folderID,
	}, nil
}

// NewDriveServiceFromServiceAccount creates a DriveService using a service account key.
// Supports both file path and inline JSON. If impersonateEmail is set, uses domain-wide delegation.
func NewDriveServiceFromServiceAccount(saKey, folderID, impersonateEmail string) (*DriveService, error) {
	if saKey == "" {
		return nil, fmt.Errorf("service account key not configured")
	}

	ctx := context.Background()
	var keyData []byte
	if len(saKey) > 0 && saKey[0] == '{' {
		keyData = []byte(saKey)
	} else {
		var err error
		keyData, err = os.ReadFile(saKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account key file: %w", err)
		}
	}

	var srv *drive.Service
	var err error

	if impersonateEmail != "" {
		log.Printf("Drive: using impersonation as %s", impersonateEmail)
		jwtCfg, err := google.JWTConfigFromJSON(keyData, drive.DriveScope)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service account key: %w", err)
		}
		jwtCfg.Subject = impersonateEmail
		client := jwtCfg.Client(ctx)
		srv, err = drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service with impersonation: %w", err)
		}
	} else {
		log.Println("Drive: using direct service account (requires Shared Drive)")
		srv, err = drive.NewService(ctx, option.WithCredentialsJSON(keyData))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service: %w", err)
		}
	}

	return &DriveService{
		client:       srv,
		rootFolderID: folderID,
	}, nil
}

// GetDriveService tries OAuth2 first (from DB), then falls back to service account (from env).
// This is the main entry point used by routes.Setup.
func GetDriveService(db *sql.DB, saKey, folderID, impersonateEmail string) *DriveService {
	// Try OAuth2 first
	ds, err := NewDriveServiceFromOAuth(db)
	if err == nil {
		return ds
	}

	// Fall back to service account
	if saKey != "" {
		ds, err = NewDriveServiceFromServiceAccount(saKey, folderID, impersonateEmail)
		if err != nil {
			log.Printf("Drive service account init failed: %v", err)
			return nil
		}
		return ds
	}

	log.Println("Drive: not configured (no OAuth token and no service account)")
	return nil
}

// UploadFile uploads a file to Google Drive within the specified folder.
func (s *DriveService) UploadFile(ctx context.Context, fileName string, content io.Reader, mimeType string, folderID string) (fileID string, webViewLink string, err error) {
	if folderID == "" {
		folderID = s.rootFolderID
	}

	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{folderID},
	}

	created, err := s.client.Files.Create(driveFile).
		Media(content).
		Fields("id, webViewLink").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file to drive: %w", err)
	}

	return created.Id, created.WebViewLink, nil
}

// DownloadFile downloads a file from Google Drive by its file ID.
func (s *DriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := s.client.Files.Get(fileID).
		SupportsAllDrives(true).
		Context(ctx).
		Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file from drive: %w", err)
	}
	return resp.Body, nil
}

// DeleteFile permanently deletes a file from Google Drive.
func (s *DriveService) DeleteFile(ctx context.Context, fileID string) error {
	return s.client.Files.Delete(fileID).
		SupportsAllDrives(true).
		Context(ctx).
		Do()
}

// CreateShareLink creates a publicly accessible link for the specified file.
func (s *DriveService) CreateShareLink(ctx context.Context, fileID string) (string, error) {
	perm := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	_, err := s.client.Permissions.Create(fileID, perm).
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to create share permission: %w", err)
	}

	file, err := s.client.Files.Get(fileID).
		Fields("webViewLink").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to get file share link: %w", err)
	}

	return file.WebViewLink, nil
}

// CreateFolder creates a new folder in Google Drive under the specified parent.
func (s *DriveService) CreateFolder(ctx context.Context, name string, parentID string) (string, error) {
	if parentID == "" {
		parentID = s.rootFolderID
	}

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	created, err := s.client.Files.Create(folder).
		Fields("id").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to create folder in drive: %w", err)
	}

	return created.Id, nil
}
