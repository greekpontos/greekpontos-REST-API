package main

import (
	"log"
	"net/http"

	"tasks-api/internal/handlers"
	customHttp "tasks-api/internal/http"
	"tasks-api/internal/storage"
)

func main() {
	// Инициализируем потокобезопасное хранилище в памяти
	store := storage.NewMemoryStorage()

	// Создаем структуру обработчиков
	h := handlers.New(store)

	// Получаем настроенный роутер с JSON-обработкой ошибок 404/405
	apiRouter := h.SetupRoutes()

	// Оборачиваем весь роутер целиком в Middleware для глобального логирования всех запросов
	loggingRouter := customHttp.Logger(apiRouter.ServeHTTP)

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", loggingRouter); err != nil {
		log.Fatal(err)
	}
}
