package supabase

import (
	"bytes"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func UploadToSupabase(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	bucketName := os.Getenv("SUPABASE_BUCKET_NAME")
	filename := "raw/" + header.Filename

	storageClient := supabaseClient.Storage

	// upload
	_, err = storageClient.UploadFile(bucketName, filename, bytes.NewReader(buf.Bytes()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to Supabase"})
		return
	}

	// get public URL
	resp := storageClient.GetPublicUrl(bucketName, filename)
	publicUrl := resp.SignedURL

	c.JSON(http.StatusOK, gin.H{
		"message":    "File uploaded successfully",
		"public_url": publicUrl,
	})
}
