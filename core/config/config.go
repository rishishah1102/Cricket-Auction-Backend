package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	DbName   string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		MongoURI: os.Getenv("MONGODB_URI"),
		DbName:   os.Getenv("DATABASE_NAME"),
	}

	return cfg
}
