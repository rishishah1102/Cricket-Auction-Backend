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

type APIrequest struct {
	ID        primitive.ObjectID `json:"team_id"`
	AuctionID primitive.ObjectID `json:"auction_id"`
}

func DeleteTeamController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request APIrequest
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind delete team request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{
			"_id":        request.ID,
			"auction_id": request.AuctionID,
		}

		_, err := db.Collection(constants.TeamCollection).DeleteOne(ctx, filter)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logger.Warn("no team found for delete", zap.Any(constants.Err, err))
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found or you are not authorized to delete it"})
				return
			}
			logger.Error("failed to delete team in database", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Team deleted successfully",
		})
	}
}
