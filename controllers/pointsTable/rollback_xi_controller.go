package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// RollbackXIController reverses the Change XI operation:
//   - currentX1 → nextX1
//   - prevX1 → currentX1
//   - prevX1 stays as-is
//   - Restores prev earned/benched/total points
//   - Resets matches array to zeros
func RollbackXIController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
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

		// Fetch all match IDs for this auction's players.
		cursor, err := db.Collection(constants.PlayerCollection).Find(ctx,
			bson.M{
				"auction_id": request.AuctionID,
				"match":      bson.M{"$exists": true},
			},
			options.Find().SetProjection(bson.M{"match": 1, "_id": 0}),
		)
		if err != nil {
			logger.Error("failed to fetch player match IDs", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		type matchRef struct {
			Match primitive.ObjectID `bson:"match"`
		}
		var refs []matchRef
		cursor.All(ctx, &refs)
		cursor.Close(ctx)

		if len(refs) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "No players to update"})
			return
		}

		matchIDs := make([]primitive.ObjectID, 0, len(refs))
		for _, r := range refs {
			if !r.Match.IsZero() {
				matchIDs = append(matchIDs, r.Match)
			}
		}

		// Rollback: current→next, prev→current, restore prev points, reset matches.
		updatePipeline := mongo.Pipeline{
			{{Key: "$set", Value: bson.D{
				{Key: "nextX1", Value: "$currentX1"},
				{Key: "currentX1", Value: "$prevX1"},
				{Key: "earnedPoints", Value: "$prevEarnedPoints"},
				{Key: "benchedPoints", Value: "$prevBenchedPoints"},
				{Key: "totalPoints", Value: "$prevTotalPoints"},
				{Key: "matches", Value: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			}}},
		}

		result, err := db.Collection(constants.MatchCollection).UpdateMany(ctx,
			bson.M{"_id": bson.M{"$in": matchIDs}},
			updatePipeline,
		)
		if err != nil {
			logger.Error("failed to rollback XI", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback XI"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "XI rollback successful",
			"updated": result.ModifiedCount,
		})
	}
}
