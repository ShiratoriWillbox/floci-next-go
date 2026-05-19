package dto

import "time"

type TodosFiles struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TodoID    uint      `json:"todo_id" gorm:"not null;uniqueIndex:idx_todos_files_pair"`
	FileID    uint      `json:"file_id" gorm:"not null;uniqueIndex:idx_todos_files_pair"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Todo      Todo      `json:"todo" gorm:"foreignKey:TodoID"`
	File      File      `json:"file" gorm:"foreignKey:FileID"`
}
