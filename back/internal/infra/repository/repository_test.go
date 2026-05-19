package repository

import (
	"testing"

	"github.com/floci-next-go/back/internal/infra/dto"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dto.Todo{}, &dto.File{}, &dto.TodosFiles{}))
	return db
}
