package auction

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// UpdateAuctionAndTeamsController updates the auction and its associated teams.
func UpdateAuctionAndTeamsController(c *gin.Context) {
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

	var request struct {
		Auction schemas.Auction `json:"auction"`
		Teams   []schemas.Team  `json:"teams"`
	}

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Unable to bind the request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Fetch the auctions collection
	auctionsCollection, err := models.AuctionCollection(logger)
	if err != nil {
		logger.Error("Unable to get auctions collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update the auction
	filter := bson.M{"_id": request.Auction.Id, "createdBy": email}
	update := bson.M{"$set": request.Auction}

	var updatedAuction schemas.Auction
	err = auctionsCollection.FindOneAndUpdate(context.Background(), filter, update).Decode(&updatedAuction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Error("No auction found", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "No auction found"})
			return
		}
		logger.Error("Error updating the auction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the auction"})
		return
	}

	// Fetch the teams collection
	teamsCollection, err := models.TeamsCollection(logger)
	if err != nil {
		logger.Error("Unable to get teams collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update or insert teams
	for _, team := range request.Teams {
		// If team.TeamId is not provided (zero value), generate a new ObjectID
		if team.TeamId == primitive.NilObjectID {
			team.TeamId = primitive.NewObjectID()
		}
		teamFilter := bson.M{"_id": team.TeamId, "auctionId": team.AuctionId}
		teamUpdate := bson.M{"$set": team}

		_, err := teamsCollection.UpdateOne(context.Background(), teamFilter, teamUpdate, options.Update().SetUpsert(true))
		if err != nil {
			logger.Error("Error updating team", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating team"})
			return
		}
	}

	// Return the updated auction and teams in the response
	c.JSON(http.StatusOK, gin.H{
		"message": "Auction and teams updated successfully",
		"auction": updatedAuction,
		"teams":   request.Teams,
	})
}
