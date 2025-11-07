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

type teamAPIRequest struct {
	AuctionID primitive.ObjectID `json:"auction_id"`
}

func GetAllTeamsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request teamAPIRequest
			teams   []models.Team
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind get all teams request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{"auction_id": request.AuctionID}
		cursor, err := db.Collection(constants.TeamCollection).Find(ctx, filter)
		if err != nil {
			logger.Error("failed to fetch teams", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &teams); err != nil {
			logger.Error("failed to decode teams", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while decoding"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Teams fetched successfully",
			"teams":   teams,
		})
	}
}
