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
			if err := s.sendWebhook(response); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}

		filename := fmt.Sprintf("cv_%s_formatted%s", uuid.New().String(), ".pdf")
		fileURL, err := s.saveToGCS(processedFile, filename)
		if err != nil {
			response.Success = false
			response.Error = "Failed to save formatted file: " + err.Error()
			if err := s.sendWebhook(response); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}
		response.FormattedFile = fileURL

		// generate cover letter
		coverLetter, err := ai.GenerateCoverLetter(fileData, jobDescription, s.cfg.GeminiAPIKey)
		if err != nil {
			log.Printf("Failed to generate cover letter: %v", err)
		} else {
			coverFilename := fmt.Sprintf("cover_letter_%s.txt", uuid.New().String())
			coverURL, err := s.saveToGCS([]byte(coverLetter), coverFilename)
			if err != nil {
				log.Printf("Failed to save cover letter: %v", err)
			} else {
				response.CoverLetterFile = coverURL
			}
		}

	case "roast":
		feedback, err := s.docProc.RoastCV(fileReader, ext)
		if err != nil {
			response.Success = false
			response.Error = "Failed to roast CV: " + err.Error()
			if err := s.sendWebhook(response); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
			return
		}
		response.Feedback = feedback
	}

	// send webhook for successful processing
	if err := s.sendWebhook(response); err != nil {
		log.Printf("Failed to send webhook: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Processing started",
		"documentID": documentID,
	})
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
