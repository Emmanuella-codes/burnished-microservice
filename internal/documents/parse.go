package documents

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type Formatter struct {
	config *config.Config
}

func NewFormatter(cfg *config.Config) *Formatter {
	return &Formatter{
		config: cfg,
	}
}

type CVSections struct {
	Education   []string
	Experiences []string
	Projects    []string
	Skills      []string
}

func (f *Formatter) ParseCV(fileData []byte) (*CVSections, error) {
	pdfReader, err := model.NewPdfReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	sections := &CVSections{}
	var currSection string

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			continue
		}

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
	}
	return sections, nil
}

func (f *Formatter) Format(sections *CVSections, jobDescription string) ([]byte, error) {
	c := creator.New()

	font, err := model.NewStandard14Font("Helvetica")
	if err != nil {
		return nil, fmt.Errorf("failed to load standard font: %w", err)
	}
	c.NewPage()
	p := c.NewParagraph("")
	p.SetFont(font)
	p.SetFontSize(16)
	p.SetPos(50, 50)
	c.Draw(p)
	y := 100.0

	// skills
	y += 10
	p = c.NewParagraph("Skills")
	p.SetFont(font)
	p.SetFontSize(14)
	p.SetPos(50, y)
	c.Draw(p)
	y += 20
	for _, skill := range sections.Skills {
		p = c.NewParagraph(skill)
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetPos(50, y)
		c.Draw(p)
		y += 15
	}

	// experience
	p.SetFont(font)
	p.SetFontSize(14)
	p.SetPos(50, y)
	c.Draw(p)
	y += 20
	for _, exp := range sections.Experiences {
		p = c.NewParagraph(exp)
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetPos(50, y)
		c.Draw(p)
		y += 15
	}

	// education
	p = c.NewParagraph("Education")
	p.SetFont(font)
	p.SetFontSize(14)
	p.SetPos(50, y)
	c.Draw(p)
	y += 20
	for _, edu := range sections.Education {
		p = c.NewParagraph(edu)
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetPos(50, y)
		c.Draw(p)
		y += 15
	}

	// projects
	y += 10
	p = c.NewParagraph("Projects")
	p.SetFont(font)
	p.SetFontSize(14)
	p.SetPos(50, y)
	c.Draw(p)
	y += 20
	for _, proj := range sections.Projects {
		p = c.NewParagraph(proj)
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetPos(50, y)
		c.Draw(p)
		y += 15
	}

	var buf bytes.Buffer
	if err := c.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}
