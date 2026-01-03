package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func UnsoldPlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request struct {
				PlayerID  primitive.ObjectID `json:"player_id"`
				AuctionID primitive.ObjectID `json:"auction_id"`
			}
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind unsold players request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{
			"_id":        request.PlayerID,
			"auction_id": request.AuctionID,
		}

		update := bson.M{
			"$set": bson.M{
				"hammer": "unsold",
			},
		}

		result, err := db.Collection(constants.PlayerCollection).UpdateOne(ctx, filter, update)
		if err != nil {
			logger.Error("failed to update player status to unsold", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while updating player status"})
			return
		}

		if result.MatchedCount == 0 {
			logger.Error("no player found with the given ID", zap.Any("player_id", request.PlayerID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Player marked as unsold successfully"})
	}
}
