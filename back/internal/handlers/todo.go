package handlers

import (
	"net/http"
	"strconv"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/floci-next-go/back/internal/infra/repository"
	"github.com/floci-next-go/back/internal/infra/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TodoHandler struct {
	Todos    *repository.TodoRepository
	Files    *repository.FileRepository
	S3Client storage.IS3Client
}

type createTodoBody struct {
	Title     string `json:"title" binding:"required"`
	Completed *bool  `json:"completed"`
}

type updateTodoBody struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func (h *TodoHandler) List(c *gin.Context) {
	rows, err := h.Todos.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *TodoHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := h.Todos.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *TodoHandler) Create(c *gin.Context) {
	var body createTodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row := dto.Todo{Title: body.Title}
	if body.Completed != nil {
		row.Completed = *body.Completed
	}
	if err := h.Todos.Create(&row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *TodoHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body updateTodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.Todos.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if body.Title != nil {
		row.Title = *body.Title
	}
	if body.Completed != nil {
		row.Completed = *body.Completed
	}
	if err := h.Todos.Save(row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fresh, err := h.Todos.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fresh)
}

func (h *TodoHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	affected, err := h.Todos.DeleteByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(n), true
}
