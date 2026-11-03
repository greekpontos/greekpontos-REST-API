package storage

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"tasks-api/internal/models"
	"time"
)

// Кастомные типизированные ошибки для точного сопоставления со статусами HTTP
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTitle = errors.New("title cannot be empty or contain only whitespaces")
)

type Storage interface {
	List() ([]models.Task, error)
	Create(title string, done bool) (models.Task, error)
	Get(id int) (models.Task, error)
	Update(id int, title string, done bool) (models.Task, error)
	Delete(id int) error
}

type MemoryStorage struct {
	mu     sync.RWMutex
	tasks  map[int]models.Task
	nextID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (s *MemoryStorage) List() ([]models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		list = append(list, task)
	}

	// Замечание преподавателя: Гарантируем стабильный и предсказуемый порядок (сортировка по ID)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})

	return list, nil
}

func (s *MemoryStorage) Create(title string, done bool) (models.Task, error) {
	// Замечание преподавателя: Валидация с учетом пробельных символов через TrimSpace
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return models.Task{}, ErrInvalidTitle
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Замечание преподавателя: Сервер сам полностью контролирует время создания, игнорируя ввод клиента
	task := models.Task{
		ID:        s.nextID,
		Title:     trimmedTitle,
		Done:      done,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	s.tasks[task.ID] = task
	s.nextID++

	return task, nil
}

func (s *MemoryStorage) Get(id int) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return models.Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (s *MemoryStorage) Update(id int, title string, done bool) (models.Task, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return models.Task{}, ErrInvalidTitle
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existingTask, exists := s.tasks[id]
	if !exists {
		return models.Task{}, ErrTaskNotFound
	}

	// КРИТИЧЕСКОЕ ЗАМЕЧАНИЕ: Сохраняем исходную дату создания (existingTask.CreatedAt)
	// Обновляем только те поля, которые разрешено менять клиенту
	existingTask.Title = trimmedTitle
	existingTask.Done = done

	s.tasks[id] = existingTask
	return existingTask, nil
}

func (s *MemoryStorage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)
	return nil
}
