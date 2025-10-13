package documents

import (
	"fmt"
	"io"
	"log"

	"github.com/Emmanuella-codes/burnished-microservice/internal/ai"
	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
	"github.com/Emmanuella-codes/burnished-microservice/internal/dtos"
)

type Processor struct {
	config *config.Config
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

func (p *Processor) FormatForATS(file io.Reader, fileExt, jobDesc string) (*dtos.Resume, error) {
	var processor DocumentProcessor
	var err error

	switch fileExt {
	case ".pdf":
		processor = NewPDFProcessor()
	case ".docx":
		processor = NewDOCXProcessor()
	default:
		return nil, fmt.Errorf("unsupported file format: %s", fileExt)
	}

	// extract text from cv
	text, err := processor.ExtractText(file)
	if err != nil {
		return nil, fmt.Errorf("extracting text from %s: %w", fileExt, err)
	}

	log.Printf("Extracted text length: %d characters", len(text))
    if len(text) == 0 {
        return nil, fmt.Errorf("extracted text is empty")
    }

	// parse + optimize CV into structured JSON
	resume, err := ai.ParseAndOptimizeCV(text, jobDesc, p.config.GeminiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("optimizing CV for ATS: %w", err)
	}

	return resume, nil
}

func (p *Processor) RoastCV(file io.Reader, fileExt string) (string, error) {
	var processor DocumentProcessor
	var err error

	switch fileExt {
	case ".pdf":
		processor = NewPDFProcessor()
	case ".docx":
		processor = NewDOCXProcessor()
	default:
		return "", fmt.Errorf("unsupported file format: %s", fileExt)
	}

	// extract text from cv
	text, err := processor.ExtractText(file)
	if err != nil {
		return "", fmt.Errorf("extracting text from %s: %w", fileExt, err)
	}

	// use AI to critique the CV
	feedback, err := ai.RoastCV(text, p.config.GeminiAPIKey)
	if err != nil {
		return "", fmt.Errorf("roasting CV: %w", err)
	}

	return feedback, nil
}

type DocumentProcessor interface {
	ExtractText(file io.Reader) (string, error)
	CreateFormattedDocument(content string) ([]byte, error)
}
