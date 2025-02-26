package documents

import (
	"bytes"
	"io"

	"github.com/Emmanuella-codes/burnished-microservice/pkg/utils"
	"github.com/unidoc/unioffice/document"
)

type DOCXProcessor struct {
	templatePath string
}

func NewDOCXProcessor(templatePath string) (*DOCXProcessor, error) {
	return &DOCXProcessor{
		templatePath: templatePath,
	}, nil
}

func (p *DOCXProcessor) ExtractText(file io.Reader) (string, error) {
	// read the entire file
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Create a bytes.Reader from the byte slice
	reader := bytes.NewReader(fileBytes)

	// open the DOCX file
	doc, err := document.Read(reader, reader.Size())
	if err != nil {
		return "", err
	}

	var text string
	// extract text from paragraphs
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			text += run.Text()
		}
		text += "\n"
	}

	return text, nil
}

func (p *DOCXProcessor) CreateFormattedDocument(content string) ([]byte, error) {
	// load the template
	doc, err := document.Open(p.templatePath)
	if err != nil {
		// create a new document if the template doesn't exist
		doc = document.New()
		para := doc.AddParagraph()
		run := para.AddRun()
		run.AddText(content)
	} else {
		// clear the template and add the new document
		for i := len(doc.Paragraphs()) - 1; i >= 0; i-- {
			doc.RemoveParagraph(doc.Paragraphs()[i])
		}

		// split content by new lines and paragraph
		lines := utils.SplitLines(content)
		for _, line := range lines {
			para := doc.AddParagraph()
			run := para.AddRun()
			run.AddText(line)
		}
	}

	// save the document to a byte array
	buf := new(bytes.Buffer)
	err = doc.Save(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
