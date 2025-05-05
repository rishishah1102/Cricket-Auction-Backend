package auth

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

// This route is for getting username and email from frontend and sending otp via email
func RegisterController(ctx *gin.Context) {
	var request, response schemas.User

	// Reading logger
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.Error("unable to bind the request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	usersCollection, err := models.UsersCollection(logger)
	if err != nil {
		logger.Error("unable to get users collection", zap.Error(err))
		return
	}

	err = usersCollection.FindOne(context.Background(), bson.M{"email": request.Email}).Decode(&response)
	if err == mongo.ErrNoDocuments {
		// Generating OTP
		otp := utils.GenerateRandomNumber()
		logger.Info("otp generated successfully")

		// Sending email
		if err := utils.SendEmail(request.Email, "Registration Otp", otp); err != nil {
			logger.Error("unable to send email", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}

		logger.Info("OTP sent to the registered Email Id: " + request.Email)
		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Otp is send to " + request.Email,
			"otp":     otp,
		})
		return
	}
	
	if err == nil {
		logger.Warn("user already exist")
		ctx.JSON(http.StatusConflict, gin.H{
			"message": "User already exists",
		})
		return
	}

	logger.Error("an error occured", zap.Error(err))
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
}
