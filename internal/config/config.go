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
	cfg := &Config{
		Port:         "8080",                  
		MaxFileSize:  10 * 1024 * 1024, // Default: 10MB.
		DocxTemplate: "./templates/cv_template.docx", 
		PdfTemplate:  "./templates/cv_template.pdf",
	}
	
	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}

	cfg.GeminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	if docxTemplate := os.Getenv("DOCX_TEMPLATE"); docxTemplate != "" {
		cfg.DocxTemplate = docxTemplate
	}

	if pdfTemplate := os.Getenv("PDF_TEMPLATE"); pdfTemplate != "" {
		cfg.PdfTemplate = pdfTemplate
	}

	if maxFileSizeStr := os.Getenv("MAX_FILE_SIZE"); maxFileSizeStr != "" {
		size, err := strconv.ParseInt(maxFileSizeStr, 10, 64)
		if err != nil {
				return nil, fmt.Errorf("invalid MAX_FILE_SIZE value %q: %w", maxFileSizeStr, err)
		}
		if size <= 0 {
				return nil, fmt.Errorf("MAX_FILE_SIZE must be positive, got %d", size)
		}
		cfg.MaxFileSize = size
	}

	cfg.StorageBucket = os.Getenv("STORAGE_BUCKET")
    if cfg.StorageBucket == "" {
        return nil, fmt.Errorf("environment variable STORAGE_BUCKET is required")
    }

	return cfg, nil
}
