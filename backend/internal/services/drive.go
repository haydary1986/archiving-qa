package services

import (
	"context"
	"fmt"
	"io"

	"github.com/haydary1986/archiving-qa/internal/config"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveService handles interactions with Google Drive API.
type DriveService struct {
	client        *drive.Service
	rootFolderID  string
}

// NewDriveService initializes a new DriveService using the provided Google configuration.
// It authenticates using a service account key file.
func NewDriveService(cfg *config.GoogleConfig) (*DriveService, error) {
	ctx := context.Background()

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(cfg.ServiceAccountKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	return &DriveService{
		client:       srv,
		rootFolderID: cfg.DriveFolderID,
	}, nil
}

// UploadFile uploads a file to Google Drive within the specified folder.
// It returns the file ID, web view link, and any error encountered.
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
		Context(ctx).
		Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file to drive: %w", err)
	}

	return created.Id, created.WebViewLink, nil
}

// DownloadFile downloads a file from Google Drive by its file ID.
// The caller is responsible for closing the returned ReadCloser.
func (s *DriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := s.client.Files.Get(fileID).
		Context(ctx).
		Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file from drive: %w", err)
	}

	return resp.Body, nil
}

// DeleteFile permanently deletes a file from Google Drive.
func (s *DriveService) DeleteFile(ctx context.Context, fileID string) error {
	err := s.client.Files.Delete(fileID).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to delete file from drive: %w", err)
	}

	return nil
}

// CreateShareLink creates a publicly accessible link for the specified file.
// It sets the file permission to "anyone with the link can read".
func (s *DriveService) CreateShareLink(ctx context.Context, fileID string) (string, error) {
	perm := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	_, err := s.client.Permissions.Create(fileID, perm).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to create share permission: %w", err)
	}

	file, err := s.client.Files.Get(fileID).
		Fields("webViewLink").
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to get file share link: %w", err)
	}

	return file.WebViewLink, nil
}

// CreateFolder creates a new folder in Google Drive under the specified parent.
// It returns the newly created folder's ID.
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
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to create folder in drive: %w", err)
	}

	return created.Id, nil
}
