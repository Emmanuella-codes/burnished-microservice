package api

import (
	"bytes"
	"context"
	"log"
	"os"

	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"

	"github.com/Emmanuella-codes/burnished-microservice/internal/ai"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type ProcessCVRequest struct {
	Format				 			 string `json:"format" binding:"required,oneof=ats roast"`
	JobDescription 			 string `json:"jobDescription,omitempty"`
	GenerateCoverLetter  bool 	`json:"generateCoverLetter"`
}

type ProcessResponse struct {
	DocumentID  string  `json:"documentID"`
	Success 		bool 		`json:"success"`
	// Message 		string	`json:"message,omitempty"`
	// FileURL 		string	`json:"fileUrl,omitempty"`
	// CoverLetter string	`json:"coverLetter,omitempty"`
	FormattedFile   string `json:"formattedFile,omitempty"`
  CoverLetterFile string `json:"coverLetter,omitempty"`
  Feedback        string `json:"feedback,omitempty"`
  Error           string `json:"error,omitempty"`
}

func (s *Server) saveToGCS(fileData []byte, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()
	// create a GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()
	
	// replace with bucket name
	bucketName := s.cfg.StorageBucket

	// create a bucket handle
	bucket := client.Bucket(bucketName)

	// create an object handle
	obj := bucket.Object(filename)

	// create a writer
	w := obj.NewWriter(ctx)

	w.ContentType = getContentType(filepath.Ext(filename))
	w.CacheControl = "public, max-age=86400"

	if _, err := w.Write(fileData); err != nil {
			return "", fmt.Errorf("writing to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
			return "", fmt.Errorf("closing GCS writer: %w", err)
	}

	// make the object publicly accessible.
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("setting GCS ACL: %w", err)
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filename), nil
}

func (s *Server) healthHandler(c *gin.Context) {
	response := gin.H{
		"status": "ok",
		"time":		time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) processCVHandler(c *gin.Context) {
	var req ProcessCVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
	}

	file, header, err := c.Request.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CV file provided"})
		return
	}
	defer file.Close()

	// check the file size
	if header.Size > s.cfg.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File size exceeds limit: %d bytes", s.cfg.MaxFileSize)})
		return
	}

	// validate file type
	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type. Only PDF and DOCX are allowed"})
		return
	}

	// read the file
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	response := ProcessResponse{Success: true}

	fileReader := bytes.NewReader(fileData)

	switch req.Format {
	case "ats":
		if req.JobDescription == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Job description required for ATS formatting"})
			return
		}
		processedFile, err := s.docProc.FormatForATS(fileReader, ext, req.JobDescription)
		if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process CV: " + err.Error()})
				return
		}

		filename := fmt.Sprintf("cv_%s_optimized%s", uuid.New().String(), ext)
		fileURL, err := s.saveToGCS(processedFile, filename)
		if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save processed file: " + err.Error()})
				return
		}
		response.FileURL = fileURL

		if req.GenerateCoverLetter {
				coverLetter, err := ai.GenerateCoverLetter(fileData, req.JobDescription, s.cfg.GeminiAPIKey)
				if err != nil {
						log.Printf("Failed to generate cover letter: %v", err)
				} else {
						response.CoverLetter = coverLetter
				}
		}
	case "roast":
		feedback, err := s.docProc.RoastCV(fileReader, ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to roast CV"})
			return
		}
		response.Feedback = feedback
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) processCVHandler(c *gin.Context) {
	//authenticate request
	auth := c.GetHeader("Authorization")
	expectedAuth := "Bearer " + os.Getenv("BURNISHED_WEB_API_KEY")
	if auth != expectedAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	//parse multipart form
	if err := c.Request.ParseMultipartForm(s.cfg.MaxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
    return
	}

	documentID := c.PostForm("documentID")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "documentId is required"})
		return
	}

	mode := c.PostForm("mode")
	if mode != "roast" && mode != "format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'roast' or 'format'"})
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
	response := ProcessResponse{
		DocumentID: documentID,
		Success:    true,
	}

	// process based on mode
	fileReader := bytes.NewReader(fileData)
	switch mode {
	case "format":
		processedFile, err := s.docProc.FormatForATS(fileReader, ext, jobDescription)
		if err != nil {
			response.Success = false
			response.Error = "Failed to format CV: " + err.Error()
			s.s
		}
	}
}

func (s *Server) formatCVHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CV file provided"})
		return
	}
	defer file.Close()

	jobDesc := c.PostForm("jobDescription")
	if jobDesc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job description is required"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type; only PDF and DOCX are allowed"})
		return
	}

	processedFile, err := s.docProc.FormatForATS(file, ext, jobDesc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format CV"})
		return
	}

	// set appropriate headers for file download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=formatted_cv%s", ext))
	c.Data(http.StatusOK, getContentType(ext), processedFile)
}

func (s *Server) roastCVHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CV file provided"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	feedback, err := s.docProc.RoastCV(file, ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to roast CV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func (s *Server) generateCoverLetterHandler(c *gin.Context) {
	file, _, err := c.Request.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CV file provided"})
		return
	}
	defer file.Close()

	jobDesc := c.PostForm("jobDescription")
	if jobDesc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job description is required"})
		return
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	coverLetter, err := ai.GenerateCoverLetter(fileData, jobDesc, s.cfg.GeminiAPIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cover letter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coverLetter": coverLetter})
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
