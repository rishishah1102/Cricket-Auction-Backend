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

func GetTeamDetailsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			AuctionID primitive.ObjectID `json:"auction_id" binding:"required"`
			TeamName  string             `json:"team_name" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Single aggregation: team → lookup squad players → lookup matches → sort
		pipeline := mongo.Pipeline{
			// Match the specific team
			{{Key: "$match", Value: bson.M{
				"auction_id": request.AuctionID,
				"team_name":  request.TeamName,
			}}},

			// Lookup players from squad
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.PlayerCollection,
				"localField":   "squad",
				"foreignField": "_id",
				"as":           "players",
			}}},

			// Unwind players
			{{Key: "$unwind", Value: bson.M{
				"path":                       "$players",
				"preserveNullAndEmptyArrays": false,
			}}},

			// Lookup match for each player
			{{Key: "$lookup", Value: bson.M{
				"from":         constants.MatchCollection,
				"localField":   "players.match",
				"foreignField": "_id",
				"as":           "match_data",
			}}},

			// Unwind match
			{{Key: "$unwind", Value: bson.M{
				"path":                       "$match_data",
				"preserveNullAndEmptyArrays": true,
			}}},

			// Sort by earned points descending, then benched points descending
			{{Key: "$sort", Value: bson.D{
				{Key: "match_data.earnedPoints", Value: -1},
				{Key: "match_data.benchedPoints", Value: -1},
			}}},

			// Project needed fields
			{{Key: "$project", Value: bson.M{
				"_id":         0,
				"player_name": "$players.player_name",
				"ipl_team":    "$players.ipl_team",
				"role":        "$players.role",
				"match_data":  1,
			}}},
		}

		cursor, err := db.Collection(constants.TeamCollection).Aggregate(ctx, pipeline)
		if err != nil {
			logger.Error("team details aggregation failed", zap.Error(err))
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
			"message": "Team details fetched successfully",
			"players": players,
		})
	}
}
