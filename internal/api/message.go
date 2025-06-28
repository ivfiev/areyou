package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ivfiev/areyou/internal/db"
	"github.com/ivfiev/areyou/internal/svc"
)

func Get(w http.ResponseWriter, r *http.Request) {
	urlq := r.URL.Query()
	query, ok := urlq["keywords"]
	if !ok {
		writeError(w, http.StatusBadRequest, "bad query")
		return
	}
	kws := strings.Split(query[0], ",")
	hash, err := svc.Hash(kws)
	if err != nil {
		handleSvcErr(err, w)
		return
	}
	msg, ok, err := db.Read(hash)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "%s"}`, msg)
}

func Post(w http.ResponseWriter, r *http.Request) {
	var body PostMessage
	json.NewDecoder(r.Body).Decode(&body)
	hash, err := svc.Hash(body.Keywords)
	if err != nil {
		handleSvcErr(err, w)
		return
	}
	err = db.Write(hash, body.Message)
	if err != nil {
		handleSvcErr(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleSvcErr(err error, w http.ResponseWriter) {
	if err != nil {
		// switch default 500
		writeError(w, http.StatusBadRequest, err.Error())
	}
}

func writeError(w http.ResponseWriter, statusCode int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, err)
}
