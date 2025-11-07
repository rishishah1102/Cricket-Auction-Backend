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

func SavePlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			players []models.Player
			match   models.Match
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&players); err != nil {
			logger.Error("failed to bind save player request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		isIPLAuction := c.Query("isIPLAuction")
		if isIPLAuction == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request query params"})
			return
		}

		for _, player := range players {
			if isIPLAuction == "true" {
				match = models.Match{
					Id:                primitive.NewObjectID(),
					Matches:           []int{},
					PrevX1:            false,
					CurrentX1:         false,
					NextX1:            false,
					EarnedPoints:      0,
					BenchedPoints:     0,
					TotalPoints:       0,
					PrevTotalPoints:   0,
					PrevEarnedPoints:  0,
					PrevBenchedPoints: 0,
				}

				// Save match to database
				_, err := db.Collection(constants.MatchCollection).InsertOne(ctx, match)
				if err != nil {
					logger.Error("failed to create match document for player", zap.Any(constants.Err, err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match record"})
					return
				}

				if player.PrevFantasyPoints == 0 {
					player.PrevFantasyPoints = 0
				}
			}

			player.Match = match.Id
			if len(player.PrevTeam) == 0 {
				player.PrevTeam = ""
			}
			player.CurrentTeam = ""
			player.Hammer = "upcoming"
			player.SellingPrice = float64(0)
			player.CreatedAt = time.Now()
			player.UpdatedAt = time.Now()

			// Save player to database
			_, err := db.Collection(constants.PlayerCollection).InsertOne(ctx, player)
			if err != nil {
				logger.Error("failed to save player", zap.Any(constants.Err, err))
				// Try to cleanup the match document if player creation fails
				if isIPLAuction == "true" {
					_, err = db.Collection(constants.MatchCollection).DeleteOne(ctx, bson.M{"_id": match.Id})
					if err != nil {
						logger.Error("failed to delete match document for player", zap.Any(constants.Err, err), zap.Any("match", match.Id))
					}
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save player"})
				return
			}

			match = models.Match{}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Players saved successfully",
			"players": len(players),
		})
	}
}
