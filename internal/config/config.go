package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port 					string
	GeminiAPIKey  string
	DocxTemplate  string
	PdfTemplate   string
	MaxFileSize   int64
	StorageBucket string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	docxTemplate := os.Getenv("DOCX_TEMPLATE")
	if docxTemplate == "" {
		docxTemplate = "./templates/cv_template.docx"
	}

	pdfTemplate := os.Getenv("PDF_TEMPLATE")
	if docxTemplate == "" {
		docxTemplate = "./templates/cv_template.pdf"
	}

	maxFileSizeStr := os.Getenv("MAX_FILE_SIZE")
	maxFileSize := int64(10 * 1024 * 1024)
	if maxFileSizeStr != "" {
		size, err := strconv.ParseInt(maxFileSizeStr, 10, 64)
		if err == nil {
			maxFileSize = size
		}
	}

	return &Config{
		Port: 				port,
		GeminiAPIKey: geminiAPIKey,
		DocxTemplate: docxTemplate,
		PdfTemplate: 	pdfTemplate,
		MaxFileSize: 	maxFileSize,
	}, nil
}
