package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OCRService provides optical character recognition capabilities
// using Tesseract OCR for text extraction from images and PDFs.
type OCRService struct{}

// NewOCRService creates a new OCRService instance.
func NewOCRService() *OCRService {
	return &OCRService{}
}

// ExtractText extracts text from an image file using Tesseract OCR.
// It supports both Arabic and English languages.
func (s *OCRService) ExtractText(ctx context.Context, filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	cmd := exec.CommandContext(ctx, "tesseract", filePath, "stdout", "-l", "ara+eng")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("tesseract failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to run tesseract: %w", err)
	}

	text := strings.TrimSpace(string(output))
	return text, nil
}

// ExtractTextFromPDF extracts text from a PDF file by first converting
// each page to an image using pdftoppm, then running OCR on each image.
func (s *OCRService) ExtractTextFromPDF(ctx context.Context, filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Create a temporary directory to store converted images
	tempDir, err := os.MkdirTemp("", "ocr-pdf-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Convert PDF pages to PNG images using pdftoppm
	outputPrefix := filepath.Join(tempDir, "page")
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", "300", filePath, outputPrefix)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to convert PDF to images: %s: %w", string(output), err)
	}

	// Find all generated page images
	pattern := filepath.Join(tempDir, "page-*.png")
	pages, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to find converted pages: %w", err)
	}

	if len(pages) == 0 {
		return "", fmt.Errorf("no pages were extracted from the PDF")
	}

	// Run OCR on each page and collect the results
	var results []string
	for i, pagePath := range pages {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		text, err := s.ExtractText(ctx, pagePath)
		if err != nil {
			// Log the error but continue with remaining pages
			results = append(results, fmt.Sprintf("[صفحة %d: فشل في استخراج النص]", i+1))
			continue
		}

		if text != "" {
			results = append(results, fmt.Sprintf("--- صفحة %d ---\n%s", i+1, text))
		}
	}

	if len(results) == 0 {
		return "", fmt.Errorf("failed to extract text from any page of the PDF")
	}

	return strings.Join(results, "\n\n"), nil
}
