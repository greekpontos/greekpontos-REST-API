package models

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"` // Обязательное поле
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at,omitempty"`
}
