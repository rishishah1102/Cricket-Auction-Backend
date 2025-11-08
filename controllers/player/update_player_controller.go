package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func UpdatePlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var player models.Player

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&player); err != nil {
			logger.Error("failed to bind update player request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		if player.Id.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID is required"})
			return
		}

		// Set updated timestamp
		player.UpdatedAt = time.Now()

		filter := bson.M{"_id": player.Id}
		updateOptions := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var updatedPlayer models.Player
		err := db.Collection(constants.PlayerCollection).FindOneAndUpdate(ctx, filter, player, updateOptions).Decode(&updatedPlayer)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
				return
			}
			logger.Error("failed to update player", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Player updated successfully",
			"player":  updatedPlayer,
		})
	}
}
