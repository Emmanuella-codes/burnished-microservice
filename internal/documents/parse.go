package documents

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
	"github.com/jung-kurt/gofpdf"
)

type Formatter struct {
	config 			 *config.Config
	pdfProcessor *PDFProcessor
}

func NewFormatter(cfg *config.Config, pdfProcessor *PDFProcessor) *Formatter {
	return &Formatter{
		config: 			cfg,
		pdfProcessor: pdfProcessor,
	}
}

type CVSections struct {
	Education   []string
	Experiences []string
	Projects    []string
	Skills      []string
}

func (f *Formatter) ParseCV(fileData []byte) (*CVSections, error) {
	reader := bytes.NewReader(fileData)

	text, err := f.pdfProcessor.ExtractText(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
		}

	sections := &CVSections{}
	var currSection string

	// split text into lines
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "education") {
			currSection = "education"
			continue
		} else if strings.Contains(lowerLine, "experience") || strings.Contains(lowerLine, "work history") {
			currSection = "experience"
			continue
		} else if strings.Contains(lowerLine, "projects") {
			currSection = "projects"
			continue
		} else if strings.Contains(lowerLine, "skills") {
			currSection = "skills"
			continue
		}

		switch currSection {
		case "education":
			sections.Education = append(sections.Education, line)
		case "experience":
			sections.Experiences = append(sections.Experiences, line)
		case "projects":
			sections.Projects = append(sections.Projects, line)
		case "skills":
			sections.Skills = append(sections.Skills, line)
		}
	}
	return sections, nil
}

func (f *Formatter) Format(sections *CVSections, jobDescription string) ([]byte, error) {
	if sections == nil {
		return nil, fmt.Errorf("sections cannot be nil")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// set margins
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// skills section
	if len(sections.Skills) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Skills")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 12)
		for _, skill := range sections.Skills {
			pdf.Cell(0, 6, skill)
			pdf.Ln(6)
		}
		pdf.Ln(5)
	}

	// experience section
	if len(sections.Experiences) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Experience")
		pdf.Ln(15)

		pdf.SetFont("Arial", "", 12)
		for _, exp := range sections.Experiences {
			pdf.Cell(0, 6, exp)
			pdf.Ln(8)
		}
		pdf.Ln(5)
	}

	// education section
	if len(sections.Education) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Education")
		pdf.Ln(15)
		
		pdf.SetFont("Arial", "", 12)
		for _, edu := range sections.Education {
			pdf.Cell(0, 6, edu)
			pdf.Ln(8)
		}
		pdf.Ln(5)
	}

	// projects section
	if len(sections.Projects) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Projects")
		pdf.Ln(15)
		
		pdf.SetFont("Arial", "", 12)
		for _, proj := range sections.Projects {
			pdf.Cell(0, 6, proj)
			pdf.Ln(8)
		}
	}

	var buf bytes.Buffer
	 err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}
