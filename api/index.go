package handler

import (
	"auction-backend/database"
	"auction-backend/routes"
	"auction-backend/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	Router *gin.Engine
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allowing multiple origins
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://192.168.29.239:3000",
			// TODO
		}

		for _, o := range allowedOrigins {
			if origin == o {
				c.Header("Access-Control-Allow-Origin", origin) // CORS
				break
			}
		}

		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func init() {
	gin.SetMode(gin.ReleaseMode)

	// Reading yaml file
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error("failed to initialize custom logger", zap.Error(err))
		return
	}

	// mongo config
	var mongoConfig utils.MongoConfig
	mongoConfig.MongoUri = os.Getenv("MONGODB_URI")
	mongoConfig.Database = os.Getenv("DATABASE_NAME")

	// gin instance
	router := gin.Default()
	Router = router

	// cors
	router.Use(CORSMiddleware())

	// Connecting with database
	err = database.ConnectDB(logger, mongoConfig)
	if err != nil {
		logger.Error("unable to connect with mongodb database", zap.Error(err))
		return
	}
	// defer database.DisconnectDB(logger)

	//user_routes
	routes.EndPoints(router)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// serverless
	Router.ServeHTTP(w, r)
}
