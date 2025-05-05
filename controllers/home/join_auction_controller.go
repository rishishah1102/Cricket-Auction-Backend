package home

import (
	"auction-backend/models"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// JoinAuctionController allows a user to join an auction by adding their email and name to the joinedBy array.
func JoinAuctionController(c *gin.Context) {
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

	var request struct {
		ID string `json:"auctionId"`
	}

	// Bind the request body to the request struct
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Unable to bind the request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert the string ID to an ObjectId
	objID, err := primitive.ObjectIDFromHex(request.ID)
	if err != nil {
		logger.Error("Invalid auction ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid auction ID"})
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

	// Get the auctions collection
	auctionsCollection, err := models.AuctionCollection(logger)
	if err != nil {
		logger.Error("Unable to get auctions collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Define the filter to find the auction document
	filter := bson.M{"_id": objID}

	// Define the update to add the email and name to the joinedBy array
	update := bson.M{
		"$push": bson.M{"joinedBy": bson.M{"email": userEmail, "name": fullName}},
	}

	// Perform the update operation
	result := auctionsCollection.FindOneAndUpdate(context.TODO(), filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			logger.Error("No auction found with provided ID", zap.Error(result.Err()))
			c.JSON(http.StatusNotFound, gin.H{"error": "No auction found with the provided ID"})
		} else {
			logger.Error("Error updating the auction", zap.Error(result.Err()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the auction"})
		}
		return
	}

	// Fetch all auctions for the given email
	newAuctions, err := fetchAllAuctions(userEmail)
	if err != nil {
		logger.Error("Unable to fetch all auctions for the given email", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email added successfully", "auctions": newAuctions})
}
