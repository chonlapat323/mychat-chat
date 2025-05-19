package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ImageURL string `json:"image_url"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenStr string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("❗ JWT_SECRET is empty. Please check your environment variable.")
		return nil, errors.New("missing JWT secret")
	}
	log.Println("🔐 JWT_SECRET loaded")

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("❌ Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		log.Println("❌ Error parsing JWT:", err)
		return nil, err
	}

	if !token.Valid {
		log.Println("❌ Token is invalid")
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		log.Println("❌ Cannot cast token.Claims to *Claims")
		return nil, errors.New("invalid token claims")
	}

	log.Printf("✅ Token is valid. ExpiresAt: %v | Current time: %v", claims.ExpiresAt.Time, time.Now())

	return claims, nil
}
