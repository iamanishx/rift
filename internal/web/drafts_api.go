package web

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/storage"
	"io"
	"net/http"
)

func UploadDraftImage(c *gin.Context) {
	did := c.Param("id")
	// parse uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open image"})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	client, err := storage.NewR2Client()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "R2 init failed"})
		return
	}
	key := "images/drafts/" + did + "/" + file.Filename
	if err := client.Upload(context.Background(), key, data, file.Header.Get("Content-Type")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}
	url := client.GetPublicURL(key)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
