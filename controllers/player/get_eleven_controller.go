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

func GetElevenController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TeamID primitive.ObjectID `json:"team_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("failed to bind get eleven request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{
				"_id": req.TeamID,
			}}},
			{{Key: "$unwind", Value: "$squad"}},
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.PlayerCollection,
				"localField":   "squad",
				"foreignField": "_id",
				"as":           "player",
			}}},
			{{Key: "$unwind", Value: "$player"}},
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.MatchCollection,
				"localField":   "player.match",
				"foreignField": "_id",
				"as":           "match",
			}}},
			{{Key: "$unwind", Value: "$match"}},
			{{Key: "$match", Value: bson.M{
				"match.currentX1": true,
			}}},
			{{Key: "$project", Value: bson.M{
				"_id":                 "$player._id",
				"player_name":         "$player.player_name",
				"role":                "$player.role",
				"country":             "$player.country",
				"current_team":        "$player.current_team",
				"selling_price":       "$player.selling_price",
				"match_id":            "$player.match",
				"prev_team":           "$player.prev_team",
			}}},
		}

		cursor, err := db.Collection(constants.TeamCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("aggregation failed", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playing eleven"})
			return
		}
		defer cursor.Close(ctx)

		var playingEleven = make([]bson.M, 0)
		if err := cursor.All(ctx, &playingEleven); err != nil {
			logger.Error("failed to decode aggregation result", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode result"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Playing 11 fetched successfully",
			"playing11": playingEleven,
		})
	}
}
