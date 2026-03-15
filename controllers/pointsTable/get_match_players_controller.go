package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type PlayerWithMatchData struct {
	models.Player
	MatchData *models.Match `bson:"-" json:"match_data,omitempty"`
}

func GetMatchPlayersController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			AuctionID primitive.ObjectID `json:"auction_id" binding:"required"`
			IPLTeam1  string             `json:"ipl_team1" binding:"required"`
			IPLTeam2  string             `json:"ipl_team2" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// 1. Fetch players from both IPL teams in one query
		cursor, err := db.Collection(constants.PlayerCollection).Find(ctx, bson.M{
			"auction_id": request.AuctionID,
			"ipl_team":   bson.M{"$in": []string{request.IPLTeam1, request.IPLTeam2}},
		})
		if err != nil {
			logger.Error("failed to fetch players", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer cursor.Close(ctx)

		var players []models.Player
		if err = cursor.All(ctx, &players); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
			return
		}

		// 2. Collect all match IDs
		matchIDs := make([]primitive.ObjectID, 0, len(players))
		for _, p := range players {
			if !p.Match.IsZero() {
				matchIDs = append(matchIDs, p.Match)
			}
		}

		// 3. Batch fetch all match documents in one query
		matchMap := make(map[primitive.ObjectID]models.Match, len(matchIDs))
		if len(matchIDs) > 0 {
			mCursor, err := db.Collection(constants.MatchCollection).Find(ctx,
				bson.M{"_id": bson.M{"$in": matchIDs}},
			)
			if err == nil {
				var matches []models.Match
				if mCursor.All(ctx, &matches) == nil {
					for _, m := range matches {
						matchMap[m.Id] = m
					}
				}
				mCursor.Close(ctx)
			}
		}

		// 4. Join in memory
		result := make([]PlayerWithMatchData, 0, len(players))
		for _, p := range players {
			pm := PlayerWithMatchData{Player: p}
			if !p.Match.IsZero() {
				if match, ok := matchMap[p.Match]; ok {
					matchCopy := match
					pm.MatchData = &matchCopy
				}
			}
			result = append(result, pm)
		}

		// 5. Sort by earned points descending, then benched points descending
		sort.Slice(result, func(i, j int) bool {
			ei, ej := 0, 0
			bi, bj := 0, 0
			if result[i].MatchData != nil {
				ei = result[i].MatchData.EarnedPoints
				bi = result[i].MatchData.BenchedPoints
			}
			if result[j].MatchData != nil {
				ej = result[j].MatchData.EarnedPoints
				bj = result[j].MatchData.BenchedPoints
			}
			if ei != ej {
				return ei > ej
			}
			return bi > bj
		})

		c.JSON(http.StatusOK, gin.H{
			"message": "Players fetched successfully",
			"players": result,
		})
	}
}
