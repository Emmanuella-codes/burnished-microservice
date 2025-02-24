package documents

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/gomutex/godocx"
)

type DocxProcessor struct{}

func (p *DocxProcessor) Extract(filePath string) (string, error) {
	// open the docx file 
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open docx file: %v", err)
	}
	defer r.Close()

	// find document.xml in the archive
	var documentFile *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			documentFile = f
			break
		}
	}

	if documentFile == nil {
		return "", fmt.Errorf("document.xml not found in docx file")
	}

	// read document.xml
	documentReader, err := documentFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open document.xml: %v", err)
	}
	defer documentReader.Close()

	documentBytes, err := io.ReadAll(documentReader)
	if err != nil {
		return "", fmt.Errorf("failed to read document.xml: %v", err)
	}

	// extract text from xml
	text, err := extractTextFromDocumentXML(documentBytes)
	if err != nil {
		return "", fmt.Errorf("failed to extract text from document.xml: %v", err)
	}
	return text, nil
}

func (p *DocxProcessor) Create(content string, outputPath string) (string, error) {
	doc, err := godocx.NewDocument()
	if err != nil {
		return "", fmt.Errorf("failed to create document: %v", err)
	}
	defer doc.Close()

	doc.AddParagraph(content)
	// save document to specified output path
	if err = doc.Save(); err != nil {
		return "", fmt.Errorf("failed to save document: %v", err)
	}

	
	return outputPath, nil
}

func extractTextFromDocumentXML(documentBytes []byte) (string, error) {
	type Text struct {
		XMLName xml.Name `xml:"t"`
		Content string 	 `xml:",chardata"`
	}

	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(documentBytes))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				var text Text
				if err := decoder.DecodeElement(&text, &se); err != nil {
					return "", err
				}
				buf.WriteString(text.Content)
				buf.WriteString(" ")
			}
		}
	}

	return strings.TrimSpace(buf.String()), nil
}
