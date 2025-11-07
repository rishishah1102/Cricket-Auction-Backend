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

func UpdateTeamController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request  models.Team
			response models.Team
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind update team request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{
			"_id":        request.ID,
			"auction_id": request.AuctionId,
		}
		update := bson.M{
			"$set": bson.M{
				"team_name":   request.TeamName,
				"team_image":  request.TeamImage,
				"team_owners": request.TeamOwners,
				"updated_at":  time.Now(),
			},
		}

		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err := db.Collection(constants.TeamCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&response)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logger.Warn("no team found for update", zap.Any(constants.Err, err))
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found or you are not authorized to update it"})
				return
			}
			logger.Error("failed to update team in database", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Team updated successfully",
			"team":    response,
		})
	}
}
