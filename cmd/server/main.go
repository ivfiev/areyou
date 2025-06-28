package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/ivfiev/areyou/internal/api"
)

func main() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.Get(w, r)
		case http.MethodPost:
			api.Post(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	slog.Info("server running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
