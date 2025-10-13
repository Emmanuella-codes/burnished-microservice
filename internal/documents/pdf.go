package documents

import (
	"bytes"
	"fmt"
	"io"
	"log"

	// "os/exec"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/ledongthuc/pdf"
)

type PDFProcessor struct {

}

func NewPDFProcessor() *PDFProcessor {
	return &PDFProcessor{}
}

func (p *PDFProcessor) ExtractText(file io.Reader) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[ExtractText] Failed to read file: %v", err)
		return "", fmt.Errorf("reading PDF file: %w", err)
	}

	// create a reader for the PDF
	reader := bytes.NewReader(fileBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(fileBytes)))
	if err != nil {
		log.Printf("[ExtractText] Failed to parse PDF: %v", err)
		return "", fmt.Errorf("parsing PDF: %w", err)
	}

	var allText string
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("error extracting text from page %d: %w", i, err)
		}

		allText += text + "\n"
	}

	return allText, nil
}

func (p *PDFProcessor) CreateFormattedDocument(content string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// set margins
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// set default font
	pdf.SetFont("Arial", "", 11)

	// split content into paragraphs
	paragraphs := strings.Split(content, "\n\n")

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// handle long paragraphs that might need wrapping
		lines := strings.Split(paragraph, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				pdf.Ln(5) // add space for empty lines
				continue
			}

			// multicell for automatic text wrapping
			pdf.MultiCell(0, 6, line, "", "", false)
			pdf.Ln(2) // small space between lines
		}

		pdf.Ln(5) // space between paragraphs
	}

	// output to bytes buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("generating PDF: %w", err)
	}

	return buf.Bytes(), nil
}
