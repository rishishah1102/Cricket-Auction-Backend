package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"cric-auction-monolith/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// loginRequest is the struct for login controller request body
type loginRequest struct {
	Email string `json:"email"`
}

// This route logs in the user. It takes the email from user and sends otp
func LoginController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request loginRequest
			user    models.User
			otpDoc  models.Otp
		)
		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to login bind request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if user exits
		err := db.Collection(constants.UserCollection).FindOne(ctx, bson.M{"email": request.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logger.Error("failed to fetch user data", zap.Any(constants.Err, err))
				c.JSON(http.StatusNotFound, gin.H{"error": "User does not exists in db"})
				return
			}

			logger.Error("failed to fetch data from db", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		otp := utils.GenerateRandomNumber()

		filter := bson.M{"email": request.Email}
		update := bson.M{
			"$set": bson.M{
				"otp":        otp,
				"expires_at": time.Now().Add(5 * time.Minute),
				"updated_at": time.Now(),
			},
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
		}
		opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

		// Save OTP to DB
		err = db.Collection(constants.OtpCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&otpDoc)
		if err != nil {
			logger.Error("failed to save otp", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
			return
		}

		go utils.SendEmail(request.Email, "Login OTP", otp, logger)

		c.JSON(http.StatusOK, gin.H{
			"message": "OTP sent to email",
		})
	}
}
