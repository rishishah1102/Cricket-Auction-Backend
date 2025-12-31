package router

import (
	auth "cric-auction-monolith/controllers/auth"
	auction "cric-auction-monolith/controllers/auction"
	profile "cric-auction-monolith/controllers/profile"
	players "cric-auction-monolith/controllers/player"
	bidding "cric-auction-monolith/controllers/bidding"
	"cric-auction-monolith/pkg/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// NewGinRouter create a new gin router for each micro service
func NewGinRouter(logger *zap.Logger, db *mongo.Database) (router *gin.Engine) {
	router = gin.Default()

	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Auction Server",
		})
	})

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", auth.RegisterController(logger, db))

		authGroup.POST("/login", auth.LoginController(logger, db))

		authGroup.POST("/rotp", auth.RegisterOtpController(logger, db))

		authGroup.POST("/lotp", auth.LoginOtpController(logger, db))
	}

	api := router.Group("/api/v1")
	api.Use(middlewares.VerifyToken(logger))
	profileGroup := api.Group("/profile")
	{
		profileGroup.POST("/save", profile.SaveProfileController(logger, db))
		profileGroup.GET("/get", profile.GetProfileController(logger, db))
	}

	auctionGroup := api.Group("/auction")
	{
		auctionGroup.GET("/all", auction.GetAllAuctionsController(logger, db))

		auctionGroup.POST("/get", auction.GetAuctionController(logger, db))

		auctionGroup.POST("/create", auction.CreateAuctionController(logger, db))

		auctionGroup.POST("/join", auction.JoinAuctionController(logger, db))

		auctionGroup.PATCH("/update", auction.UpdateAuctionController(logger, db))

		auctionGroup.PATCH("/team", auction.UpdateTeamController(logger, db))

		auctionGroup.POST("/team/all", auction.GetAllTeamsController(logger, db))

		auctionGroup.POST("/team", auction.CreateTeamController(logger, db))

		auctionGroup.DELETE("/team", auction.DeleteTeamController(logger, db))
	}

	playersGroup := api.Group("/players")
	{
		playersGroup.POST("/get", players.GetAllPlayersController(logger, db))

		playersGroup.POST("/save", players.SavePlayerController(logger, db))

		playersGroup.PATCH("/update", players.UpdatePlayerController(logger, db))

		playersGroup.DELETE("/delete", players.DeletePlayerController(logger, db))

		playersGroup.POST("/squad", players.SquadsController(logger, db))
	}

	biddingGroup := api.Group("/bidding")
	{
		biddingGroup.POST("/teams/all", bidding.GetAllTeamsController(logger, db))
	}

	return router
}
