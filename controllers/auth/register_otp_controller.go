package auth

import (
	"auction-backend/middlewares"
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterOtpController(c *gin.Context) {
	var request schemas.User

	// Reading logger
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	usersCollection, err := models.UsersCollection(logger)
	if err != nil {
		logger.Error("unable to get matches collection", zap.Error(err))
		return
	}

	_, err = usersCollection.InsertOne(context.Background(), request)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generating jwt token
	token, err := middlewares.GenerateToken(request.Email)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   err.Error(),
			"message": "Unable to generate token for authentication",
		})
		return
	}

	logger.Info("Registration successfully with " + request.Email)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration Successfully",
		"token":   token,
	})
}
