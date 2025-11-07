package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type joinAuctionRequest struct {
	AuctionID primitive.ObjectID `json:"auction_id"`
}

func JoinAuctionController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		var (
			request  joinAuctionRequest
			response models.Auction
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind join auction request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		email := c.GetString("email")
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		// First check if user is already joined
		alreadyJoinedFilter := bson.M{
			"_id":       request.AuctionID,
			"joined_by": email,
		}

		count, err := db.Collection(constants.AuctionCollection).CountDocuments(ctx, alreadyJoinedFilter)
		if err != nil {
			logger.Error("failed to check if user already joined", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already joined this auction"})
			return
		}

		filter := bson.M{
			"_id": request.AuctionID,
		}
		updateQuery := bson.M{
			"$addToSet": bson.M{
				"joined_by": email, // Changed from object to simple string
			},
		}

		err = db.Collection(constants.AuctionCollection).FindOneAndUpdate(ctx, filter, updateQuery).Decode(&response)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				logger.Warn("invalid auction id", zap.Any("object_id", request.AuctionID))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Auction not found"})
				return
			}

			logger.Error("failed to join auction", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully joined the auction",
			"auction": response,
		})
	}
}
