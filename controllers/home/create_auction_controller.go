package home

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// CreateAuctionController creates a new auction
func CreateAuctionController(c *gin.Context) {
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.L().Error("Failed to configure logger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Fetch email from the context (assumed to be set by middleware)
	email, exists := c.Get("email")
	if !exists {
		logger.Error("Email not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found in context"})
		return
	}
	userEmail := email.(string)

	var request schemas.Auction
	// Bind the request body to the Auction struct
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Unable to bind the request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the auctions collection
	auctionsCollection, err := models.AuctionCollection(logger)
	if err != nil {
		logger.Error("Unable to get auctions collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get user details from the database
	user, err := getUserDetails(userEmail, logger)
	if err != nil {
		logger.Error("Error fetching user details", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user details"})
		return
	}

	// Construct full name
	fullName := user.FirstName + " " + user.LastName

	// Create the auction document
	auctionDoc := bson.M{
		"auctionImg":         request.AuctionImg,
		"auctionName":        request.AuctionName,
		"createdBy":          email,
		"joinedBy":           []bson.M{{"email": userEmail, "name": fullName}},
		"createdAt":          time.Now().Unix(),
		"pointsTableChecked": false,
	}

	// Insert the auction into the database
	_, err = auctionsCollection.InsertOne(context.Background(), auctionDoc)
	if err != nil {
		logger.Error("Failed to create an auction instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create auction"})
		return
	}

	newAuctions, err := fetchAllAuctions(email)
	if err != nil {
		logger.Error("Unable to fetch all the auction for the given email", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}

	// Return success response
	logger.Info("Auction created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message":  "New Auction created successfully",
		"auctions": newAuctions,
	})
}
