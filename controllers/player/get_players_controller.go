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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// GetAllPlayersController with Redis caching
func GetAllPlayersController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			AuctionID primitive.ObjectID `json:"auction_id" binding:"required"`
		}
		var players []models.Player

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind get all players request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{"auction_id": request.AuctionID}

		findOptions := options.Find().SetSort(bson.D{{Key: "player_number", Value: 1}}) // Sort Asc

		cursor, err := db.Collection(constants.PlayerCollection).Find(ctx, filter, findOptions)
		if err != nil {
			logger.Error("failed to fetch players", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &players); err != nil {
			logger.Error("failed to decode players", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while decoding"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Players fetched successfully",
			"players": players,
		})
	}
}
