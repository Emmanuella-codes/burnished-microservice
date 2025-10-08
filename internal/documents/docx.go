package documents

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unioffice/document"
)

type DOCXProcessor struct {}

func NewDOCXProcessor() *DOCXProcessor {
	return &DOCXProcessor{}
}

// ExtractText extracts plain text from a DOCX file.
func (p *DOCXProcessor) ExtractText(file io.Reader) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("reading DOCX file: %w", err)
	}

	// Open DOCX from in-memory buffer
	doc, err := document.Read(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("parsing DOCX: %w", err)
	}
	defer doc.Close()

	var text strings.Builder
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			text.WriteString(run.Text())
		}
		text.WriteString("\n")
	}

	return text.String(), nil
}

// CreateFormattedDocument creates a new DOCX from content.
// If a templatePath was provided, it will clear and reuse it.
func (p *DOCXProcessor) CreateFormattedDocument(content string) ([]byte, error) {
	var doc *document.Document
	var err error

	// Split content by lines
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			para := doc.AddParagraph()
			run := para.AddRun()
			run.AddText(line)
		}
	}

	// Write to memory
	buf := new(bytes.Buffer)
	err = doc.Save(buf)
	if err != nil {
		return nil, fmt.Errorf("saving DOCX: %w", err)
	}

	return buf.Bytes(), nil
}
