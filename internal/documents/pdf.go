package documents

import (
	"bytes"
	"fmt"
	"io"
	// "os/exec"
	"strings"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type PDFProcessor struct {
	templatePath string
}

func NewPDFProcessor(templatePath string) (*PDFProcessor, error) {
	if templatePath != "" {
		// Validate template existence (optional, could log instead).
		doc, err := document.Open(templatePath)
		if err != nil {
				return nil, fmt.Errorf("invalid template path %s: %w", templatePath, err)
		}
		doc.Close() // Close since we’re just validating.
}
	return &PDFProcessor{
		templatePath: templatePath,
	}, nil
}

func (p *PDFProcessor) ExtractText(file io.Reader) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("reading PDF file: %w", err)
	}

	// create a reader for the PDF
	pdfReader, err := model.NewPdfReader(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("parsing PDF: %w", err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("getting page count: %w", err)
	}

	var allText string
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return "", fmt.Errorf("getting page %d: %w", i, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", fmt.Errorf("creating extractor for page %d: %w", i, err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			return "", fmt.Errorf("extracting text from page %d: %w", i, err)
		}

		allText += text + "\n"
	}

	return allText, nil
}

func (p *PDFProcessor) CreateFormattedDocument(content string) ([]byte, error) {
	creator := creator.New()

	// set document properties
	creator.SetPageMargins(50, 50, 50, 50)
	creator.NewPage()

	// add content to the document
	// textStyle := creator.NewTextStyle()
	// textStyle.FontSize = 11

	// split content into paragraphs
	paragraphs := strings.Split(content, "\n\n")
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			continue
		}

		// create a paragraph
		p := creator.NewParagraph(paragraph)
		p.SetMargins(0, 0, 10, 10)

		// add the paragraph to the creator
		err := creator.Draw(p)
		if err != nil {
			return nil, fmt.Errorf("drawing paragraph: %w", err)
		}
	}

	// output bytes
	var buf bytes.Buffer
	err := creator.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
