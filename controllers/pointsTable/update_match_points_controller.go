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

func UpdateMatchPointsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			MatchIndex int `json:"match_index"`
			Updates    []struct {
				MatchID primitive.ObjectID `json:"match_id"`
				Points  int               `json:"points"`
			} `json:"updates" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if request.MatchIndex < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "match_index must be >= 0"})
			return
		}

		// 1. Collect match IDs to fetch
		matchIDs := make([]primitive.ObjectID, 0, len(request.Updates))
		for _, u := range request.Updates {
			matchIDs = append(matchIDs, u.MatchID)
		}

		// 2. Batch fetch all current match docs
		mCursor, err := db.Collection(constants.MatchCollection).Find(ctx,
			bson.M{"_id": bson.M{"$in": matchIDs}},
		)
		if err != nil {
			logger.Error("failed to fetch matches", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		type matchDoc struct {
			ID                primitive.ObjectID `bson:"_id"`
			Matches           []int              `bson:"matches"`
			CurrentX1         bool               `bson:"currentX1"`
			PrevEarnedPoints  int                `bson:"prevEarnedPoints"`
			PrevBenchedPoints int                `bson:"prevBenchedPoints"`
		}
		var docs []matchDoc
		mCursor.All(ctx, &docs)
		mCursor.Close(ctx)

		// Build points lookup from request
		pointsReq := make(map[primitive.ObjectID]int, len(request.Updates))
		for _, u := range request.Updates {
			pointsReq[u.MatchID] = u.Points
		}

		// 3. Build bulk write operations (single DB round-trip, no goroutines)
		ops := make([]mongo.WriteModel, 0, len(docs))
		for _, d := range docs {
			newPoints := pointsReq[d.ID]

			// Extend array if needed
			matches := make([]int, len(d.Matches))
			copy(matches, d.Matches)
			for len(matches) <= request.MatchIndex {
				matches = append(matches, 0)
			}
			matches[request.MatchIndex] = newPoints

			weekTotal := 0
			for _, p := range matches {
				weekTotal += p
			}

			// Cumulative: prev + this week's match points
			earned := d.PrevEarnedPoints
			benched := d.PrevBenchedPoints
			if d.CurrentX1 {
				earned += weekTotal
			} else {
				benched += weekTotal
			}
			total := earned + benched

			op := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": d.ID}).
				SetUpdate(bson.M{"$set": bson.M{
					"matches":       matches,
					"totalPoints":   total,
					"earnedPoints":  earned,
					"benchedPoints": benched,
				}})
			ops = append(ops, op)
		}

		if len(ops) > 0 {
			_, err := db.Collection(constants.MatchCollection).BulkWrite(ctx, ops)
			if err != nil {
				logger.Error("bulk update failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save points"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Points updated successfully"})
	}
}
