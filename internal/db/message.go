package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
)

type DbError = sqlite.Error

var db *sql.DB
var migrated = false

func Init() (func() error, error) {
	var err error
	db, err = sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	err = runMigrations()
	if err != nil {
		return nil, err
	}
	return func() error { return db.Close() }, nil
}

func runMigrations() error {
	if migrated {
		return errors.New("already migrated")
	}
	migrations := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA synchronous=NORMAL;`,
		`PRAGMA foreign_keys=ON;`,
		`PRAGMA busy_timeout=5000;`,
		`PRAGMA temp_store=MEMORY;`,

		`CREATE TABLE IF NOT EXISTS messages (
				key TEXT PRIMARY KEY,
				msg TEXT NOT NULL,
				ts  INTEGER NOT NULL
		);`,
	}
	for _, stmt := range migrations {
		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("migration failed at [%s]\n%v", stmt, err)
		}
	}
	migrated = true
	return nil
}

func QueryKey(key string) (string, bool, error) {
	rows, err := db.Query("SELECT msg FROM (messages) WHERE key = ?", key)
	if err != nil {
		return "", false, fmt.Errorf("error querying %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var msg string
		err = rows.Scan(&msg)
		if err != nil {
			return "", false, fmt.Errorf("error scanning %w", err)
		}
		return msg, true, nil
	}
	return "", false, nil
}

func Insert(key, msg string) error {
	now := time.Now().UnixMilli()
	_, err := db.Exec("INSERT INTO messages (key, msg, ts) values (?, ?, ?)", key, msg, now)
	return err
}

func QueryOlder(ts int64) ([]string, error) {
	rows, err := db.Query("SELECT key FROM messages WHERE ts < ?", ts)
	if err != nil {
		return nil, fmt.Errorf("error querying %w", err)
	}
	defer rows.Close()
	keys := make([]string, 0, 20)
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, fmt.Errorf("error scanning %w", err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func DeleteKey(key string) error {
	_, err := db.Exec("DELETE FROM messages WHERE key = ?", key)
	if err != nil {
		return err
	}
	return nil
}
