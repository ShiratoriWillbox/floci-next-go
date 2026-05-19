package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/floci-next-go/back/internal/infra/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dto.Todo{}, &dto.File{}, &dto.TodosFiles{}))
	return db
}

func newTestRouter(h *TodoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.GET("/todos", h.List)
	api.GET("/todos/:id", h.Get)
	api.POST("/todos", h.Create)
	api.PUT("/todos/:id", h.Update)
	api.DELETE("/todos/:id", h.Delete)
	api.POST("/todos/:id/files", h.CreateTodoFileUpload)
	api.PUT("/todos/:id/files/:file_id", h.AttachTodoFile)
	api.GET("/todos/:id/files/:file_id", h.GetTodoFileDownload)
	return r
}

func TestTodoHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid", body: `{"title":"hello"}`, wantStatus: http.StatusCreated},
		{name: "missing title", body: `{}`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openHandlerTestDB(t)
			h := &TodoHandler{
				Todos: repository.NewTodoRepository(db),
				Files: repository.NewFileRepository(db),
			}
			r := newTestRouter(h)

			req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusCreated {
				var got dto.Todo
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
				assert.Equal(t, "hello", got.Title)
			}
		})
	}
}

func TestTodoHandler_Get(t *testing.T) {
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "x", Completed: false}).Error)

	h := &TodoHandler{
		Todos: repository.NewTodoRepository(db),
		Files: repository.NewFileRepository(db),
	}
	r := newTestRouter(h)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "ok", path: "/api/todos/1", wantStatus: http.StatusOK},
		{name: "invalid id", path: "/api/todos/0", wantStatus: http.StatusBadRequest},
		{name: "not found", path: "/api/todos/99", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestTodoHandler_List(t *testing.T) {
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "a"}).Error)

	h := &TodoHandler{
		Todos: repository.NewTodoRepository(db),
		Files: repository.NewFileRepository(db),
	}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var rows []dto.Todo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "a", rows[0].Title)
}

func TestTodoHandler_List_withFiles(t *testing.T) {
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "with-files"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "readme", Path: "todos/1/k1"}).Error)
	require.NoError(t, db.Create(&dto.TodosFiles{TodoID: 1, FileID: 1}).Error)

	h := &TodoHandler{
		Todos: repository.NewTodoRepository(db),
		Files: repository.NewFileRepository(db),
	}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var rows []dto.Todo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rows))
	require.Len(t, rows, 1)
	require.Len(t, rows[0].Files, 1)
	assert.Equal(t, "readme", rows[0].Files[0].Name)
	assert.Equal(t, uint(1), rows[0].Files[0].ID)
}

func TestTodoHandler_Delete(t *testing.T) {
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "z"}).Error)

	h := &TodoHandler{
		Todos: repository.NewTodoRepository(db),
		Files: repository.NewFileRepository(db),
	}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/todos/1", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestTodoHandler_Update(t *testing.T) {
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "old", Completed: false}).Error)

	h := &TodoHandler{
		Todos: repository.NewTodoRepository(db),
		Files: repository.NewFileRepository(db),
	}
	r := newTestRouter(h)

	body := `{"title":"new","completed":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/todos/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got dto.Todo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "new", got.Title)
	assert.True(t, got.Completed)
}
