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

// This route is for getting mobile and email from frontend and sending otp via email
func RegisterController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request, user models.User

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind register request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Check if user exits
		err := db.Collection(constants.UserCollection).FindOne(ctx, bson.M{"email": request.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				var otpDoc models.Otp

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
				err := db.Collection(constants.OtpCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&otpDoc)
				if err != nil {
					logger.Error("failed to save otp", zap.Any(constants.Err, err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
					return
				}

				logger.Info("sending register otp on email", zap.Any("email", request.Email))

				err = utils.SendEmail(request.Email, "Registration OTP", otp, logger)
				if err != nil {
					logger.Error("failed to send otp email", zap.Any(constants.Err, err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"message": "OTP sent to email",
				})
				return
			}

			logger.Error("failed to fetch data from db", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		logger.Warn("user already exists", zap.String("email", request.Email))
		c.JSON(http.StatusConflict, gin.H{"error": "Account already exists"})
	}
}
