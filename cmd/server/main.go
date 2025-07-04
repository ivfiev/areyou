package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/ivfiev/areyou/internal/api"
	"github.com/ivfiev/areyou/internal/cron"
	"github.com/ivfiev/areyou/internal/db"
)

func main() {
	closer, err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer closer()
	cron.Start()
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

// e2e tests
// ttl - ctx & signal shutdown
// chains/breadcrumbs
// rate limit
