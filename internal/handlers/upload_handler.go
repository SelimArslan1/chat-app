package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SelimArslan1/chat-app/internal/storage"
)

type UploadHandler struct {
	Storage *storage.MinioClient
}

func NewUploadHandler(s *storage.MinioClient) *UploadHandler {
	return &UploadHandler{Storage: s}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > storage.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 10MB"})
		return
	}

	// Check file extension
	if !storage.IsAllowedExtension(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type, only jpg, png, gif, webp allowed"})
		return
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if !storage.IsAllowedImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type"})
		return
	}

	// Upload to MinIO
	url, err := h.Storage.Upload(file, header.Filename, header.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// GetFile serves a file from MinIO storage
func (h *UploadHandler) GetFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	reader, contentType, err := h.Storage.GetFile(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}
