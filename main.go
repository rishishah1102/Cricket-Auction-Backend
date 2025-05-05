package main

import (
	"auction-backend/database"
	"auction-backend/routes"
	"auction-backend/utils"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, err := utils.ConfigLogger()
		if err != nil {
			zap.Must(zap.NewProduction()).Error(err.Error())
			return
		}

		origin := c.Request.Header.Get("Origin")
		logger.Info("cors" + origin + c.Request.Method)
		c.Header("Access-Control-Allow-Origin", origin) // CORS
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

func main() {
	// loading env
	godotenv.Load()

	// Reading logger
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return
	}

	// mongo config
	var mongoConfig utils.MongoConfig
	mongoConfig.MongoUri = os.Getenv("MONGODB_URI")
	mongoConfig.Database = os.Getenv("DATABASE_NAME")

	// gin instance
	router := gin.Default()
	router.Use(CORSMiddleware())

	// New Server
	server := &http.Server{
		Addr:    "localhost:" + os.Getenv("PORT"),
		Handler: router,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// user_routes
	routes.EndPoints(router)

	// Running the server in a goroutine
	go func() {
		logger.Info("The server is running on http://localhost:" + os.Getenv("PORT"))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("unable to start the server", zap.Error(err))
		}
	}()

	// Connecting with database
	err = database.ConnectDB(logger, mongoConfig)
	if err != nil {
		logger.Error("unable to connect with database", zap.Error(err))
		return
	}
	defer database.DisconnectDB(logger)

	// Waiting for the interrupt signal
	<-stop

	// Shutting down the server gracefully
	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error("error in gracefull shutdown of server", zap.Error(err))
	}
	logger.Info("Server stopped")
}
