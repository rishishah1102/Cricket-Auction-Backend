package controllers

import (
	"context"
	"cric-auction-monolith/core/constants"
	"cric-auction-monolith/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type auctionAPIResp struct {
	AuctionID    primitive.ObjectID `json:"id"`
	AuctionName  string             `json:"auction_name"`
	AuctionImage string             `json:"auction_image"`
	IsIPLAuction bool               `json:"is_ipl_auction"`
}

func GetAllAuctionsController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			auctions []models.Auction
			resp     []auctionAPIResp
			filter   bson.M
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		email := c.GetString("email")
		if email == "" {
			logger.Error("failed to fetch email from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
			return
		}

		auctionType := c.DefaultQuery("type", "all") // all | create | join

		switch auctionType {
		case "create":
			filter = bson.M{"created_by": email}
		case "join":
			filter = bson.M{"joined_by": email}
		case "all":
			filter = bson.M{
				"$or": []bson.M{
					{"created_by": email},
					{"joined_by": email},
				},
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query type"})
			return
		}
		findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

		cursor, err := db.Collection(constants.AuctionCollection).Find(ctx, filter, findOptions)
		if err != nil {
			logger.Error("failed to fetch auctions", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error from db"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &auctions); err != nil {
			logger.Error("failed to decode auctions", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while decoding"})
			return
		}

		for _, auction := range auctions {
			var auctionResp auctionAPIResp
			auctionResp.AuctionID = auction.ID
			auctionResp.AuctionName = auction.AuctionName
			auctionResp.AuctionImage = auction.AuctionImage
			auctionResp.IsIPLAuction = auction.IsIPLAuction
			resp = append(resp, auctionResp)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Auctions fetched successfully",
			"auctions": resp,
		})
	}
}
