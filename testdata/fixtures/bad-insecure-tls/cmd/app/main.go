package main

import "crypto/tls"

func main() {
	_ = tls.Config{InsecureSkipVerify: true}
}
