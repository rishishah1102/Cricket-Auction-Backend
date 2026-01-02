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

func UpdatePlayerController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			models.Player
			TeamID primitive.ObjectID `json:"current_team_id,omitempty"`
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if request.Player.Id.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID required"})
			return
		}

		// Get current player state
		var currentPlayer models.Player
		err := db.Collection(constants.PlayerCollection).FindOne(ctx, bson.M{"_id": request.Player.Id}).Decode(&currentPlayer)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		if err != nil {
			logger.Error("failed to fetch player", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		request.Player.UpdatedAt = time.Now()

		// Check if team assignment changed
		stateChanged := currentPlayer.Hammer != request.Player.Hammer || currentPlayer.CurrentTeam != request.Player.CurrentTeam

		if stateChanged {
			err = updatePlayerWithTeam(ctx, db, request.Player, request.TeamID, &currentPlayer)
		} else {
			err = updatePlayer(ctx, db, request.Player)
		}

		if err != nil {
			logger.Error("failed to update player", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Player updated successfully",
			"player":  request.Player,
		})
	}
}

func updatePlayerWithTeam(ctx context.Context, db *mongo.Database, player models.Player, teamID primitive.ObjectID, oldPlayer *models.Player) error {
	session, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// Update player with bid info
		player.Bids = []models.Bids{{
			TeamName: player.CurrentTeam,
			Bid:      player.SellingPrice,
		}}

		if err := updatePlayer(sc, db, player); err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Handle team changes
		var update bson.M
		if player.CurrentTeam != oldPlayer.CurrentTeam && oldPlayer.CurrentTeam != "" {
			// Remove from old team
			var oldTeam models.Team
			if err := db.Collection(constants.TeamCollection).FindOne(sc, bson.M{"squad": player.Id}).Decode(&oldTeam); err != nil {
				session.AbortTransaction(sc)
				return err
			}

			if _, err := db.Collection(constants.TeamCollection).UpdateOne(sc, bson.M{"_id": oldTeam.ID}, bson.M{"$pull": bson.M{"squad": player.Id}}); err != nil {
				session.AbortTransaction(sc)
				return err
			}
			update = bson.M{"$addToSet": bson.M{"squad": player.Id}} // Add to new team
		}

		// Add/remove from team based on hammer state
		if shouldAddToTeam(oldPlayer.Hammer, player.Hammer) {
			update = bson.M{"$addToSet": bson.M{"squad": player.Id}}
		} else if shouldRemoveFromTeam(oldPlayer.Hammer, player.Hammer) {
			update = bson.M{"$pull": bson.M{"squad": player.Id}}
		}

		if update != nil || teamID != primitive.NilObjectID {
			if _, err := db.Collection(constants.TeamCollection).UpdateOne(sc, bson.M{"_id": teamID}, update); err != nil {
				session.AbortTransaction(sc)
				return err
			}
		}

		return session.CommitTransaction(sc)
	})
}

func updatePlayer(ctx context.Context, db *mongo.Database, player models.Player) error {
	_, err := db.Collection(constants.PlayerCollection).UpdateOne(ctx,
		bson.M{"_id": player.Id},
		bson.M{"$set": player})
	return err
}

func shouldAddToTeam(oldState, newState string) bool {
	return (oldState == "upcoming" || oldState == "unsold") && newState == "sold"
}

func shouldRemoveFromTeam(oldState, newState string) bool {
	return oldState == "sold" && (newState == "unsold" || newState == "upcoming")
}
