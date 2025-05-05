package models

import (
	"auction-backend/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func TeamsCollection(logger *zap.Logger) (*mongo.Collection, error) {
	if database.DB == nil {
		logger.Warn(errDatabase.Error())
		return nil, errDatabase
	}

	teamsCollection := database.DB.Collection("teams")
	if teamsCollection == nil {
		logger.Error(errTeamssCollections.Error())
		return nil, errTeamssCollections
	}

	return teamsCollection, nil
}
