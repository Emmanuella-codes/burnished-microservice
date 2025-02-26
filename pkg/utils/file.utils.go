package utils

import (
	"bytes"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// generatePDF creates a PDF document from text content
func GeneratePDF(content, templatePath string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	lines := SplitLines(content)
	for _, line := range lines {
		pdf.MultiCell(0, 10, line, "0", "L", false)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// split text content into lines
func SplitLines(content string) []string {
	return strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
}