package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tasks-api/internal/models"
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

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

// TasksCollection обрабатывает /tasks (GET, POST)
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks := h.Store.List()
		respondWithJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		if t.Title == "" {
			respondWithError(w, http.StatusBadRequest, "Field 'title' is required")
			return
		}

		createdTask, err := h.Store.Create(t)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusCreated, createdTask)

	default:
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
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, exists := h.Store.Get(id)
		if !exists {
			respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		respondWithJSON(w, http.StatusOK, task)

	case http.MethodPut:
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		if t.Title == "" {
			respondWithError(w, http.StatusBadRequest, "Field 'title' is required")
			return
		}

		updatedTask, err := h.Store.Update(id, t)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, updatedTask)

	case http.MethodDelete:
		err := h.Store.Delete(id)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HealthCheck обрабатывает PRO-эндпоинт /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "OK"})
}
