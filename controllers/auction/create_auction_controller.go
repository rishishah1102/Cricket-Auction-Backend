package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func CreateAuctionController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.Auction

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind create auction request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		email := c.GetString("email")
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		auctionDoc := bson.M{
			"auction_name":   request.AuctionName,
			"auction_image":  request.AuctionImage,
			"auction_date":   request.AuctionDate,
			"created_by":     email,
			"is_ipl_auction": request.IsIPLAuction,
			"joined_by":      []string{},
			"created_at":     time.Now(),
			"updated_at":     time.Now(),
		}

		res, err := db.Collection(constants.AuctionCollection).InsertOne(ctx, auctionDoc)
		if err != nil {
			logger.Error("failed to create auction", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}

		request.ID = res.InsertedID.(primitive.ObjectID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Auction created successfully",
			"auction": request,
		})
	}
}
