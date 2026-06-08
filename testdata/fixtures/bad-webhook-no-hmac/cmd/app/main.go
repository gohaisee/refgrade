package main

import "net/http"

func stripeWebhook(w http.ResponseWriter, r *http.Request) {}

func main() {
	http.HandleFunc("/webhook/stripe", stripeWebhook)
}
