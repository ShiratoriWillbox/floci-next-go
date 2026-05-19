package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/floci-next-go/back/internal/infra/repository"
	"github.com/floci-next-go/back/internal/infra/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestTodoHandler_CreateTodoFileUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name          string
		s3            storage.IS3Client
		todoSeed      bool
		body          string
		wantStatus    int
		checkFileGone bool // presign fails: file row must be rolled back
	}{
		{
			name:       "s3 not configured",
			s3:         nil,
			todoSeed:   true,
			body:       `{}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "todo not found",
			s3: (func() storage.IS3Client {
				mock := storage.NewMockIS3Client(ctrl)
				mock.EXPECT().PresignPutObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()
				return mock
			})(),
			todoSeed:   false,
			body:       `{}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "ok with default name",
			s3: (func() storage.IS3Client {
				mock := storage.NewMockIS3Client(ctrl)
				mock.EXPECT().PresignPutObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storage.PresignedPutRequest{URL: "https://example.test/presigned", Method: "PUT"}, nil).AnyTimes()
				return mock
			})(),
			todoSeed:   true,
			body:       `{}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "presign error rolls back file",
			s3: (func() storage.IS3Client {
				mock := storage.NewMockIS3Client(ctrl)
				mock.EXPECT().PresignPutObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("presign failed")).AnyTimes()
				return mock
			})(),
			todoSeed:      true,
			body:          `{}`,
			wantStatus:    http.StatusInternalServerError,
			checkFileGone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openHandlerTestDB(t)
			if tt.todoSeed {
				require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
			}
			h := &TodoHandler{
				Todos:    repository.NewTodoRepository(db),
				Files:    repository.NewFileRepository(db),
				S3Client: tt.s3,
			}
			r := newTestRouter(h)

			req := httptest.NewRequest(http.MethodPost, "/api/todos/1/files", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusCreated {
				var resp struct {
					FileID       uint   `json:"file_id"`
					UploadURL    string `json:"upload_url"`
					UploadMethod string `json:"upload_method"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotZero(t, resp.FileID)
				assert.Equal(t, "PUT", resp.UploadMethod)
				assert.Contains(t, resp.UploadURL, "example.test")
			}
			if tt.checkFileGone {
				var n int64
				require.NoError(t, db.Model(&dto.File{}).Count(&n).Error)
				assert.Zero(t, n)
			}
		})
	}
}

func TestTodoHandler_AttachTodoFile(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		seedTodo   bool
		filePath   string // path for file id 1 when seedTodo && filePath != ""
		wantStatus int
	}{
		{name: "invalid file_id", path: "/api/todos/1/files/0", seedTodo: true, filePath: "todos/1/k", wantStatus: http.StatusBadRequest},
		{name: "todo not found", path: "/api/todos/2/files/1", seedTodo: false, filePath: "todos/1/k", wantStatus: http.StatusNotFound},
		{name: "file not found", path: "/api/todos/1/files/9", seedTodo: true, filePath: "todos/1/k", wantStatus: http.StatusNotFound},
		{name: "path mismatch", path: "/api/todos/1/files/1", seedTodo: true, filePath: "todos/999/k", wantStatus: http.StatusBadRequest},
		{name: "ok", path: "/api/todos/1/files/1", seedTodo: true, filePath: "todos/1/k", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openHandlerTestDB(t)
			if tt.seedTodo {
				require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
			}
			if tt.filePath != "" {
				require.NoError(t, db.Create(&dto.File{Name: "f", Path: tt.filePath}).Error)
			}

			h := &TodoHandler{
				Todos: repository.NewTodoRepository(db),
				Files: repository.NewFileRepository(db),
			}
			r := newTestRouter(h)

			req := httptest.NewRequest(http.MethodPut, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var link dto.TodosFiles
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &link))
				assert.Equal(t, uint(1), link.TodoID)
				assert.Equal(t, uint(1), link.FileID)
			}
		})
	}
}

func TestTodoHandler_GetTodoFileDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := storage.NewMockIS3Client(ctrl)
	mock.EXPECT().PresignGetObject(gomock.Any(), "todos/1/the-key", gomock.Any()).Return(
		&storage.PresignedGetRequest{URL: "https://example.test/get", Method: "GET"},
		nil,
	)

	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "f", Path: "todos/1/the-key"}).Error)
	require.NoError(t, db.Create(&dto.TodosFiles{TodoID: 1, FileID: 1}).Error)

	h := &TodoHandler{
		Todos:    repository.NewTodoRepository(db),
		Files:    repository.NewFileRepository(db),
		S3Client: mock,
	}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/todos/1/files/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		DownloadURL    string `json:"download_url"`
		DownloadMethod string `json:"download_method"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "GET", resp.DownloadMethod)
	assert.Contains(t, resp.DownloadURL, "example.test")
}

func TestTodoHandler_GetTodoFileDownload_notLinked(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := storage.NewMockIS3Client(ctrl)

	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "f", Path: "todos/1/the-key"}).Error)

	h := &TodoHandler{
		Todos:    repository.NewTodoRepository(db),
		Files:    repository.NewFileRepository(db),
		S3Client: mock,
	}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/todos/1/files/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestParseFileID_indirect(t *testing.T) {
	// exercise parseFileID via handler when path is malformed
	db := openHandlerTestDB(t)
	require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "f", Path: "todos/1/x"}).Error)

	h := &TodoHandler{Todos: repository.NewTodoRepository(db), Files: repository.NewFileRepository(db)}
	r := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/todos/1/files/%s", "abc"), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
