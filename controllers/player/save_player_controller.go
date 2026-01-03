package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func SavePlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		defer func() {
			logger.Info("SavePlayerController completed", zap.Duration("duration", time.Since(startTime)))
		}()

		var req struct {
			Auction_Id primitive.ObjectID `json:"auction_id"`
			Players    []models.Player    `json:"players"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("failed to bind save player request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		isIPLAuction := c.Query("isIPLAuction") == "true"

		// Use MongoDB session for transaction
		session, err := db.Client().StartSession()
		if err != nil {
			logger.Error("failed to start session", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer session.EndSession(ctx)

		// Execute in transaction
		err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
			if err := session.StartTransaction(); err != nil {
				return err
			}

			// Batch insert logic
			if err := batchInsertPlayers(sc, logger, db, req.Players, req.Auction_Id, isIPLAuction); err != nil {
				session.AbortTransaction(sc)
				return err
			}

			return session.CommitTransaction(sc)
		})

		if err != nil {
			logger.Error("transaction failed", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save players",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Players saved successfully",
			"players": len(req.Players),
		})
	}
}

func batchInsertPlayers(ctx context.Context, logger *zap.Logger, db *mongo.Database, players []models.Player, auctionID primitive.ObjectID, isIPLAuction bool) error {
	now := time.Now()
	playerDocs := make([]interface{}, 0, len(players))
	matchDocs := make([]interface{}, 0, len(players))

	for i := range players {
		player := &players[i]

		// // TODO: Remove this simulated error
		// if player.PlayerNumber == 50 {
		// 	logger.Error("simulated error")
		// 	return errors.New("simulated error for player number 50")
		// }

		player.Id = primitive.NewObjectID()
		player.AuctionId = auctionID
		player.CreatedAt = now
		player.UpdatedAt = now
		player.CurrentTeam = ""
		player.Hammer = "upcoming"
		player.SellingPrice = 0

		if len(player.PrevTeam) == 0 {
			player.PrevTeam = ""
		}

		if isIPLAuction {
			matchID := primitive.NewObjectID()
			player.Match = matchID

			if player.PrevFantasyPoints == 0 {
				player.PrevFantasyPoints = 0
			}

			match := models.Match{
				Id:                matchID,
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
			matchDocs = append(matchDocs, match)
		}

		playerDocs = append(playerDocs, player)
	}

	// Batch insert matches first (if IPL auction)
	if isIPLAuction && len(matchDocs) > 0 {
		_, err := db.Collection(constants.MatchCollection).InsertMany(ctx, matchDocs)
		if err != nil {
			logger.Error("failed to batch insert matches", zap.Any(constants.Err, err))
			return errors.New("failed to insert matches")
		}
	}

	// Batch insert players
	_, err := db.Collection(constants.PlayerCollection).InsertMany(ctx, playerDocs)
	if err != nil {
		logger.Error("failed to batch insert players", zap.Any(constants.Err, err))
		return errors.New("failed to insert players")
	}

	return nil
}
