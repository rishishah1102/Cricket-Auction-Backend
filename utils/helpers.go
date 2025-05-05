package utils

type MongoConfig struct {
    MongoUri string `yaml:"mongoUri"`
    Database string `yaml:"dataBaseName"`
}