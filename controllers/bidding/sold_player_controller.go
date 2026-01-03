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

type soldPlayerRequest struct {
	PlayerID     primitive.ObjectID `json:"player_id" binding:"required"`
	AuctionID    primitive.ObjectID `json:"auction_id" binding:"required"`
	TeamID       primitive.ObjectID `json:"team_id" binding:"required"`
	SellingPrice float64            `json:"selling_price" binding:"required,gt=0"`
	TeamName     string             `json:"team_name" binding:"required"`
}

func SoldPlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req soldPlayerRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("invalid request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := markPlayerAsSold(ctx, db, req); err != nil {
			logger.Error("failed to mark player as sold", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark player as sold"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Player marked as sold"})
	}
}

func markPlayerAsSold(ctx context.Context, db *mongo.Database, req soldPlayerRequest) error {
	session, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// Update player status
		playerUpdate := bson.M{
			"$set": bson.M{
				"hammer":        "sold",
				"current_team":  req.TeamName,
				"selling_price": req.SellingPrice,
			},
		}

		result, err := db.Collection(constants.PlayerCollection).UpdateOne(sc,
			bson.M{"_id": req.PlayerID, "auction_id": req.AuctionID},
			playerUpdate)

		if err != nil || result.MatchedCount == 0 {
			session.AbortTransaction(sc)
			return err
		}

		// Add player to team squad
		teamUpdate := bson.M{
			"$addToSet": bson.M{"squad": req.PlayerID},
		}

		_, err = db.Collection(constants.TeamCollection).UpdateOne(sc,
			bson.M{"_id": req.TeamID, "auction_id": req.AuctionID},
			teamUpdate)

		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		return session.CommitTransaction(sc)
	})
}
