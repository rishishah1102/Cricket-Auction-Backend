package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func GetProfileController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		// Fetch email from token
		email := c.GetString(constants.EmailKey)
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		err := db.Collection(constants.UserCollection).FindOne(ctx, bson.M{"email": email}).Decode(&user)
		if err != nil {
			logger.Error("failed to fetch user profile", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"profile": user, "message": "Profile fetched successfully"})
	}
}
