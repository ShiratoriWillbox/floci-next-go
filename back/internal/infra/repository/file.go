package repository

import (
	"github.com/floci-next-go/back/internal/infra/dto"
	"gorm.io/gorm"
)

// FileRepository persists dto.File rows and todo–file links.
type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(row *dto.File) error {
	return r.db.Create(row).Error
}

func (r *FileRepository) GetByID(id uint) (*dto.File, error) {
	var row dto.File
	if err := r.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FileRepository) DeleteByID(id uint) error {
	return r.db.Delete(&dto.File{}, id).Error
}

// FirstOrCreateTodoFileLink ensures a row exists in todos_files for the pair (todoID, fileID).
func (r *FileRepository) FirstOrCreateTodoFileLink(todoID, fileID uint) (*dto.TodosFiles, error) {
	link := dto.TodosFiles{TodoID: todoID, FileID: fileID}
	if err := r.db.Where("todo_id = ? AND file_id = ?", todoID, fileID).FirstOrCreate(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// TodoLinksFile は todos_files に (todoID, fileID) の行があるかを返す。
func (r *FileRepository) TodoLinksFile(todoID, fileID uint) (bool, error) {
	var n int64
	err := r.db.Model(&dto.TodosFiles{}).Where("todo_id = ? AND file_id = ?", todoID, fileID).Count(&n).Error
	return n > 0, err
}
