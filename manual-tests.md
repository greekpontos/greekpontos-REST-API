# Результаты ручного тестирования

## Создание задачи (POST /tasks)
```bash
curl -X POST -H "Content-Type: application/json" -d "{\"title\":\"Изучить паттерны в Go\",\"done\":false}" http://localhost:8080/tasks
```

## Получение списка задач (GET /tasks)
```bash
curl -X GET http://localhost:8080/tasks
```
