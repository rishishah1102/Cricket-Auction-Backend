package home

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// fetchAllAuctions fetches all the auction for the given email
func fetchAllAuctions(email any) (auctions []schemas.Auction, err error) {
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.L().Error("Failed to configure logger", zap.Error(err))
		return auctions, err
	}

	// Fetch the auctions collection
	auctionsCollection, err := models.AuctionCollection(logger)
	if err != nil {
		logger.Error("Unable to get auctions collection", zap.Error(err))
		return auctions, err
	}

	// Find all auctions where the email is in the joinedBy array and sort by createdAt in descending order
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // -1 for descending order

	// FIX: Use $elemMatch to check if the email exists inside the joinedBy array
	cursor, err := auctionsCollection.Find(context.Background(), bson.M{
		"joinedBy": bson.M{"$elemMatch": bson.M{"email": email}},
	}, findOptions)
	if err != nil {
		logger.Error("Error in fetching the data", zap.Error(err))
		return auctions, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &auctions); err != nil {
		logger.Error("Error decoding auctions", zap.Error(err))
		return auctions, err
	}

	return auctions, nil
}


// getUserDetails fetches the user's details based on email
func getUserDetails(email string, logger *zap.Logger) (schemas.User, error) {
	var user schemas.User

	// Get users collection
	usersCollection, err := models.UsersCollection(logger)
	if err != nil {
		logger.Error("Unable to get users collection", zap.Error(err))
		return user, err
	}

	// Define the filter
	filter := bson.M{"email": email}

	// Fetch the user details
	err = usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		logger.Error("User not found", zap.Error(err))
		return user, err
	}

	return user, nil
}
