package home

import (
	"auction-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HomeController fetches all auctions where the user has joined.
func HomeController(c *gin.Context) {
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

	auctions, err := fetchAllAuctions(email)
	if err != nil {
		logger.Error("Unable to fetch all the auction for the given email", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}

	logger.Info("auctions", zap.Any("auction", auctions))

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Successfully fetched the data",
		"auctions": auctions,
	})
}
