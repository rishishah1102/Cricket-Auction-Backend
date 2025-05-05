package profile

import (
	"auction-backend/models"
	"auction-backend/schemas"
	"auction-backend/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// ProfileController saves the user profile.
func ProfileController(c *gin.Context) {
	logger, err := utils.ConfigLogger()
	if err != nil {
		zap.L().Error("Failed to configure logger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Fetch email from token
	email, exists := c.Get("email")
	if !exists {
		logger.Error("Email not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found in context"})
		return
	}

	var request schemas.User
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Unable to bind the request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	usersCollection, err := models.UsersCollection(logger)
	if err != nil {
		logger.Error("Unable to get users collection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	filter := bson.M{"email": email}
	update := bson.M{
		"firstName": request.FirstName, 
		"lastName": request.LastName,
		"imgUrl": request.ImgUrl,
		"role": request.Role,
		"battingHand": request.BattingHand,
		"battingOrder": request.BattingOrder,
		"battingStyle": request.BattingStyle,
		"bowlingArm": request.BowlingArm,
		"bowlingType": request.BowlingType,
	}

	var updatedUser schemas.User
	err = usersCollection.FindOneAndUpdate(context.Background(), filter, bson.M{"$set": update}).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Error("No user found with provided email", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"message": "No user found with the provided email"})
		} else {
			logger.Error("Error in fetching the data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in fetching the data", "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": updatedUser})
}