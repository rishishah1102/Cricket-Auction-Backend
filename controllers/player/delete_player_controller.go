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

func DeletePlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			PlayerID primitive.ObjectID `json:"player_id" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind delete player request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// First get the player to know the auction ID for cache cleanup
		var player models.Player
		err := db.Collection(constants.PlayerCollection).FindOne(ctx, bson.M{"_id": request.PlayerID}).Decode(&player)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
				return
			}
			logger.Error("failed to find player for deletion", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find player"})
			return
		}

		// Delete the corresponding match document
		if !player.Match.IsZero() {
			matchFilter := bson.M{"_id": player.Match}
			_, err = db.Collection(constants.MatchCollection).DeleteOne(ctx, matchFilter)
			if err != nil {
				logger.Warn("failed to delete match document for player", zap.Any(constants.Err, err), zap.Any("matchId", player.Match))
			}
		}

		// Delete the player
		_, err = db.Collection(constants.PlayerCollection).DeleteOne(ctx, bson.M{"_id": request.PlayerID})
		if err != nil {
			logger.Error("failed to delete player", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete player"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Player deleted successfully",
		})
	}
}
