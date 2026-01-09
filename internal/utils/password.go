package utils

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	cryptEncoding = base64.NewEncoding("./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
)

func PasswordHash(password, method string) (string, error) {
	var passwordHash string

	switch method {
	case "bcrypt":
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = "{CRYPT}" + string(hashedPassword)
	default:
		return "", fmt.Errorf("unsupported password hashing method: %s", method)
	}

	return passwordHash, nil
}
