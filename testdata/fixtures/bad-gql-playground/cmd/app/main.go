package main

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	http.Handle("/playground", playground.Handler("API", "/query"))
}
