package documents

import (
	"fmt"
	"os/exec"
)

type PDFProcessor struct{}

func (p *PDFProcessor) Extract(filePath string) (string, error) {
	cmd := exec.Command("pdftotext", filePath, "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %v, output: %s", err, output)
	}
	return string(output), nil
}

func (p *PDFProcessor) Create(context string, outputPath string) (string, error) {
	return "", fmt.Errorf("PDF creation not implemented yet")
}
