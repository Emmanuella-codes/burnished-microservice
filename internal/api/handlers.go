package api

import (
	"bytes"
	// "encoding/json"
	"log"
	"os"
	"strings"

	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Emmanuella-codes/burnished-microservice/internal/ai"
	"github.com/Emmanuella-codes/burnished-microservice/internal/dtos"

	"github.com/gin-gonic/gin"
)

type ProcessingStatus string

const (
	StatusCompleted ProcessingStatus = "completed"
	StatusFailed    ProcessingStatus = "failed"
)

type ProcessCVRequest struct {
	File          	[]byte  `form:"file"`
  Filename      	string  `form:"filename"`
	Mode				 		string  `json:"mode" binding:"required,oneof=roast format"`
	JobDescription 	string  `json:"jobDescription"`
	// GenerateCoverLetter bool 	 `json:"generateCoverLetter"`
}

type ProcessResponse struct {
	// DocumentID  			string  					`json:"documentID"`
	Status 						ProcessingStatus 	`json:"status"`
	FormattedResume   *dtos.Resume 			`json:"formattedResume,omitempty"`
  CoverLetter 			string 						`json:"coverLetter,omitempty"`
  Feedback        	string 						`json:"feedback,omitempty"`
  Error           	string 						`json:"error,omitempty"`
}

func (s *Server) serveFileHandler(c *gin.Context) {
	filename := filepath.Clean(c.Param("filename"))

	// security check
	if strings.Contains(filename, "..") || filepath.IsAbs(filename) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	uploadDir := s.cfg.UploadDir
	filePath := filepath.Join(uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	contentType := getContentType(filepath.Ext(filename))

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
	c.File(filePath)
}

func (s *Server) healthHandler(c *gin.Context) {
	response := gin.H{
		"status": "ok",
		"time":		time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) processCVHandler(c *gin.Context) {
	// authenticate request
	auth := c.GetHeader("Authorization")
	expectedAuth := "Bearer " + os.Getenv("BURNISHED_WEB_API_KEY")
	log.Printf("Received auth: %s", auth)
	log.Printf("Expected auth: %s", expectedAuth)
	if auth != expectedAuth {
		log.Printf("Auth failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// parse multipart form
	if err := c.Request.ParseMultipartForm(s.cfg.MaxFileSize); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
    return
	}

	mode := c.PostForm("mode")
	log.Printf("Received mode: %s", mode)
	if mode != "roast" && mode != "format" && mode != "letter" {
		log.Printf("Invalid mode: %s", mode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'roast' or 'format' or 'letter'"})
		return
	}

	jobDescription := c.PostForm("jobDescription")
	if mode == "format" && jobDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobDescription is required for format mode"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// validate file type
	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type; only PDF and DOCX are allowed"})
		return
	}

	// check file size
	if header.Size > s.cfg.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File size exceeds limit: %d bytes", s.cfg.MaxFileSize),
		})
		return
	}

	// read file
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// prepare webhook response
	var response ProcessResponse
	response.Status = StatusCompleted

	log.Println("starting ParseCV...")

	// process based on mode
	switch mode {
	case "format":
		fileReader := bytes.NewReader(fileData)
		resume, err := s.docProc.FormatForATS(fileReader, ext, jobDescription)
		if err != nil {
			response.Status = StatusFailed
			response.Error = "Failed to format CV: " + err.Error()
			log.Printf("🟢 About to send JSON: %+v", response)
			if err := s.sendWebhook(response); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}

		response.FormattedResume = resume

		if err := s.sendWebhook(response); err != nil {
			log.Printf("Failed to send webhook: %v", err)
		}

	case "roast":
		fileReader := bytes.NewReader(fileData)
		feedback, err := s.docProc.RoastCV(fileReader, ext)
		if err != nil {
			response.Status = StatusFailed
			response.Error = "Failed to roast CV: " + err.Error()
			if err := s.sendWebhook(response); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}
		response.Feedback = feedback

		if err := s.sendWebhook(response); err != nil {
			log.Printf("Failed to send webhook: %v", err)
		}
	
	case "letter":
		coverLetter, err := ai.GenerateCoverLetter(fileData, jobDescription, s.cfg.GeminiAPIKey)
		if err != nil {
			response.Status = StatusFailed
			response.Error = fmt.Sprintf("Failed to generate cover letter: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}

		response.CoverLetter = coverLetter

		if err := s.sendWebhook(response); err != nil {
			log.Printf("Failed to send webhook: %v", err)
		}
	}

	log.Printf("Sending response for mode: %s", mode)
	c.JSON(http.StatusOK, response)
	log.Printf("Response sent successfully")
}


func getContentType(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
    return "text/plain"
	default:
		return "application/octet-stream"
	}
}
