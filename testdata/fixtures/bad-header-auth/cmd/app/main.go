package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = r.Header.Get("X-Is-Admin")
	})
	_ = http.ListenAndServe(":8080", nil)
}
