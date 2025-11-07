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
	"go.uber.org/zap"
)

type auctionAPIRequest struct {
	AuctionID primitive.ObjectID `json:"auction_id"`
}

type auctionUserAPIResp struct {
	models.Auction
	UserNames []userTeam `json:"user_names"`
}

type userTeam struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func GetAuctionController(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			request   auctionAPIRequest
			response  auctionUserAPIResp
			userNames []userTeam
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), constants.DBTimeout)
		defer cancel()

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Error("failed to bind create team request", zap.Any(constants.Err, err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var auction models.Auction
		filter := bson.M{"_id": request.AuctionID}
		err := db.Collection(constants.AuctionCollection).FindOne(ctx, filter).Decode(&auction)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logger.Warn("no auction found", zap.Error(err), zap.Any("auction_id", request.AuctionID))
				c.JSON(http.StatusNotFound, gin.H{"error": "Auction not found or you are not authorized to update it"})
				return
			}
			logger.Error("failed to find auction in database", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find auction"})
			return
		}

		cursor, err := db.Collection(constants.UserCollection).Find(ctx, bson.M{"email": bson.M{"$in": auction.JoinedBy}})
		if err != nil {
			logger.Error("failed to fetch user names from database", zap.Any(constants.Err, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user names from database"})
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var user models.User
			if err := cursor.Decode(&user); err != nil {
				logger.Error("failed to decode user name", zap.Any(constants.Err, err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user name"})
				return
			}
			userTeam := userTeam{
				Email: user.Email,
				Name:  user.FirstName + " " + user.LastName,
			}
			userNames = append(userNames, userTeam)
		}

		response.ID = auction.ID
		response.AuctionName = auction.AuctionName
		response.AuctionImage = auction.AuctionImage
		response.CreatedBy = auction.CreatedBy
		response.AuctionDate = auction.AuctionDate
		response.IsIPLAuction = auction.IsIPLAuction
		response.CreatedAt = auction.CreatedAt
		response.UpdatedAt = auction.UpdatedAt
		response.JoinedBy = append(response.JoinedBy, auction.JoinedBy...)
		response.UserNames = append(response.UserNames, userNames...)

		c.JSON(http.StatusOK, gin.H{
			"message": "Auction fetched successfully",
			"auction": response,
		})
	}
}
