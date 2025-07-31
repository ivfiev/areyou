package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ivfiev/areyou/internal/db"
)

var errRequest = errors.New("bad request")
var errNotFound = errors.New("not found")

func Get(w http.ResponseWriter, r *http.Request) {
	urlq := r.URL.Query()
	query, ok := urlq["keywords"]
	if !ok {
		handleErr(errRequest, w)
		return
	}
	kws := strings.Split(query[0], ",")
	if len(kws) < 1 || len(kws) > 9 {
		handleErr(errRequest, w)
		return
	}
	for _, kw := range kws {
		if len(kw) < 2 || len(kw) > 16 {
			handleErr(errRequest, w)
			return
		}
	}
	kws = dedup(kws)
	cat := cat(kws)
	hash := hash(cat)
	msgs, ok, err := db.QueryKey(hash)
	if err != nil {
		handleErr(err, w)
		return
	}
	if !ok {
		handleErr(errNotFound, w)
		return
	}
	dmsgs := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		dmsg, err := decrypt(encryptionKey(hash, cat), msg)
		if err != nil {
			handleErr(err, w)
			return
		}
		dmsgs = append(dmsgs, dmsg)
	}
	resp := GetMessagesResponse{
		Messages: dmsgs,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func Post(w http.ResponseWriter, r *http.Request) {
	var body PostMessageRequest
	json.NewDecoder(r.Body).Decode(&body)
	kws := body.Keywords
	msg := body.Message
	if len(kws) < 1 || len(kws) > 9 {
		handleErr(errRequest, w)
		return
	}
	for _, kw := range kws {
		if len(kw) < 2 || len(kw) > 16 {
			handleErr(errRequest, w)
			return
		}
	}
	kws = dedup(kws)
	cat := cat(kws)
	hash := hash(cat)
	encr, err := encrypt(encryptionKey(hash, cat), msg)
	if err != nil {
		handleErr(err, w)
		return
	}
	err = db.Insert(hash, encr)
	if err != nil {
		handleErr(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleErr(err error, w http.ResponseWriter) {
	switch err {
	case errRequest:
		writeError(w, http.StatusBadRequest, err.Error())
	case errNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	default:
		switch err.(type) {
		case *db.DbError:
			writeError(w, http.StatusConflict, err.Error())
		default:
			slog.Error("internal server error", "err", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
	}
}

func writeError(w http.ResponseWriter, statusCode int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, err)
}
