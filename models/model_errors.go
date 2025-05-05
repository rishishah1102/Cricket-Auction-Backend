package models

import "errors"

var (
	errDatabase            = errors.New("database client is nil")
	errAuctionsCollections = errors.New("failed to create auctions collection")
	errMatchesCollections  = errors.New("failed to create matches collection")
	errPlayersCollections  = errors.New("failed to create players collection")
	errUsersCollections    = errors.New("failed to create users collection")
	errTeamssCollections   = errors.New("failed to create teams collection")
)
