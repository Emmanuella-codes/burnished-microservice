package documents

import (
	"fmt"
	"io"

	"github.com/Emmanuella-codes/burnished-microservice/internal/ai"
	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
)

type Processor struct {
	config *config.Config
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

func (p *Processor) FormatForATS(file io.Reader, fileExt, jobDesc string) ([]byte, error) {
	var processor DocumentProcessor
	var err error

	switch fileExt {
	case ".pdf":
		processor, err = NewPDFProcessor(p.config.PdfTemplate)
	case ".docx":
		processor, err = NewPDFProcessor(p.config.DocxTemplate)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", fileExt)
	}

	if err != nil {
		return nil, err
	}

	// extract text from cv
	text, err := processor.ExtractText(file)
	if err != nil {
		return nil, err
	}

	// use AI to optimize the CV content
	optimizedText, err := ai.FormatForATS(text, jobDesc, p.config.GeminiAPIKey)
	if err != nil {
		return nil, err
	}

	// create a new document with the optimized content
	formattedDoc, err := processor.CreateFormattedDocument(optimizedText)
	if err != nil {
		return nil, err
	}

	return formattedDoc, nil
}

func (p *Processor) RoastCV(file io.Reader, fileExt string) (string, error) {
	var processor DocumentProcessor
	var err error

	switch fileExt {
	case ".pdf":
		processor, err = NewPDFProcessor(p.config.PdfTemplate)
	case ".docx":
		processor, err = NewDOCXProcessor(p.config.DocxTemplate)
	default:
		return "", fmt.Errorf("unsupported file format: %s", fileExt)
	}

	if err != nil {
		return "", err
	}

	// extract text from cv
	text, err := processor.ExtractText(file)
	if err != nil {
		return "", err
	}

	// use AI to critique the CV
	feedback, err := ai.RoastCV(text, p.config.GeminiAPIKey)
	if err != nil {
		return "", err
	}

	return feedback, nil
}

type DocumentProcessor interface {
	ExtractText(file io.Reader) (string, error)
	CreateFormattedDocument(content string) ([]byte, error)
}
