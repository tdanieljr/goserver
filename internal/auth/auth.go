package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", nil
	}
	return hash, nil

}
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), Subject: userID.String()})
	out, err := t.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return out, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	t, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	idString, err := t.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.UUID{}, nil
	}
	return id, nil
}
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no auth header")
	}
	if !strings.HasPrefix(authHeader, "Bearer") {
		return "", fmt.Errorf("Wrong Prefix")

	}
	authList := strings.Fields(authHeader)
	if len(authList) != 2 {
		return "", fmt.Errorf("Not enough fields")

	}
	authToken := strings.TrimSpace(authList[1])
	return authToken, nil
}
func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := hex.EncodeToString(b)
	return s, nil
}
