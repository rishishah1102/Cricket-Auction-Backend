package models

import (
	"auction-backend/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func PlayersCollection(logger *zap.Logger) (*mongo.Collection, error) {
	if database.DB == nil {
		logger.Warn(errDatabase.Error())
		return nil, errDatabase
	}

	playersCollection := database.DB.Collection("players")
	if playersCollection == nil {
		logger.Error(errPlayersCollections.Error())
		return nil, errPlayersCollections
	}

	return playersCollection, nil
}
