package dto

import "time"

type Todo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"size:512;not null"`
	Completed bool      `json:"completed" gorm:"default:false;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Files     []File    `json:"files,omitempty" gorm:"many2many:todos_files"`
}
