package auth

import "github.com/golang-jwt/jwt/v5"

func AcceptNone() {
	_ = jwt.SigningMethodNone
}
