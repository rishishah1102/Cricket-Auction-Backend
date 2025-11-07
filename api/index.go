package handler

import (
	"context"
	"cric-auction-monolith/core/config"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/core/database"
	"cric-auction-monolith/core/logger"
	"cric-auction-monolith/core/router"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	Router *gin.Engine
)

func init() {
	gin.SetMode(gin.ReleaseMode)

	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	logger := logger.Get()

	// mongo config
	var mongoConfig config.Config
	mongoConfig.MongoURI = os.Getenv("MONGODB_URI")
	mongoConfig.DbName = os.Getenv("DATABASE_NAME")

	client, err := database.NewMongoClient(ctx, mongoConfig.MongoURI, logger)
	if err != nil {
		logger.Error("failed to create mongo client", zap.Any(constants.Err, err))
	}
	db := client.Database(mongoConfig.DbName)

	router := router.NewGinRouter(logger, db)
	Router = router
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// serverless
	Router.ServeHTTP(w, r)
}
