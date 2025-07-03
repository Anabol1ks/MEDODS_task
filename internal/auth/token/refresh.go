package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const refreshTokenLength = 48

func GenerateRefreshToken() (string, string, error) {
	raw := make([]byte, refreshTokenLength)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	base64Token := base64.StdEncoding.EncodeToString(raw)

	hashed, err := bcrypt.GenerateFromPassword([]byte(base64Token), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return base64Token, string(hashed), nil
}

func CompareRefreshToken(token string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	if err != nil {
		return errors.New("invalid refresh token")
	}
	return nil
}

func ValidateRefreshToken(token, hash string, expiresAt time.Time) error {
	if time.Now().After(expiresAt) {
		return errors.New("refresh token expired")
	}
	return CompareRefreshToken(token, hash)
}
