package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	UserID string `json:"userId"`
	Phone  string `json:"phone"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func GenerateJWT(userId, name, phone string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	claims := JwtClaims{
		UserID: userId,
		Phone:  phone,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "delivery-app",
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
