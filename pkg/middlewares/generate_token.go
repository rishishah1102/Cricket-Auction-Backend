package middlewares

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GenerateToken(id primitive.ObjectID, email string) (string, error) {
	// Jwt Secret
	jwtKey := []byte(os.Getenv("TOKEN_SECRET"))

	now := time.Now()
	expTime := now.Add(365 * 24 * time.Hour)

	claims := &Claims{
		Email: email,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id.String(), // Standard `sub` claim
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signing the token with the secret key and fetching the token
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
