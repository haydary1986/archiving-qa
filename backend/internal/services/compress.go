package services

import (
	"fmt"
	"os/exec"
)

// CompressService provides file compression capabilities
// for images (via ImageMagick) and PDFs (via Ghostscript).
type CompressService struct{}

// NewCompressService creates a new CompressService instance.
func NewCompressService() *CompressService {
	return &CompressService{}
}

// CompressImage compresses an image file using ImageMagick's convert command.
// It reduces quality to 75% and resizes large images while preserving readability.
func (s *CompressService) CompressImage(inputPath, outputPath string) error {
	cmd := exec.Command(
		"convert",
		inputPath,
		"-quality", "75",
		"-resize", "1920x1920>",
		"-strip",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compress image: %s: %w", string(output), err)
	}

	return nil
}

// CompressPDF compresses a PDF file using Ghostscript.
// It uses the /ebook setting which provides a good balance between
// file size reduction and readable quality.
func (s *CompressService) CompressPDF(inputPath, outputPath string) error {
	cmd := exec.Command(
		"gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/ebook",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		fmt.Sprintf("-sOutputFile=%s", outputPath),
		inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compress PDF: %s: %w", string(output), err)
	}

	return nil
}
