package db

import (
	"sync"
)

var db sync.Map

func Read(key string) (string, bool, error) {
	val, ok := db.Load(key)
	if !ok {
		return "", false, nil
	}
	return val.(string), ok, nil
}

func Write(key, msg string) error {
	db.Store(key, msg)
	return nil
}
