package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"tasks-api/internal/handlers"
	"tasks-api/internal/models"
	"tasks-api/internal/storage"
)

func Test_Tasks_API_Workflow(t *testing.T) {
	// Замечание преподавателя: Создаем изолированное хранилище и тестовый сервер.
	// Теперь тесты не зависят от внешне запущенного сервера и не засоряют основную базу данных.
	store := storage.NewMemoryStorage()
	h := handlers.New(store)
	server := httptest.NewServer(h.SetupRoutes())
	defer server.Close()

	var createdTaskID int

	// --- ТЕСТ 1: Создание задачи (POST /tasks) - УСПЕХ ---
	t.Run("POST /tasks - Success", func(t *testing.T) {
		payload := []byte(`{"title":"Изучить SOLID в Go","done":false}`)
		req, err := http.NewRequest(http.MethodPost, server.URL+"/tasks", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("Не удалось создать запрос: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Не удалось отправить запрос: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Ожидался статус 201, получен: %d", resp.StatusCode)
		}

		var createdTask models.Task
		if err := json.NewDecoder(resp.Body).Decode(&createdTask); err != nil {
			t.Fatalf("Не удалось декодировать ответ: %v", err)
		}

		// Замечание преподавателя: Глубокая проверка полей ответа
		if createdTask.ID <= 0 {
			t.Errorf("Ожидался валидный положительный ID, получен: %d", createdTask.ID)
		}
		if createdTask.Title != "Изучить SOLID в Go" {
			t.Errorf("Ожидался Title 'Изучить SOLID в Go', получен: '%s'", createdTask.Title)
		}
		if createdTask.CreatedAt == "" {
			t.Error("Поле created_at пустое, ожидалась валидная дата RFC3339")
		}

		// Сохраняем динамический ID для последующих тестов, убирая жесткую привязку к ID=1
		createdTaskID = createdTask.ID
	})

	// --- ТЕСТ 2: Валидация пробелов и пустого Title ---
	t.Run("POST /tasks - Failure (Whitespaces)", func(t *testing.T) {
		payload := []byte(`{"title":"    ","done":false}`)
		resp, err := http.Post(server.URL+"/tasks", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400 для строки из пробелов (TrimSpace), получен: %d", resp.StatusCode)
		}
	})

	// --- ТЕСТ 3: Проверка строгого JSON (Неизвестные поля) ---
	t.Run("POST /tasks - Failure (Unknown Fields)", func(t *testing.T) {
		payload := []byte(`{"title":"Валидная задача","unknown_field_test": true}`)
		resp, err := http.Post(server.URL+"/tasks", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400 из-за DisallowUnknownFields, получен: %d", resp.StatusCode)
		}
	})

	// --- ТЕСТ 4: Получение списка задач (GET /tasks) ---
	t.Run("GET /tasks - Success", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/tasks")
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var tasks []models.Task
		if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
			t.Fatalf("Ошибка декодирования списка: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("Ожидался размер списка 1, получен: %d", len(tasks))
		}
	})

	// --- ТЕСТ 5: Получение по динамическому ID ---
	t.Run("GET /tasks/{id} - Success", func(t *testing.T) {
		idStr := strconv.Itoa(createdTaskID)
		resp, err := http.Get(server.URL + "/tasks/" + idStr)
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var task models.Task
		if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
			t.Fatalf("Ошибка декодирования: %v", err)
		}

		if task.ID != createdTaskID {
			t.Errorf("ID в ответе (%d) не совпадает с динамически созданным (%d)", task.ID, createdTaskID)
		}
	})

	// --- ТЕСТ 6: Негативный тест на невалидный диапазон ID ---
	t.Run("GET /tasks/-5 - Failure (Invalid ID)", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/tasks/-5")
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400 для отрицательного ID, получен: %d", resp.StatusCode)
		}
	})

	// --- ТЕСТ 7: Обновление (PUT) с проверкой сохранения даты создания ---
	t.Run("PUT /tasks/{id} - Success (Keep CreatedAt)", func(t *testing.T) {
		idStr := strconv.Itoa(createdTaskID)

		// Сначала получаем оригинальный объект, чтобы зафиксировать исходное время
		respGet, _ := http.Get(server.URL + "/tasks/" + idStr)
		var original models.Task
		_ = json.NewDecoder(respGet.Body).Decode(&original)
		respGet.Body.Close()

		payload := []byte(`{"title":"Новое название задачи","done":true}`)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/tasks/"+idStr, bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
		}

		var updated models.Task
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf("Ошибка декодирования: %v", err)
		}

		// Проверяем, что критический баг устранен и дата осталась прежней
		if updated.CreatedAt != original.CreatedAt {
			t.Errorf("ОШИБКА: Дата создания затерлась или изменилась! Было: %s, стало: %s", original.CreatedAt, updated.CreatedAt)
		}
	})

	// --- ТЕСТ 8: Тестирование эндпоинта /health ---
	t.Run("GET /health - Full Verification", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200 для health, получен: %d", resp.StatusCode)
		}

		var bodyMap map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&bodyMap); err != nil {
			t.Fatalf("Не удалось распарсить JSON: %v", err)
		}

		if bodyMap["status"] != "OK" {
			t.Errorf("Ожидалось {'status':'OK'}, получено: %v", bodyMap)
		}
	})

	// --- ТЕСТ 9: Удаление и повторное удаление (Проверка 404) ---
	t.Run("DELETE /tasks/{id} - Success & Re-delete Failure", func(t *testing.T) {
		idStr := strconv.Itoa(createdTaskID)
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/tasks/"+idStr, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Запрос не ушел: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Ожидался статус 204 при удалении, получен: %d", resp.StatusCode)
		}

		// Замечание преподавателя: Негативный тест на повторное удаление уже удаленного ресурса
		req2, _ := http.NewRequest(http.MethodDelete, server.URL+"/tasks/"+idStr, nil)
		resp2, _ := http.DefaultClient.Do(req2)
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusNotFound {
			t.Errorf("Ожидался статус 404 при повторном удалении ресурса, получен: %d", resp2.StatusCode)
		}
	})
}

// --- ТЕСТ 10: Проверка потокобезопасности и отсутствия состояний гонки (Data Races) ---
func Test_Tasks_API_Concurrency(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := handlers.New(store)
	server := httptest.NewServer(h.SetupRoutes())
	defer server.Close()

	var wg sync.WaitGroup
	workers := 30

	for i := 0; i < workers; i++ {
		wg.Add(2)

		// Конкурентное создание задач (Писатели)
		go func(val int) {
			defer wg.Done()
			payload := []byte(`{"title":"Конкурентная задача","done":false}`)
			resp, err := http.Post(server.URL+"/tasks", "application/json", bytes.NewBuffer(payload))
			if err == nil {
				resp.Body.Close()
			}
		}(i)

		// Конкурентное чтение списка (Читатели)
		go func() {
			defer wg.Done()
			resp, err := http.Get(server.URL + "/tasks")
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
}
