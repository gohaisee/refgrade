package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
)

func setup(h *handler.Server) {
	h.Use(extension.Introspection{})
}
