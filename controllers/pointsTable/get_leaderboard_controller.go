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

func GetLeaderboardController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
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

		// Single aggregation pipeline: teams → lookup players → lookup matches → group
		pipeline := mongo.Pipeline{
			// Match teams in this auction
			{{Key: "$match", Value: bson.M{"auction_id": request.AuctionID}}},

			// Lookup players from squad
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.PlayerCollection,
				"localField":   "squad",
				"foreignField": "_id",
				"as":           "players",
			}}},

			// Unwind players (preserve teams with empty squads)
			{{Key: "$unwind", Value: bson.M{
				"path":                       "$players",
				"preserveNullAndEmptyArrays": true,
			}}},

			// Lookup match for each player
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.MatchCollection,
				"localField":   "players.match",
				"foreignField": "_id",
				"as":           "match",
			}}},

			// Unwind match
			{{Key: "$unwind", Value: bson.M{
				"path":                       "$match",
				"preserveNullAndEmptyArrays": true,
			}}},

			// Group by team — sum points across all squad players
			{{Key: "$group", Value: bson.M{
				"_id":            "$_id",
				"team_name":      bson.M{"$first": "$team_name"},
				"team_image":     bson.M{"$first": "$team_image"},
				"earned_points":  bson.M{"$sum": bson.M{"$ifNull": []any{"$match.earnedPoints", 0}}},
				"benched_points": bson.M{"$sum": bson.M{"$ifNull": []any{"$match.benchedPoints", 0}}},
				"total_points":   bson.M{"$sum": bson.M{"$ifNull": []any{"$match.totalPoints", 0}}},
			}}},

			// Sort by earned points descending, then benched points descending
			{{Key: "$sort", Value: bson.D{
				{Key: "earned_points", Value: -1},
				{Key: "benched_points", Value: -1},
			}}},

			// Clean up output
			{{Key: "$project", Value: bson.M{
				"_id":            0,
				"team_name":      1,
				"team_image":     1,
				"earned_points":  1,
				"benched_points": 1,
				"total_points":   1,
			}}},
		}

		cursor, err := db.Collection(constants.TeamCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("leaderboard aggregation failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer cursor.Close(ctx)

		var leaderboard []bson.M
		if err = cursor.All(ctx, &leaderboard); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
			return
		}

		if leaderboard == nil {
			leaderboard = []bson.M{}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Leaderboard fetched successfully",
			"leaderboard": leaderboard,
		})
	}
}
