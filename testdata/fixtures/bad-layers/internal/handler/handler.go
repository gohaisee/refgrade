package handler

import "github.com/gohaisee/refgrade/fixtures/bad-layers/internal/repo"

type Handler struct {
	r *repo.Repo
}

func New() *Handler {
	return &Handler{r: repo.New()}
}
