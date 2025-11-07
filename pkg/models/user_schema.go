package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the struct for auction users
type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Email        string             `bson:"email" json:"email"`
	Mobile       string             `bson:"mobile" json:"mobile"`
	FirstName    string             `bson:"first_name" json:"first_name"`
	LastName     string             `bson:"last_name" json:"last_name"`
	ImageURL     string             `bson:"image_url" json:"image_url"`
	Role         string             `bson:"role" json:"role"`
	BattingHand  string             `bson:"batting_hand" json:"batting_hand"`
	BattingOrder string             `bson:"batting_order" json:"batting_order"`
	BattingStyle string             `bson:"batting_style" json:"batting_style"`
	BowlingArm   string             `bson:"bowling_arm" json:"bowling_arm"`
	BowlingType  string             `bson:"bowling_type" json:"bowling_type"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
