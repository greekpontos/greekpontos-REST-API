package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tasks-api/internal/storage"
)

type Handler struct {
	Store storage.Storage
}

func New(s storage.Storage) *Handler {
	return &Handler{Store: s}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		// Замечание: Логируем внутреннюю ошибку кодирования на сервере
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("ошибка сериализации JSON: %v", err)
		}
	}
}

// SetupRoutes настраивает маршрутизацию и обрабатывает неизвестные пути в формате JSON
func (h *Handler) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", h.TasksCollection)
	mux.HandleFunc("/tasks/", h.TaskItem)
	mux.HandleFunc("/health", h.HealthCheck)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/tasks" || path == "/health" || strings.HasPrefix(path, "/tasks/") {
			mux.ServeHTTP(w, r)
			return
		}
		// Замечание: Согласованный JSON-формат для неизвестных маршрутов
		respondWithError(w, http.StatusNotFound, "Resource not found")
	})
}

// TasksCollection обрабатывает /tasks (GET, POST)
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := h.Store.List()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		respondWithJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		// Замечание: Проверка заголовка Content-Type
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		// Замечание: Ограничение размера тела запроса (10 КБ)
		r.Body = http.MaxBytesReader(w, r.Body, 10240)

		// Замечание: Строгий декодер
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req TaskRequest
		if err := dec.Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON structure")
			return
		}

		// Проверка на отсутствие лишних данных после первого объекта
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			respondWithError(w, http.StatusBadRequest, "Request body must contain only a single JSON object")
			return
		}

		if req.Title == nil {
			respondWithError(w, http.StatusBadRequest, "Field 'title' is required")
			return
		}

		doneVal := false
		if req.Done != nil {
			doneVal = *req.Done
		}

		createdTask, err := h.Store.Create(*req.Title, doneVal)
		if err != nil {
			if errors.Is(err, storage.ErrInvalidTitle) {
				respondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		respondWithJSON(w, http.StatusCreated, createdTask)

	default:
		// Замечание: Установка заголовка Allow при ошибке 405
		w.Header().Set("Allow", "GET, POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// TaskItem обрабатывает /tasks/{id} (GET, PUT, DELETE)
func (h *Handler) TaskItem(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing task ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	// Замечание: Явное определение допустимого диапазона ID (> 0)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID. ID must be a positive integer")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := h.Store.Get(id)
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				respondWithError(w, http.StatusNotFound, "Task not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		respondWithJSON(w, http.StatusOK, task)

	case http.MethodPut:
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 10240)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req TaskRequest
		if err := dec.Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON structure")
			return
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			respondWithError(w, http.StatusBadRequest, "Request body must contain only a single JSON object")
			return
		}

		if req.Title == nil {
			respondWithError(w, http.StatusBadRequest, "Field 'title' is required")
			return
		}

		doneVal := false
		if req.Done != nil {
			doneVal = *req.Done
		}

		// Замечание: Дифференциация ошибок через errors.Is()
		updatedTask, err := h.Store.Update(id, *req.Title, doneVal)
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				respondWithError(w, http.StatusNotFound, "Task not found")
				return
			}
			if errors.Is(err, storage.ErrInvalidTitle) {
				respondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		respondWithJSON(w, http.StatusOK, updatedTask)

	case http.MethodDelete:
		err := h.Store.Delete(id)
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				respondWithError(w, http.StatusNotFound, "Task not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		respondWithJSON(w, http.StatusNoContent, nil)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HealthCheck обрабатывает эндпоинт /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "OK"})
}
