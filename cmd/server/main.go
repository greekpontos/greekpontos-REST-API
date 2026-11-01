package main

import (
	"log"
	"net/http"

	"tasks-api/internal/handlers"
	customHttp "tasks-api/internal/http"
	"tasks-api/internal/storage"
)

func main() {
	store := storage.NewMemoryStorage()
	h := handlers.New(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", customHttp.Logger(h.TasksCollection))
	mux.HandleFunc("/tasks/", customHttp.Logger(h.TaskItem))
	mux.HandleFunc("/health", customHttp.Logger(h.HealthCheck))

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
