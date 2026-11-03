package models

// Task описывает доменную модель задачи.
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"` // Обязательное поле
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"` // Сервер всегда возвращает дату создания
}
