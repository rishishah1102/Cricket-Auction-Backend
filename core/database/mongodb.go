package database

import (
	"context"
	"cric-auction-monolith/core/constants"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// NewMongoClient connects the application with mongo database and creates new mongo client
func NewMongoClient(ctx context.Context, uri string, logger *zap.Logger) (client *mongo.Client, err error) {
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		mongoErr := errors.New("failed to connect MongoDB")
		logger.Error(mongoErr.Error(), zap.Any(constants.Err, err))
		return nil, mongoErr
	}

	// Ping the database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		pingErr := errors.New("failed to ping MongoDB")
		logger.Error(pingErr.Error(), zap.Any(constants.Err, err))
		return nil, pingErr
	}

	logger.Info("Successfully connected to database!!")

	return client, nil
}

// DisconnectMongoClient disconnects the application with mongo database
func DisconnectMongoClient(ctx context.Context, client *mongo.Client, logger *zap.Logger) {
	if err := client.Disconnect(ctx); err != nil {
		logger.Warn("failed to disconnect mongo client", zap.Any(constants.Err, err))
	}
}
