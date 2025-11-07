package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type SaveProfileRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ImageURL     string `json:"image_url"`
	Role         string `json:"role"`
	BattingHand  string `json:"batting_hand"`
	BattingOrder string `json:"batting_order"`
	BattingStyle string `json:"batting_style"`
	BowlingArm   string `json:"bowling_arm"`
	BowlingType  string `json:"bowling_type"`
}

func SaveProfileController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveProfileRequest

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("failed to bind save profile request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Fetch email from token
		email := c.GetString(constants.EmailKey)
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		update := bson.M{
			"first_name":    req.FirstName,
			"last_name":     req.LastName,
			"image_url":     req.ImageURL,
			"role":          req.Role,
			"batting_hand":  req.BattingHand,
			"batting_order": req.BattingOrder,
			"batting_style": req.BattingStyle,
			"bowling_arm":   req.BowlingArm,
			"bowling_type":  req.BowlingType,
			"updated_at":    time.Now(),
		}

		_, err := db.Collection(constants.UserCollection).UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": update})
		if err != nil {
			logger.Error("failed to update user profile", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}
