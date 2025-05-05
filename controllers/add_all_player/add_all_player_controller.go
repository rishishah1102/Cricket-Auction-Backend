package add_all_player

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// This route is for adding all the players into database after uploading and converting excel file
func AddPlayerController(ctx *gin.Context) {
	var (
		players = make([]schemas.Player, 250)
		match   schemas.Match
	)

	// Reading logger
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return
	}

	// Binding data from request body
	if err := ctx.ShouldBind(&players); err != nil {
		logger.Error("unable to bind the request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, player := range players {
		matchesCollection, err := models.MatchesCollection(logger)
		if err != nil {
			logger.Error("unable to get matches collection", zap.Error(err))
			return
		}

		matchData, err := matchesCollection.InsertOne(context.Background(), bson.M{
			"match1":            match.Match1,
			"match2":            match.Match2,
			"match3":            match.Match3,
			"match4":            match.Match4,
			"match5":            match.Match5,
			"match6":            match.Match6,
			"match7":            match.Match7,
			"match8":            match.Match8,
			"match9":            match.Match9,
			"match10":           match.Match10,
			"prevX1":            match.PrevX1,
			"currentX1":         match.CurrentX1,
			"nextX1":            match.NextX1,
			"earnedPoints":      match.EarnedPoints,
			"benchedPoints":     match.BenchedPoints,
			"totalPoints":       match.TotalPoints,
			"prevTotalPoints":   match.PrevTotalPoints,
			"prevEarnedPoints":  match.PrevEarnedPoints,
			"prevBenchedPoints": match.PrevBenchedPoints,
		})
		if err != nil {
			logger.Error("unable to create a matches instance", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// adding _id of match model to each player
		insertedId, ok := matchData.InsertedID.(primitive.ObjectID)
		if !ok {
			logger.Error("Failed to convert _id to ObjectID")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Cannot convert _id to ObjectID",
			})
			return
		}
		player.Match = insertedId

		// player collection
		playersCollection, err := models.PlayersCollection(logger)
		if err != nil {
			logger.Error("unable to get players collection", zap.Error(err))
			return
		}

		// inserting player into database
		_, err = playersCollection.InsertOne(context.Background(), bson.M{
			"playerNumber":         player.PlayerNumber,
			"playerName":           player.PlayerName,
			"country":              player.Country,
			"role":                 player.Role,
			"playingTeam":          player.PlayingTeam,
			"prevTeam":             player.PrevTeam,
			"currentTeam":          player.CurrentTeam,
			"basePrice":            player.BasePrice,
			"prevFantasyPoints":    player.PrevFantasyPoints,
			"currentFantasyPoints": player.CurrentFantasyPoints,
			"sellingPrice":         player.SellingPrice,
			"match":                player.Match,
			"hammer":               player.Hammer,
		})
		if err != nil {
			logger.Error("unable to create a players instance", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// sending response
	logger.Info("All players data inserted successfully")
	ctx.JSON(http.StatusOK, gin.H{
		"succes":  "Ok",
		"message": "Data inserted successfully",
	})
}
