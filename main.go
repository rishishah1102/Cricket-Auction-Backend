package main

import (
	"context"
	"cric-auction-monolith/core/config"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/core/database"
	"cric-auction-monolith/core/logger"
	"cric-auction-monolith/core/router"
	"cric-auction-monolith/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	cfg := config.LoadConfig()
	logger := logger.Get()

	client, err := database.NewMongoClient(ctx, cfg.MongoURI, logger)
	if err != nil {
		logger.Error("failed to create mongo client", zap.Any(constants.Err, err))
	}
	defer database.DisconnectMongoClient(ctx, client, logger)
	db := client.Database(cfg.DbName)

	router := router.NewGinRouter(logger, db)

	utils.StartServer(ctx, router, logger)
}
