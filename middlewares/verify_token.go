package middlewares

import (
	"auction-backend/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

func VerifyToken(c *gin.Context) {
	// Reading logger
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return
	}

	// jwt Secret
	jwtKey := []byte(os.Getenv("TOKEN_SECRET"))

	// fetching token from header of request
	headerToken := c.Request.Header.Get("Authorization")
	if headerToken == "" {
		logger.Warn("A token is required for authentication")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "A token is required for authentication",
		})
		return
	}

	// Parse the token
	token, err := jwt.Parse(headerToken, func(token *jwt.Token) (interface{}, error) {
		// checking the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("There was an error matching token sign method.", zap.Error(err))
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		logger.Info("Token is successfully  parsed")
		return jwtKey, nil
	})

	// validating token
	if err != nil {
		logger.Error("token is unauthorized", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		return
	}
	if !token.Valid {
		logger.Warn("token is invalid")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	} 

	// fetching claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Warn("cannot convert token claims to MapClaims")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse token claims",
		})
		return
	}

	// extracting email from token claims
	email, ok := claims["email"].(string)
	if !ok {
		logger.Warn("failed to extract email from token claims")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to extract email from token claims",
		})
		c.Abort()
		return
	}

	// Store the email in the context for further use
	c.Set("email", email)

	// if token is valid forwarding request
	c.Next()
}
