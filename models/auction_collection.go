package models

import (
	"auction-backend/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func AuctionCollection(logger *zap.Logger) (*mongo.Collection, error) {
	if database.DB == nil {
		logger.Warn(errDatabase.Error())
		return nil, errDatabase
	}

	auctionsCollection := database.DB.Collection("auctions")
	if auctionsCollection == nil {
		logger.Error(errAuctionsCollections.Error())
		return nil, errAuctionsCollections
	}

	return auctionsCollection, nil
}
