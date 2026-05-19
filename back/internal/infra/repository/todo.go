package repository

import (
	"github.com/floci-next-go/back/internal/infra/dto"
	"gorm.io/gorm"
)

// TodoRepository persists dto.Todo rows.
type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) List() ([]dto.Todo, error) {
	var rows []dto.Todo
	err := r.db.Preload("Files").Order("id asc").Find(&rows).Error
	return rows, err
}

func (r *TodoRepository) GetByID(id uint) (*dto.Todo, error) {
	var row dto.Todo
	if err := r.db.Preload("Files").First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TodoRepository) Create(row *dto.Todo) error {
	return r.db.Create(row).Error
}

func (r *TodoRepository) Save(row *dto.Todo) error {
	return r.db.Save(row).Error
}

// DeleteByID removes a todo by primary key. RowsAffected is 0 when no row matched.
func (r *TodoRepository) DeleteByID(id uint) (rowsAffected int64, err error) {
	res := r.db.Delete(&dto.Todo{}, id)
	return res.RowsAffected, res.Error
}
