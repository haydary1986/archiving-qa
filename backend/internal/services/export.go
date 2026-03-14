package services

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/haydary1986/archiving-qa/internal/models"
	"github.com/xuri/excelize/v2"
)

// ExportService handles exporting documents to Excel and ZIP formats.
type ExportService struct {
	db           *sql.DB
	driveService *DriveService
}

// NewExportService creates a new ExportService with the given database
// connection and DriveService for file downloads.
func NewExportService(db *sql.DB, driveService *DriveService) *ExportService {
	return &ExportService{
		db:           db,
		driveService: driveService,
	}
}

// ExportToExcel creates an Excel file containing document metadata
// with Arabic column headers. Returns the file content as bytes.
func (s *ExportService) ExportToExcel(ctx context.Context, documents []models.Document) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "الوثائق"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Delete the default sheet
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// Ignore error if default sheet doesn't exist
		_ = err
	}

	// Set right-to-left direction for Arabic
	if err := f.SetSheetViewOptions(sheetName, 0, excelize.RightToLeft(true)); err != nil {
		return nil, fmt.Errorf("failed to set RTL: %w", err)
	}

	// Define Arabic headers
	headers := []string{
		"التسلسل",
		"عنوان الوثيقة",
		"رقم الكتاب",
		"تاريخ الكتاب",
		"نوع الوثيقة",
		"التصنيف",
		"الجهة المصدرة",
		"الجهة المستلمة",
		"الحالة",
		"الوصف",
		"تاريخ الإنشاء",
	}

	// Create header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		// Continue without style if it fails
		headerStyle = 0
	}

	// Write headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("failed to set header cell %s: %w", cell, err)
		}
		if headerStyle > 0 {
			_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	// Write document data
	for rowIdx, doc := range documents {
		row := rowIdx + 2 // start from row 2 (after headers)

		values := []interface{}{
			rowIdx + 1,
			doc.Title,
			doc.DocumentNumber,
			formatDocumentDate(doc.DocumentDate),
			translateDocumentType(doc.DocumentType),
			translateClassification(doc.Classification),
			doc.SourceEntity,
			doc.DestEntity,
			translateStatus(doc.Status),
			doc.Description,
			doc.CreatedAt.Format("2006-01-02 15:04"),
		}

		for colIdx, val := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			if err := f.SetCellValue(sheetName, cell, val); err != nil {
				return nil, fmt.Errorf("failed to set cell %s: %w", cell, err)
			}
		}
	}

	// Set column widths for readability
	columnWidths := map[string]float64{
		"A": 8,  // التسلسل
		"B": 35, // عنوان الوثيقة
		"C": 15, // رقم الكتاب
		"D": 15, // تاريخ الكتاب
		"E": 15, // نوع الوثيقة
		"F": 12, // التصنيف
		"G": 25, // الجهة المصدرة
		"H": 25, // الجهة المستلمة
		"I": 12, // الحالة
		"J": 40, // الوصف
		"K": 18, // تاريخ الإنشاء
	}
	for col, width := range columnWidths {
		_ = f.SetColWidth(sheetName, col, col, width)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportToZip creates a ZIP archive containing an Excel summary file
// and the actual document files downloaded from Google Drive.
func (s *ExportService) ExportToZip(ctx context.Context, documents []models.Document) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	// Generate and add Excel file to the ZIP
	excelData, err := s.ExportToExcel(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to generate excel for zip: %w", err)
	}

	excelWriter, err := zipWriter.Create("الوثائق.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create excel entry in zip: %w", err)
	}
	if _, err := excelWriter.Write(excelData); err != nil {
		return nil, fmt.Errorf("failed to write excel to zip: %w", err)
	}

	// Download and add each document's files from Drive
	for _, doc := range documents {
		for _, file := range doc.Files {
			if file.DriveFileID == "" {
				continue
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			// Create a subdirectory per document
			filePath := fmt.Sprintf("%s/%s", sanitizeFolderName(doc.Title), file.OriginalName)

			reader, err := s.driveService.DownloadFile(ctx, file.DriveFileID)
			if err != nil {
				// Skip files that fail to download but continue with the rest
				continue
			}

			fileWriter, err := zipWriter.Create(filePath)
			if err != nil {
				reader.Close()
				continue
			}

			if _, err := io.Copy(fileWriter, reader); err != nil {
				reader.Close()
				continue
			}
			reader.Close()
		}
	}

	// Close the zip writer to flush all data
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize zip: %w", err)
	}

	return buf.Bytes(), nil
}

// formatDocumentDate formats a document date pointer to a string.
func formatDocumentDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// translateDocumentType converts document type codes to Arabic labels.
func translateDocumentType(docType string) string {
	switch docType {
	case "incoming":
		return "وارد"
	case "outgoing":
		return "صادر"
	case "internal":
		return "داخلي"
	default:
		return docType
	}
}

// translateClassification converts classification codes to Arabic labels.
func translateClassification(classification string) string {
	switch classification {
	case "normal":
		return "عادي"
	case "confidential":
		return "سري"
	case "secret":
		return "سري للغاية"
	default:
		return classification
	}
}

// translateStatus converts status codes to Arabic labels.
func translateStatus(status string) string {
	switch status {
	case "draft":
		return "مسودة"
	case "processing":
		return "قيد المعالجة"
	case "completed":
		return "مكتمل"
	case "archived":
		return "مؤرشف"
	default:
		return status
	}
}

// sanitizeFolderName removes or replaces characters that are not safe
// for use in file/folder names inside a ZIP archive.
func sanitizeFolderName(name string) string {
	if name == "" {
		return "untitled"
	}

	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	result := r.Replace(name)

	// Truncate long names
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}
