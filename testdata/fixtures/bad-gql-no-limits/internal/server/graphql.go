package server

import "github.com/99designs/gqlgen/graphql/handler"

func NewGraphQL() *handler.Server {
	return handler.NewDefaultServer(nil)
}
