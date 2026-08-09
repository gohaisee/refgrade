package auth

import (
	"crypto/md5"
)

func HashPassword(password string) []byte {
	return md5.Sum([]byte(password))
}
