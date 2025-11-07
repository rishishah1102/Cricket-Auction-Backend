package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func CreateTeamController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.Team

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind create team request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		teamDoc := bson.M{
			"team_name":   request.TeamName,
			"team_image":  request.TeamImage,
			"auction_id":  request.AuctionId,
			"team_owners": request.TeamOwners,
			"squad":       make([]primitive.ObjectID, 0),
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		}

		res, err := db.Collection(constants.TeamCollection).InsertOne(ctx, teamDoc)
		if err != nil {
			logger.Error("failed to insert team in auction", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		request.ID = res.InsertedID.(primitive.ObjectID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Team inserted successfully",
			"team":    request,
		})
	}
}
