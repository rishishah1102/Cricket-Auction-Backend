package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/middlewares"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// loginOTPRequest is the struct for login otp controller request body
type loginOTPRequest struct {
	Email string `json:"email"`
	OTP   int    `json:"otp"`
}

// This route is for log in the user and getting a token to make requests
func LoginOtpController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request loginOTPRequest

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind login OTP verification request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "email", Value: request.Email}}}},
			{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: constants.UserCollection},
				{Key: "localField", Value: "email"},
				{Key: "foreignField", Value: "email"},
				{Key: "as", Value: "user"},
			}}},
			{{Key: "$unwind", Value: "$user"}},
		}

		cursor, err := db.Collection(constants.OtpCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("failed to fetch otp or user", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		var result struct {
			Otp  models.Otp  `bson:"inline"`
			User models.User `bson:"user"`
		}
		if cursor.Next(ctx) {
			if err := cursor.Decode(&result); err != nil {
				logger.Error("failed to decode document", zap.Any(constants.Err, err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
				return
			}
		}

		if request.OTP != result.Otp.Otp {
			logger.Error("failed to validate OTP")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
			return
		}

		if time.Now().After(result.Otp.ExpiresAt) {
			logger.Error("failed to validate time of OTP")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
			return
		}

		token, err := middlewares.GenerateToken(result.User.ID, request.Email)
		if err != nil {
			logger.Error("failed to generate jwt token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User login successful",
			"token":   token,
		})
	}
}
