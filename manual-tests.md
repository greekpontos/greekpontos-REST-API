# Сценарии ручного тестирования (Инструкция по curl)

Ниже приведены примеры `curl`-команд для полной проверки всех обязательных и дополнительных эндпоинтов REST API. Сервер должен быть запущен на `http://localhost:8080`.

---

## 1. Эндпоинт: Создание задачи (`POST /tasks`)

### Успешный сценарий
* **Запрос:**
  ```bash
  curl -X POST http://localhost:8080/tasks \
    -H "Content-Type: application/json" \
    -d '{"title": "Купить молоко", "done": false}'
  ```
* **Ожидаемый ответ (`201 Created`):**
  ```json
  {
    "id": 1,
    "title": "Купить молоко",
    "done": false,
    "created_at": "2026-11-03T14:45:00Z"
  }
  ```

### Ошибочный сценарий №1: Пустой заголовок или строка из пробелов
* **Запрос:**
  ```bash
  curl -X POST http://localhost:8080/tasks \
    -H "Content-Type: application/json" \
    -d '{"title": "   ", "done": false}'
  ```
* **Ожидаемый ответ (`400 Bad Request`):**
  ```json
  {
    "error": "Title cannot be empty or whitespaces only"
  }
  ```

### Ошибочный сценарий №2: Некорректный Content-Type
* **Запрос:**
  ```bash
  curl -X POST http://localhost:8080/tasks \
    -H "Content-Type: text/plain" \
    -d '{"title": "Новая задача"}'
  ```
* **Ожидаемый ответ (`415 Unsupported Media Type` или `400 Bad Request`):**
  ```json
  {
    "error": "Content-Type must be application/json"
  }
  ```

---

## 2. Эндпоинт: Получение списка задач (`GET /tasks`)

### Успешный сценарий (После добавления задач)
* **Запрос:**
  ```bash
  curl -X GET http://localhost:8080/tasks
  ```
* **Ожидаемый ответ (`200 OK`):**
  ```json
  [
    {
      "id": 1,
      "title": "Купить молоко",
      "done": false,
      "created_at": "2026-11-03T14:45:00Z"
    }
  ]
  ```

---

## 3. Эндпоинт: Получение задачи по ID (`GET /tasks/{id}`)

### Успешный сценарий
* **Запрос:**
  ```bash
  curl -X GET http://localhost:8080/tasks/1
  ```
* **Ожидаемый ответ (`200 OK`):**
  ```json
  {
    "id": 1,
    "title": "Купить молоко",
    "done": false,
    "created_at": "2026-11-03T14:45:00Z"
  }
  ```

### Ошибочный сценарий №1: Задача не найдена
* **Запрос:**
  ```bash
  curl -X GET http://localhost:8080/tasks/999
  ```
* **Ожидаемый ответ (`404 Not Found`):**
  ```json
  {
    "error": "Task not found"
  }
  ```

### Ошибочный сценарий №2: Некорректный ID (не число или <= 0)
* **Запрос:**
  ```bash
  curl -X GET http://localhost:8080/tasks/abc
  ```
* **Ожидаемый ответ (`400 Bad Request`):**
  ```json
  {
    "error": "Invalid task ID"
  }
  ```

---

## 4. Эндпоинт: Обновление задачи (`PUT /tasks/{id}`)

### Успешный сценарий (Обновление с сохранением created_at)
* **Запрос:**
  ```bash
  curl -X PUT http://localhost:8080/tasks/1 \
    -H "Content-Type: application/json" \
    -d '{"title": "Купить свежее молоко", "done": true}'
  ```
* **Ожидаемый ответ (`200 OK`):**
  ```json
  {
    "id": 1,
    "title": "Купить свежее молоко",
    "done": true,
    "created_at": "2026-11-03T14:45:00Z"
  }
  ```

### Ошибочный сценарий: Обновление несуществующей задачи
* **Запрос:**
  ```bash
  curl -X PUT http://localhost:8080/tasks/999 \
    -H "Content-Type: application/json" \
    -d '{"title": "Тест", "done": true}'
  ```
* **Ожидаемый ответ (`404 Not Found`):**
  ```json
  {
    "error": "Task not found"
  }
  ```

---

## 5. Эндпоинт: Удаление задачи (`DELETE /tasks/{id}`)

### Успешный сценарий
* **Запрос:**
  ```bash
  curl -X DELETE http://localhost:8080/tasks/1
  ```
* **Ожидаемый ответ (`204 No Content`):**
  *(Тело ответа пустое)*

### Ошибочный сценарий: Повторное удаление (уже удалена)
* **Запрос:**
  ```bash
  curl -X DELETE http://localhost:8080/tasks/1
  ```
* **Ожидаемый ответ (`404 Not Found`):**
  ```json
  {
    "error": "Task not found"
  }
  ```

---

## 6. Диагностический эндпоинт (`GET /health`)

### Успешный сценарий
* **Запрос:**
  ```bash
  curl -X GET http://localhost:8080/health
  ```
* **Ожидаемый ответ (`200 OK`):**
  ```json
  {
    "status": "OK"
  }
  ```

### Ошибочный сценарий: Неподдерживаемый метод
* **Запрос:**
  ```bash
  curl -X POST http://localhost:8080/health
  ```
* **Ожидаемый ответ (`405 Method Not Allowed` c заголовком `Allow: GET`):**
  ```json
  {
    "error": "Method Not Allowed"
  }
  ```
