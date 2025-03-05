package api

import (
	"bytes"
	"context"

	// "encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"

	"github.com/Emmanuella-codes/burnished-microservice/internal/ai"
	// "github.com/Emmanuella-codes/burnished-microservice/internal/documents"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type ProcessCVRequest struct {
	Format				 			 string `json:"format" binding:"required,oneof=ats roast"`
	JobDescription 			 string `json:"jobDescription,omitempty"`
	GenerateCoverLetter  bool 	`json:"generateCoverLetter"`
}

type ProcessResponse struct {
	Success 		bool 		`json:"success"`
	Message 		string	`json:"message,omitempty"`
	FileURL 		string	`json:"fileUrl,omitempty"`
	CoverLetter string	`json:"coverLetter,omitempty"`
	Feedback 		string	`json:"feedback,omitempty"`
}

func (s *Server) saveToGCS(fileData []byte, filename string) (string, error) {
	ctx := context.Background()
	// create a GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("fsiled to create storage client: %v", err)
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

	// set cache control and content type
  w.CacheControl = "public, max-age=86400"

	// detect content type based on file extension
	contentType := "application/octet-stream"
	switch filepath.Ext(filename) {
	case ".pdf":
			contentType = "application/pdf"
	case ".docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
			contentType = "text/plain"
	}
	w.ContentType = contentType

	// write the data
	if _, err := w.Write(fileData); err != nil {
		return "", fmt.Errorf("failed to write to object: %v", err)
	}

	// close the writer
	if err := w.Close(); err != nil {
			return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set object ACL: %v", err)
	}

	// generate a public URL for the object
	fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filename)

	return fileURL, nil
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large"})
		return
	}

	// Read the file
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	ext := filepath.Ext(header.Filename)
	response := ProcessResponse{Success: true}

	fileReader := bytes.NewReader(fileData)

	switch req.Format {
	case "ats":
		processedFile, err := s.docProc.FormatForATS(fileReader, ext, req.JobDescription)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process CV"})
			return
		}

		// generate a unique filename
		filename := fmt.Sprintf("%s-%s%s", uuid.New().String(), "optimized", ext)
		
		// save to gcs
		fileUrl, err := s.saveToGCS(processedFile, filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save processed file"})
			return
	}
	
	response.FileURL = fileUrl

		// generate cover letter
		if req.GenerateCoverLetter && req.JobDescription != "" {
			coverLetter, err := ai.GenerateCoverLetter(fileData, req.JobDescription, s.cfg.GeminiAPIKey)
			if err == nil {
				response.CoverLetter = coverLetter

				//save the cover letter as a document
				// if coverLetter != "" {
				// 	clFilename := fmt.Sprintf("%s-%s.pdf", uuid.New().String(), "coverletter")
				// 	coverLetterDoc, err := documents.DocumentProcessor.CreateFormattedDocument(coverLetter)
				// }
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

func (s *Server) formatCVHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CV file provided"})
		return
	}
	defer file.Close()

	jobDesc := c.PostForm("jobDescription")
	ext := filepath.Ext(header.Filename)

	processedFile, err := s.docProc.FormatForATS(file, ext, jobDesc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format CV"})
		return
	}

	// set appropriate headers for file download
	c.Header("Content-Disposition", "attachment; filename=formatted_cv"+ext)
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
	default:
		return "application/octet-stream"
	}
}
