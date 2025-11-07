package constants

import "time"

var (
	Err               = "err"
	DBTimeout         = 10 * time.Second
	MaxRetries        = 3
	EmailKey          = "email"
	UserCollection    = "users"
	AuctionCollection = "auctions"
	ProfileCollection = "profiles"
	PlayerCollection  = "players"
	TeamCollection    = "teams"
	MatchCollection   = "matches"
	OtpCollection     = "otps"
)
