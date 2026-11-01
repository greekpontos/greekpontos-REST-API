package http

import (
	"log"
	"net/http"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] Request received: %s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}
