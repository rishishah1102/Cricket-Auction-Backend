package auction

import (
	"auction-backend/models"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// FetchAllAuctionsController fetches all auctions where the user has joined along with their respective teams using aggregation.
func FetchAllAuctionsController(c *gin.Context) {
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

	// Get the auctions collection
	auctionsCollection, err := models.AuctionCollection(logger)
	if err != nil {
		logger.Error("Unable to get auctions collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Define aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"joinedBy": bson.M{"$elemMatch": bson.M{"email": email}}}}},
		{{Key: "$sort", Value: bson.M{"createdAt": -1}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "teams",
			"localField":   "_id",
			"foreignField": "auctionId",
			"as":           "teams",
		}}},
	}

	// Execute aggregation
	cursor, err := auctionsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		logger.Error("Error executing aggregation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching auctions with teams"})
		return
	}
	defer cursor.Close(context.Background())

	// Decode results
	var result []bson.M
	if err = cursor.All(context.Background(), &result); err != nil {
		logger.Error("Error decoding aggregation result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding data"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Successfully fetched the data",
		"auctions": result,
	})
}
