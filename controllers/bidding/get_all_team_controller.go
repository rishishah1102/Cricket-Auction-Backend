package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type response struct {
	Team_Purse    float64 `json:"team_purse"`
	Batter        int     `json:"batter"`
	Bowler        int     `json:"bowler"`
	All_Rounder   int     `json:"all_rounder"`
	Wicket_Keeper int     `json:"wicket_keeper"`
	Overseas      int     `json:"overseas"`
	Team_Name     string  `json:"team_name"`
	Team_Image    string  `json:"team_image"`
}

func GetAllTeamsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request struct {
				AuctionID string `json:"auction_id"`
			}

			teams []models.Team
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind get all teams request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		filter := bson.M{"auction_id": request.AuctionID}
		cursor, err := db.Collection(constants.TeamCollection).Find(ctx, filter)
		if err != nil {
			logger.Error("failed to fetch teams", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &teams); err != nil {
			logger.Error("failed to decode teams", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while decoding"})
			return
		}

		teamResp := make([]response, len(teams))
		for i, team := range teams {
			var (
				batter        = 0
				bowler        = 0
				all_rounder   = 0
				wicket_keeper = 0
				overseas      = 0
				purse         = constants.TeamPurse
			)

			for _, player_id := range team.Squad {
				var player models.Player

				err := db.Collection(constants.PlayerCollection).FindOne(ctx, bson.M{"_id": player_id}).Decode(&player)
				if err != nil {
					logger.Error("failed to fetch player", zap.Any(constants.Err, err))
				}

				switch player.Role {
				case "Batter":
					batter++
				case "Bowler":
					bowler++
				case "All-Rounder":
					all_rounder++
				case "Wicket-Keeper":
					wicket_keeper++
				}

				if player.Country != "India" {
					overseas++
				}

				purse -= player.SellingPrice
			}
			
			teamResp[i] = response{
				Team_Purse:    purse,
				Batter:        batter,
				Bowler:        bowler,
				All_Rounder:   all_rounder,
				Wicket_Keeper: wicket_keeper,
				Overseas:      overseas,
				Team_Name:     team.TeamName,
				Team_Image:    team.TeamImage,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Teams data fetched successfully",
			"teams":   teamResp,
		})
	}
}
