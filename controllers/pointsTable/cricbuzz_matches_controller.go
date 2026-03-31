package controllers

import (
	"net/http"
	"os"

	"cric-auction-monolith/services/cricbuzz"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func CricbuzzMatchesController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	client := cricbuzz.NewClient(
		os.Getenv("CRICBUZZ_API_KEY"),
		os.Getenv("CRICBUZZ_API_HOST"),
	)

	return func(c *gin.Context) {
		resp, err := client.FetchRecentMatches()
		if err != nil {
			logger.Error("failed to fetch recent matches from Cricbuzz", zap.Error(err))
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch matches from Cricbuzz"})
			return
		}

		matches := cricbuzz.FilterIPLMatches(resp)
		if matches == nil {
			matches = []cricbuzz.SimplifiedMatch{}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "IPL matches fetched successfully",
			"matches": matches,
		})
	}
}
