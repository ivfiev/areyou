package svc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ivfiev/areyou/internal/db"
)

var (
	BadQuery = errors.New("bad query")
)

func Query(kws []string) (string, bool, error) {
	if len(kws) < 1 || len(kws) > 9 {
		return "", false, BadQuery
	}
	for _, kw := range kws {
		if len(kw) < 2 || len(kw) > 16 {
			return "", false, BadQuery
		}
	}
	kws = dedup(kws)
	cat := cat(kws)
	hash := hash(cat)
	msg, ok, err := db.Read(hash)
	if err != nil {
		return "", false, fmt.Errorf("db read error %w", err)
	}
	if !ok {
		return "", false, nil
	}
	dmsg, err := decrypt(encryptionKey(hash, cat), msg)
	if err != nil {
		return "", true, fmt.Errorf("decryption error %w", err)
	}
	return dmsg, true, nil
}

func Create(kws []string, msg string) error {
	if len(kws) < 1 || len(kws) > 9 {
		return BadQuery
	}
	for _, kw := range kws {
		if len(kw) < 2 || len(kw) > 16 {
			return BadQuery
		}
	}
	kws = dedup(kws)
	cat := cat(kws)
	hash := hash(cat)
	encr, err := encrypt(encryptionKey(hash, cat), msg)
	if err != nil {
		return fmt.Errorf("encryption error %w", err)
	}
	err = db.Write(hash, encr)
	if err != nil {
		return fmt.Errorf("db write error %w", err)
	}
	return nil
}

func cat(kws []string) string {
	return strings.Join(kws, "|")
}

func encryptionKey(hash, key string) string {
	return hash + key
}
