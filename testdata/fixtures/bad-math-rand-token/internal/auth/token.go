package auth

import "math/rand"

func SessionToken() int {
	return rand.Int()
}
