package handler

import (
	"net/http"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Users(w http.ResponseWriter, r *http.Request) {
	pool, err := pgxpool.New(r.Context(), "postgres://localhost/app")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pool.Close()
	_, _ = pool.Exec(r.Context(), "SELECT 1")
}
