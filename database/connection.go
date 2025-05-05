package database

import (
	"auction-backend/utils"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	Client *mongo.Client
	DB     *mongo.Database
)

func ConnectDB(logger *zap.Logger, mongoConfig utils.MongoConfig) error {
	// Connection with mongo Atlas
	clientOptions := options.Client().ApplyURI(mongoConfig.MongoUri)
	cl, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logger.Error("failed to connect to MongoDB", zap.Error(err))
		return err
	}
	Client = cl

	// Checks whether the connection with MongoDb is established or not
	err = Client.Ping(context.Background(), nil)
	if err != nil {
		logger.Error("failed to ping MongoDB", zap.Error(err))
		return err
	}

	// Creating database
	db := Client.Database(mongoConfig.Database)
	if db == nil {
		logger.Error("unable to make database connection", zap.Error(errors.New("database name is empty")))
		return err
	}
	DB = db

	logger.Info("connected to database")
	return nil
}
