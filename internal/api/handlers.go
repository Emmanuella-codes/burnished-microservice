package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Emmanuella-codes/burnished-microservice/internal/documents"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status": "ok",
		"time": 	time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) processCVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse multipart form
	if err := r.ParseMultipartForm(s.cfg.MaxFileSize); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// get the file from the request
	file, header, err := r.FormFile("cv")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// generate a temporary filename
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	tempFile := filepath.Join(s.cfg.TempDir, "cv_"+time.Now().Format("20060102150405")+fileExt)

	// validate file extension
	if fileExt != ".pdf" && fileExt != ".docx" {
		http.Error(w, "Unsupported file format. Please upload PDF or DOCX", http.StatusBadRequest)
		return
	}

	// save the uploaded file
	dst, err := os.Create(tempFile)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Error creating temp file: %v", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Error copying file: %v", err)
		return
	}

	// extract cv content
	var processor documents.Processor
	if fileExt == ".pdf" {
		processor = &documents.DocxProcessor{}
	} else {
		processor = &documents.PDFProcessor{}
	}

	content, err := processor.Extract(tempFile)
	if err != nil {
		http.Error(w, "Failed to extract content from file", http.StatusInternalServerError)
		log.Printf("Error extracting content: %v", err)
		return
	}

	// extracted content
	response := map[string]string{
		"content": content,
		"file": 	 tempFile,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) formatCVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CVContent 		 string `json:"cvContent"`
		JobDescription string `json:"jobDescription"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// formatter := 

	response := map[string]string{
		"formattedCV": 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}