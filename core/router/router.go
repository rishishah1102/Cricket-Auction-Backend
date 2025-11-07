package router

import (
	auth "cric-auction-monolith/controllers/auth"
	profile "cric-auction-monolith/controllers/profile"
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

	return router
}
