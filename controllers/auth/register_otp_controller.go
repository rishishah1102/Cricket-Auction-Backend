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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// registerOTPRequest is the struct for register otp controller request body
type registerOTPRequest struct {
	models.User
	OTP int `json:"otp"`
}

func RegisterOtpController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request registerOTPRequest
			otp     models.Otp
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind register OTP verification request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err := db.Collection(constants.OtpCollection).FindOne(ctx, bson.M{"email": request.Email}).Decode(&otp)
		if err != nil {
			logger.Error("failed to find otp for the email", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP not found or invalid"})
			return
		}

		if request.OTP != otp.Otp {
			logger.Error("failed to validate OTP")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
			return
		}

		if time.Now().After(otp.ExpiresAt) {
			logger.Error("failed to validate time of OTP")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
			return
		}

		go db.Collection(constants.OtpCollection).DeleteOne(ctx, bson.M{"email": request.Email})

		// Save user in db
		insertedUserId, err := db.Collection(constants.UserCollection).InsertOne(ctx, bson.M{
			"email":      request.Email,
			"mobile":     request.Mobile,
			"first_name": request.FirstName,
			"last_name":  request.LastName,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		})
		if err != nil {
			logger.Error("failed to insert user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error from database"})
			return
		}

		ID, ok := insertedUserId.InsertedID.(primitive.ObjectID)
		if !ok {
			logger.Error("failed to convert primitive id", zap.Any("id", insertedUserId.InsertedID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Id generated is invalid"})
			return
		}

		token, err := middlewares.GenerateToken(ID, request.Email)
		if err != nil {
			logger.Error("failed to generate jwt token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error from token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"token":   token,
		})
	}
}
