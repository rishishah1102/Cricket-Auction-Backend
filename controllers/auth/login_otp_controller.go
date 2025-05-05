package auth

import (
	"auction-backend/middlewares"
	"auction-backend/schemas"
	"auction-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// This route is for log in the user and getting a token to make requests
func LoginOtpController(ctx *gin.Context) {
	var request schemas.User

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

	// Generating jwt token
	token, err := middlewares.GenerateToken(request.Email)
	if err != nil {
		logger.Error("unable to generate token", zap.Error(err))
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   err.Error(),
			"message": "Unable to generate token for authentication",
		})
		return
	}

	logger.Info("Login successfully with " + request.Email)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login Successfully",
		"token":   token,
	})
}
