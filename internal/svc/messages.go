package svc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ivfiev/areyou/internal/db"
)

var (
	BadQuery = errors.New("bad query")
	Conflict = errors.New("conflict")
)

func Query(kws []string) ([]string, bool, error) {
	if len(kws) < 1 || len(kws) > 9 {
		return nil, false, BadQuery
	}
	for _, kw := range kws {
		if len(kw) < 2 || len(kw) > 16 {
			return nil, false, BadQuery
		}
	}
	kws = dedup(kws)
	cat := cat(kws)
	hash := hash(cat)
	msgs, ok, err := db.QueryKey(hash)
	if err != nil {
		return nil, false, mapErr(err)
	}
	if !ok {
		return nil, false, nil
	}
	dmsgs := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		dmsg, err := decrypt(encryptionKey(hash, cat), msg)
		if err != nil {
			return nil, false, fmt.Errorf("decryption error %w", err)
		}
		dmsgs = append(dmsgs, dmsg)
	}
	return dmsgs, true, nil
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
	err = db.Insert(hash, encr)
	return mapErr(err)
}

func cat(kws []string) string {
	return strings.Join(kws, "|")
}

func encryptionKey(hash, key string) string {
	return hash + key
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case *db.DbError:
		if err.Code() == 2067 || err.Code() == 1555 {
			return Conflict
		}
		return fmt.Errorf("db error %w", err)
	}
	return err
}
