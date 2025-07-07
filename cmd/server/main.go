package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ivfiev/areyou/internal/api"
	"github.com/ivfiev/areyou/internal/cron"
	"github.com/ivfiev/areyou/internal/db"
)

func main() {
	shuttingDown := false
	closeDb, err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer closeDb()
	ctx, cancelCron := context.WithCancel(context.Background())
	defer cancelCron()
	cron.Start(ctx)
	slog.Info("cron running")

	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if shuttingDown {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("server is shutting down"))
			return
		}
		switch r.Method {
		case http.MethodGet:
			api.Get(w, r)
		case http.MethodPost:
			api.Post(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	slog.Info("server running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)
	<-stop
	slog.Info("shutting down...")
	shuttingDown = true
	cancelCron()
	time.Sleep(1 * time.Second)
}

// e2e tests
// chains/breadcrumbs
// rate limit
// tf deploy + lw
