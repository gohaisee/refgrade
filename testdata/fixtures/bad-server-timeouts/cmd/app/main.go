package main

import "net/http"

func main() {
	srv := &http.Server{Addr: ":8080"}
	_ = http.ListenAndServe(srv.Addr, nil)
}
