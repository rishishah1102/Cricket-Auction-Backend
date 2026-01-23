package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"slices"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func SaveElevenController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SquadIDs  []primitive.ObjectID `json:"squad" binding:"required"`
			PlayerIDs []primitive.ObjectID `json:"player_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("failed to bind request", zap.Error(err))
			c.JSON(400, gin.H{"error": "Exactly 11 player IDs required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		cursor, err := db.Collection(constants.PlayerCollection).Find(
			ctx,
			bson.M{"_id": bson.M{"$in": req.SquadIDs}},
			options.Find().SetProjection(bson.M{"match": 1}),
		)
		if err != nil {
			logger.Error("player fetch failed", zap.Error(err))
			c.JSON(500, gin.H{"error": "DB error"})
			return
		}
		defer cursor.Close(ctx)

		playing11MatchIDs := make([]primitive.ObjectID, 0, 11)
		nonPlaying11MatchIDs := make([]primitive.ObjectID, len(req.SquadIDs)-11)
		for cursor.Next(ctx) {
			var player models.Player
			if err := cursor.Decode(&player); err != nil {
				logger.Error("failed to decode player", zap.Error(err))
				c.JSON(500, gin.H{"error": "Failed to decode player"})
				return
			}
			if slices.Contains(req.PlayerIDs, player.Id) {
				playing11MatchIDs = append(playing11MatchIDs, player.Match)
			} else {
				nonPlaying11MatchIDs = append(nonPlaying11MatchIDs, player.Match)
			}
		}

		_, err = db.Collection(constants.MatchCollection).UpdateMany(
			ctx,
			bson.M{"_id": bson.M{"$in": playing11MatchIDs}},
			bson.M{"$set": bson.M{"nextX1": true}},
		)
		if err != nil {
			logger.Error("failed to update playing 11", zap.Error(err))
			c.JSON(500, gin.H{"error": "Failed to save playing XI"})
			return
		}

		_, err = db.Collection(constants.MatchCollection).UpdateMany(
			ctx,
			bson.M{"_id": bson.M{"$in": nonPlaying11MatchIDs}},
			bson.M{"$set": bson.M{"nextX1": false}},
		)
		if err != nil {
			logger.Error("failed to update non-playing 11", zap.Error(err))
			c.JSON(500, gin.H{"error": "Failed to save non playing XI"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Playing XI saved",
			"count":   len(playing11MatchIDs),
		})
	}
}
