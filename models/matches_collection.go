package models

import (
	"auction-backend/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func MatchesCollection(logger *zap.Logger) (*mongo.Collection, error) {
	if database.DB == nil {
		logger.Warn(errDatabase.Error())
		return nil, errDatabase
	}

	matchesCollection := database.DB.Collection("matches")
	if matchesCollection == nil {
		logger.Error(errMatchesCollections.Error())
		return nil, errMatchesCollections
	}

	return matchesCollection, nil
}
