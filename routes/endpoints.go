package routes

import (
	"auction-backend/controllers/add_all_player"
	"auction-backend/controllers/auction"
	"auction-backend/controllers/auth"
	"auction-backend/controllers/home"
	"auction-backend/controllers/profile"
	"auction-backend/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func EndPoints(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Auction backend")
	})

	// Making a group for authenticated endpoints
	authUsers := router.Group("/auth")
	authUsers.Use(middlewares.VerifyToken)

	// SIGNUP || METHOD POST
	router.POST("/register", auth.RegisterController)

	// LOGIN || METHOD POST
	router.POST("/login", auth.LoginController)

	// OTP PAGE TO SAVE USER || METHOD POST
	router.POST("/rotp", auth.RegisterOtpController)

	// OTP PAGE TO LOGIN USER || METHOD POST
	router.POST("/lotp", auth.LoginOtpController)

	// Player ADDING || METHOD POST
	authUsers.POST("/addplayer", add_all_player.AddPlayerController)

	// HOME || METHOD GET
	authUsers.GET("/user", home.UserController)

	// HOME || METHOD GET
	authUsers.GET("/home", home.HomeController)

	// HOME || METHOD POST
	authUsers.POST("/createAuction", home.CreateAuctionController)

	// HOME || METHOD POST
	authUsers.POST("/joinAuction", home.JoinAuctionController)

	// AUCTION || METHOD GET
	authUsers.GET("/auction", auction.FetchAllAuctionsController)

	// AUCTION || METHOD POST
	authUsers.POST("/saveAuction", auction.UpdateAuctionAndTeamsController)

	// PROFILE || Profile POST
	authUsers.POST("/profile", profile.ProfileController)
}
