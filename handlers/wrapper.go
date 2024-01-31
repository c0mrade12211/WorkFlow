package handlers

import (
	"database/sql"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request, *sql.DB)

func WithDB(handler HandlerFunc, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db)
	}
}
