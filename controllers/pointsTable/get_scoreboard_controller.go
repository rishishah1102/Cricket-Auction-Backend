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

func GetScoreboardController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			AuctionID primitive.ObjectID `json:"auction_id" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Single aggregation pipeline: players → lookup matches → sort → project
		pipeline := mongo.Pipeline{
			// Match IPL players in this auction
			{{Key: "$match", Value: bson.M{
				"auction_id": request.AuctionID,
				"match":      bson.M{"$exists": true},
			}}},

			// Lookup the match document for each player
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.MatchCollection,
				"localField":   "match",
				"foreignField": "_id",
				"as":           "match_data",
			}}},

			// Unwind match_data array to object
			{{Key: "$unwind", Value: bson.M{
				"path":                       "$match_data",
				"preserveNullAndEmptyArrays": true,
			}}},

			// Sort by earned points descending, then benched points descending
			{{Key: "$sort", Value: bson.D{
				{Key: "match_data.earnedPoints", Value: -1},
				{Key: "match_data.benchedPoints", Value: -1},
			}}},

			// Project only needed fields
			{{Key: "$project", Value: bson.M{
				"player_name":  1,
				"ipl_team":     1,
				"role":         1,
				"current_team": 1,
				"match_data":   1,
			}}},
		}

		cursor, err := db.Collection(constants.PlayerCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("scoreboard aggregation failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer cursor.Close(ctx)

		var players []bson.M
		if err = cursor.All(ctx, &players); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
			return
		}

		if players == nil {
			players = []bson.M{}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Scoreboard fetched successfully",
			"players": players,
		})
	}
}
