package client

import "net/http"

func Fetch() (*http.Response, error) {
	c := http.Client{}
	return c.Get("https://example.com")
}
