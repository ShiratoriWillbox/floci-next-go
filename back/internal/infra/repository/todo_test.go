package repository

import (
	"testing"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTodoRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		seed      []dto.Todo
		id        uint
		wantErr   bool
		wantTitle string
	}{
		{
			name:      "found",
			seed:      []dto.Todo{{Title: "a", Completed: false}},
			id:        1,
			wantTitle: "a",
		},
		{
			name:    "not found",
			seed:    nil,
			id:      99,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			repo := NewTodoRepository(db)

			for i := range tt.seed {
				require.NoError(t, db.Create(&tt.seed[i]).Error)
			}

			got, err := repo.GetByID(tt.id)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantTitle, got.Title)
		})
	}
}

func TestTodoRepository_List_Create_Delete(t *testing.T) {
	db := openTestDB(t)
	repo := NewTodoRepository(db)

	row := &dto.Todo{Title: "task", Completed: true}
	require.NoError(t, repo.Create(row))
	assert.NotZero(t, row.ID)

	list, err := repo.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "task", list[0].Title)
	assert.True(t, list[0].Completed)

	n, err := repo.DeleteByID(row.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	list, err = repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestTodoRepository_Save(t *testing.T) {
	db := openTestDB(t)
	repo := NewTodoRepository(db)

	row := &dto.Todo{Title: "old", Completed: false}
	require.NoError(t, repo.Create(row))

	row.Title = "new"
	row.Completed = true
	require.NoError(t, repo.Save(row))

	got, err := repo.GetByID(row.ID)
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.True(t, got.Completed)
}
