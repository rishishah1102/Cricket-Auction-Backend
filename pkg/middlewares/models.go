package middlewares

import (
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Claims defines the JWT payload structure
type Claims struct {
	Email string             `json:"email"`
	ID    primitive.ObjectID `json:"id"`
	jwt.RegisteredClaims
}
