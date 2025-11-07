package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func UpdateAuctionController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request  models.Auction
			response models.Auction
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind update auction request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		email := c.GetString("email")
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		filter := bson.M{
			"_id":        request.ID,
			"created_by": email,
		}
		update := bson.M{
			"$set": bson.M{
				"auction_name":   request.AuctionName,
				"auction_image":  request.AuctionImage,
				"auction_date":   request.AuctionDate,
				"is_ipl_auction": request.IsIPLAuction,
				"updated_at":     time.Now(),
			},
		}

		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err := db.Collection(constants.AuctionCollection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&response)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logger.Warn("no auction found for update", zap.Any(constants.Err, err))
				c.JSON(http.StatusNotFound, gin.H{"error": "Auction not found or you are not authorized to update it"})
				return
			}
			logger.Error("failed to update auction in database", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Auction updated successfully",
			"auction": response,
		})
	}
}
