package home

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// UserController fetches the image and username for the home page.
func UserController(c *gin.Context) {
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.L().Error("Failed to configure logger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Fetch email from token
	email, exists := c.Get("email")
	if !exists {
		logger.Error("Email not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in context"})
		return
	}

	// Fetch the users collection
	usersCollection, err := models.UsersCollection(logger)
	if err != nil {
		logger.Error("Unable to get users collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Find user by email
	var user schemas.User
	err = usersCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Error("No user found with provided email", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"message": "No user found with the provided email"})
		} else {
			logger.Error("Error in fetching the data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in fetching the data", "error": err.Error()})
		}
		return
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully fetched the data",
		"foundUser": user,
	})
}
