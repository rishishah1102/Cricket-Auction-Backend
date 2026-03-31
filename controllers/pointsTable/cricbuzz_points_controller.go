package controllers

import (
	"context"
	"net/http"
	"os"
	"strings"

	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"cric-auction-monolith/services/cricbuzz"
	"cric-auction-monolith/services/fantasy"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func CricbuzzPointsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	client := cricbuzz.NewClient(
		os.Getenv("CRICBUZZ_API_KEY"),
		os.Getenv("CRICBUZZ_API_HOST"),
	)

	return func(c *gin.Context) {
		var request struct {
			AuctionID      primitive.ObjectID `json:"auction_id" binding:"required"`
			CricbuzzMatchID int               `json:"cricbuzz_match_id" binding:"required"`
			IPLTeam1       string             `json:"ipl_team1" binding:"required"`
			IPLTeam2       string             `json:"ipl_team2" binding:"required"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// 1. Fetch scorecard from Cricbuzz.
		scorecard, err := client.FetchScorecard(request.CricbuzzMatchID)
		if err != nil {
			logger.Error("failed to fetch scorecard from Cricbuzz", zap.Error(err))
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch scorecard from Cricbuzz"})
			return
		}

		// 2. Calculate fantasy points from scorecard.
		fantasyPoints := fantasy.CalculateAllPoints(scorecard)

		// 3. Fetch DB players for both IPL teams in this auction.
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

		// 4. Fetch match documents for these players.
		matchIDs := make([]primitive.ObjectID, 0, len(players))
		for _, p := range players {
			if !p.Match.IsZero() {
				matchIDs = append(matchIDs, p.Match)
			}
		}

		matchMap := make(map[primitive.ObjectID]primitive.ObjectID) // player match ref -> match doc ID
		if len(matchIDs) > 0 {
			mCursor, err := db.Collection(constants.MatchCollection).Find(ctx,
				bson.M{"_id": bson.M{"$in": matchIDs}},
			)
			if err == nil {
				var matches []models.Match
				if mCursor.All(ctx, &matches) == nil {
					for _, m := range matches {
						matchMap[m.Id] = m.Id
					}
				}
				mCursor.Close(ctx)
			}
		}

		// 5. Build DB player list for name matching.
		dbPlayers := make([]fantasy.DBPlayer, len(players))
		for i, p := range players {
			matchID := ""
			if !p.Match.IsZero() {
				if _, ok := matchMap[p.Match]; ok {
					matchID = p.Match.Hex()
				}
			}
			dbPlayers[i] = fantasy.DBPlayer{
				PlayerName: p.PlayerName,
				IPLTeam:    p.IPLTeam,
				Role:       p.Role,
				MatchID:    matchID,
			}
		}

		// 6. Match Cricbuzz players to DB players and build response.
		type pointResult struct {
			MatchID      string                `json:"match_id"`
			PlayerName   string                `json:"player_name"`
			CricbuzzName string                `json:"cricbuzz_name"`
			Points       int                   `json:"points"`
			Breakdown    fantasy.PointBreakdown `json:"breakdown"`
			Confidence   float64               `json:"confidence"`
		}

		var results []pointResult
		matchedDB := make(map[int]bool)
		matchedCB := make(map[string]bool)

		// Determine which team each Cricbuzz player belongs to.
		teamForPlayer := buildTeamMap(scorecard)

		for cbName, pp := range fantasyPoints {
			cbTeam := teamForPlayer[cbName]
			match := fantasy.MatchPlayerToDB(cbName, cbTeam, dbPlayers)
			if match.DBIndex < 0 {
				continue
			}

			dbP := dbPlayers[match.DBIndex]
			if dbP.MatchID == "" {
				continue // no match doc in DB
			}

			// Apply duck rule: only for BAT, WK, AR roles.
			adjustedPoints := pp.Points
			role := strings.ToUpper(dbP.Role)
			isBowler := strings.Contains(role, "BOWL") && !strings.Contains(role, "ALL")
			if isBowler {
				// Remove duck penalty if it was applied.
				for _, d := range pp.Breakdown.Details {
					if strings.Contains(d, "Duck") {
						adjustedPoints += 2 // undo the -2
						break
					}
				}
			}

			results = append(results, pointResult{
				MatchID:      dbP.MatchID,
				PlayerName:   dbP.PlayerName,
				CricbuzzName: pp.CricbuzzName,
				Points:       adjustedPoints,
				Breakdown:    pp.Breakdown,
				Confidence:   match.Confidence,
			})

			matchedDB[match.DBIndex] = true
			matchedCB[cbName] = true
		}

		// Collect unmatched players.
		var unmatchedCB []string
		for cbName := range fantasyPoints {
			if !matchedCB[cbName] {
				unmatchedCB = append(unmatchedCB, cbName)
			}
		}

		var unmatchedDB []string
		for i, p := range dbPlayers {
			if !matchedDB[i] && p.MatchID != "" {
				unmatchedDB = append(unmatchedDB, p.PlayerName)
			}
		}

		if results == nil {
			results = []pointResult{}
		}
		if unmatchedCB == nil {
			unmatchedCB = []string{}
		}
		if unmatchedDB == nil {
			unmatchedDB = []string{}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":            "Points calculated successfully",
			"points":             results,
			"unmatched_cricbuzz": unmatchedCB,
			"unmatched_db":       unmatchedDB,
		})
	}
}

// buildTeamMap creates a map of player name -> team short name from the scorecard.
func buildTeamMap(sc *cricbuzz.ScorecardResponse) map[string]string {
	m := make(map[string]string)
	for _, inn := range sc.Scorecard {
		// Batsmen belong to the batting team.
		for _, bat := range inn.Batsmen {
			m[bat.Name] = inn.BatTeamSName
		}
		// Bowlers belong to the OTHER team. We need to figure that out.
		// The bowling team is the team that is NOT batting in this innings.
		// We can find it from other innings or leave it for now.
		// Since bowlers also appear as batsmen in the other innings,
		// they'll get their team from there.
	}
	return m
}
