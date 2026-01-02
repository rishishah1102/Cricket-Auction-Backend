package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func FetchPlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			AuctionID primitive.ObjectID `json:"auction_id" binding:"required"`
		}
		var player models.Player
		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind fetch player request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Build aggregation pipeline to fetch a random player
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{
				"auction_id": request.AuctionID,
				"hammer":     "upcoming",
			}}},
			{{Key: "$sample", Value: bson.M{"size": 1}}},
		}

		cursor, err := db.Collection(constants.PlayerCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("failed to execute aggregation", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}
		defer cursor.Close(ctx)

		if cursor.Next(ctx) {
			if err := cursor.Decode(&player); err != nil {
				logger.Error("failed to decode player", zap.Any(constants.Err, err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode player data"})
				return
			}
		} else {
			logger.Info("no player found", zap.Any("auction_id", request.AuctionID))
			c.JSON(http.StatusNotFound, gin.H{"error": "No player found with the specified criteria"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Player fetched successfully",
			"player":  player,
		})
	}
}
