package repository

import (
	"testing"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFileRepository_GetByID_notFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewFileRepository(db)

	got, err := repo.GetByID(42)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, got)
}

func TestFileRepository_FirstOrCreateTodoFileLink_idempotent(t *testing.T) {
	db := openTestDB(t)
	repo := NewFileRepository(db)
	require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "f", Path: "todos/1/k"}).Error)

	tests := []struct {
		name string
	}{
		{"first call"},
		{"second call same pair"},
	}

	var firstID uint
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := repo.FirstOrCreateTodoFileLink(1, 1)
			require.NoError(t, err)
			require.NotNil(t, link)
			assert.Equal(t, uint(1), link.TodoID)
			assert.Equal(t, uint(1), link.FileID)
			if i == 0 {
				firstID = link.ID
			} else {
				assert.Equal(t, firstID, link.ID, "same row on repeat")
			}
		})
	}
}

func TestFileRepository_Create_DeleteByID(t *testing.T) {
	db := openTestDB(t)
	repo := NewFileRepository(db)

	f := &dto.File{Name: "doc", Path: "todos/1/x"}
	require.NoError(t, repo.Create(f))
	assert.NotZero(t, f.ID)

	got, err := repo.GetByID(f.ID)
	require.NoError(t, err)
	assert.Equal(t, "doc", got.Name)

	require.NoError(t, repo.DeleteByID(f.ID))

	_, err = repo.GetByID(f.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFileRepository_TodoLinksFile(t *testing.T) {
	db := openTestDB(t)
	repo := NewFileRepository(db)
	require.NoError(t, db.Create(&dto.Todo{Title: "t"}).Error)
	require.NoError(t, db.Create(&dto.File{Name: "f", Path: "todos/1/x"}).Error)

	ok, err := repo.TodoLinksFile(1, 1)
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = repo.FirstOrCreateTodoFileLink(1, 1)
	require.NoError(t, err)

	ok, err = repo.TodoLinksFile(1, 1)
	require.NoError(t, err)
	assert.True(t, ok)
}
