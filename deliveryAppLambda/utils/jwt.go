package utils

import (
	"errors"
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

// ValidateToken verifies the JWT token and returns the claims if valid
func ValidateToken(tokenString string) (*JwtClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
