package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
	msg, ok, err := svc.Query(kws)
	if err != nil {
		handleSvcErr(err, w)
		return
	}
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
	err := svc.Create(body.Keywords, body.Message)
	if err != nil {
		handleSvcErr(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleSvcErr(err error, w http.ResponseWriter) {
	if err != nil {
		switch err {
		case svc.BadQuery:
			writeError(w, http.StatusBadRequest, err.Error())
		case svc.Conflict:
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
