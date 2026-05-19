package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/floci-next-go/back/internal/infra/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type createTodoFileBody struct {
	Name string `json:"name"`
}

// CreateTodoFileUpload POST /todos/:id/files — ファイル行を作成し、アップロード用の presigned PUT URL を返す。
func (h *TodoHandler) CreateTodoFileUpload(c *gin.Context) {
	if h.S3Client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "object storage is not configured"})
		return
	}

	todoID, ok := parseID(c)
	if !ok {
		return
	}

	todo, err := h.Todos.GetByID(todoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var body createTodoFileBody
	_ = c.ShouldBindJSON(&body)
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "upload"
	}

	objectKey := fmt.Sprintf("todos/%d/%s", todo.ID, uuid.New().String())

	fileRow := dto.File{
		Name: name,
		Path: objectKey,
	}
	if err := h.Files.Create(&fileRow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	presigned, err := h.S3Client.PresignPutObject(c.Request.Context(), objectKey, storage.DefaultPutExpires)
	if err != nil {
		_ = h.Files.DeleteByID(fileRow.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"file_id":       fileRow.ID,
		"upload_url":    presigned.URL,
		"upload_method": presigned.Method,
		"headers":       presigned.SignedHeader,
	})
}

// AttachTodoFile PUT /todos/:id/files/:file_id — todo と file を todos_files で紐づける。
func (h *TodoHandler) AttachTodoFile(c *gin.Context) {
	todoID, ok := parseID(c)
	if !ok {
		return
	}
	fileID, ok := parseFileID(c)
	if !ok {
		return
	}

	todo, err := h.Todos.GetByID(todoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file, err := h.Files.GetByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	expectedPrefix := fmt.Sprintf("todos/%d/", todo.ID)
	if !strings.HasPrefix(file.Path, expectedPrefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file does not belong to this todo"})
		return
	}

	link, err := h.Files.FirstOrCreateTodoFileLink(todoID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

// GetTodoFileDownload GET /todos/:id/files/:file_id — 紐づいたオブジェクトの presigned GET URL を返す。
func (h *TodoHandler) GetTodoFileDownload(c *gin.Context) {
	if h.S3Client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "object storage is not configured"})
		return
	}

	todoID, ok := parseID(c)
	if !ok {
		return
	}
	fileID, ok := parseFileID(c)
	if !ok {
		return
	}

	linked, err := h.Files.TodoLinksFile(todoID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !linked {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not linked to this todo"})
		return
	}

	file, err := h.Files.GetByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	expectedPrefix := fmt.Sprintf("todos/%d/", todoID)
	if !strings.HasPrefix(file.Path, expectedPrefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file does not belong to this todo"})
		return
	}

	presigned, err := h.S3Client.PresignGetObject(c.Request.Context(), file.Path, storage.DefaultGetExpires)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url":    presigned.URL,
		"download_method": presigned.Method,
		"headers":         presigned.SignedHeader,
	})
}

func parseFileID(c *gin.Context) (uint, bool) {
	raw := c.Param("file_id")
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id"})
		return 0, false
	}
	return uint(n), true
}
