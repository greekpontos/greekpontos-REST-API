package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at,omitempty"`
}

func Test_Tasks_API_Workflow(t *testing.T) {
	// --- ТЕСТ 1: Создание задачи (POST /tasks) - УСПЕХ ---
	t.Run("POST /tasks - Success", func(t *testing.T) {
		payload := []byte(`{"title":"Изучить SOLID в Go","done":false}`)
		resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Ожидался статус 201, получен: %d", resp.StatusCode)
		}

		var createdTask Task
		json.NewDecoder(resp.Body).Decode(&createdTask)
		t.Logf("УСПЕХ: Задача создана! ID: %d, Title: %s", createdTask.ID, createdTask.Title)
	})

	// --- ТЕСТ 2: Создание задачи (POST /tasks) - ОШИБКА (пустой title) ---
	t.Run("POST /tasks - Failure (Empty Title)", func(t *testing.T) {
		payload := []byte(`{"title":"","done":false}`)
		resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400, получен: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		t.Logf("УСПЕХ: Получена ожидаемая ошибка от сервера: %s", string(body))
	})

	// --- ТЕСТ 3: Получение списка задач (GET /tasks) - УСПЕХ ---
	t.Run("GET /tasks - Success", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/tasks")
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var tasks []Task
		json.NewDecoder(resp.Body).Decode(&tasks)
		t.Logf("УСПЕХ: Получен список задач. Количество: %d", len(tasks))
	})

	// --- ТЕСТ 4: Получение задачи по ID (GET /tasks/{id}) - УСПЕХ ---
	t.Run("GET /tasks/1 - Success", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/tasks/1")
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var task Task
		json.NewDecoder(resp.Body).Decode(&task)
		t.Logf("УСПЕХ: Задача №1 найдена! Title: %s, Done: %v", task.Title, task.Done)
	})

	// --- ТЕСТ 5: Получение по ID (GET /tasks/{id}) - ОШИБКА (404 Not Found) ---
	t.Run("GET /tasks/999 - Failure (Not Found)", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/tasks/999")
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Ожидался статус 404, получен: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		t.Logf("УСПЕХ: Получена ожидаемая ошибка 404 от сервера: %s", string(body))
	})

	// --- ТЕСТ 6: Обновление задачи (PUT /tasks/{id}) - УСПЕХ ---
	t.Run("PUT /tasks/1 - Success", func(t *testing.T) {
		payload := []byte(`{"title":"Изучить SOLID и паттерны в Go","done":true}`)
		req, _ := http.NewRequest(http.MethodPut, baseURL+"/tasks/1", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var updatedTask Task
		json.NewDecoder(resp.Body).Decode(&updatedTask)
		t.Logf("УСПЕХ: Задача №1 обновлена! Новый Title: %s, Статус Done: %v", updatedTask.Title, updatedTask.Done)
	})

	// --- ТЕСТ 7: Проверка эндпоинта /health ---
	t.Run("GET /health - Success", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("УСПЕХ: Сервер здоров! Статус: %s", string(body))
	})

	// --- ТЕСТ 8: Удаление задачи (DELETE /tasks/{id}) - УСПЕХ ---
	t.Run("DELETE /tasks/1 - Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/tasks/1", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Ожидался статус 204, получен: %d", resp.StatusCode)
		}
		t.Log("УСПЕХ: Задача №1 успешно удалена с сервера!")
	})
}
