package models

import (
	"auction-backend/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func UsersCollection(logger *zap.Logger) (*mongo.Collection, error) {
	if database.DB == nil {
		logger.Warn(errDatabase.Error())
		return nil, errDatabase
	}

	usersCollection := database.DB.Collection("users")
	if usersCollection == nil {
		logger.Error(errUsersCollections.Error())
		return nil, errUsersCollections
	}

	return usersCollection, nil
}
