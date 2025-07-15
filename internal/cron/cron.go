package cron

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/ivfiev/areyou/internal/db"
)

var MSG_TTL_MS = int64(48 * 3600 * 1000)
var MSG_TTL_FREQ_MS = int64(5 * 60 * 1000)

func Start(ctx context.Context) {
	setup()
	go cronTTL(ctx)
}

func cronTTL(ctx context.Context) {
	ttl := time.NewTicker(time.Millisecond * time.Duration(MSG_TTL_FREQ_MS))
	for {
		select {
		case <-ttl.C:
			now := time.Now().UnixMilli()
			keys, err := db.QueryOlder(now - MSG_TTL_MS)
			if err != nil {
				slog.Error("error querying expired messages", "err", err)
				continue
			}
			fails := 0
			for _, key := range keys {
				err = db.DeleteKey(key)
				if err != nil {
					slog.Error("error deleting expired message "+key, "err", err)
					fails++
				}
			}
			slog.Info("deleted expired items", "successes", len(keys), "fails", fails)
		case <-ctx.Done():
			slog.Info("shutting down cron")
			return
		}
	}
}

func setup() {
	ttl, ok := parseTime("MSG_TTL_MS")
	if ok {
		MSG_TTL_MS = ttl
	}
	freq, ok := parseTime("MSG_TTL_FREQ_MS")
	if ok {
		MSG_TTL_FREQ_MS = freq
	}
}

func parseTime(key string) (int64, bool) {
	str, ok := os.LookupEnv(key)
	if ok {
		val, err := strconv.Atoi(str)
		if err != nil {
			slog.Error("error parsing "+key, "err", err)
			return 0, false
		}
		return int64(val), true
	}
	return 0, false
}
