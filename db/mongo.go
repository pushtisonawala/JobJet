package db

import (
	"context"
	"jobjet/logger"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectMongo() {
	_ = godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		logger.Log.Error(" MONGO_URI not found in env")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		logger.Log.Error(" Mongo connection failed:", "error", err)
		os.Exit(1)
	}

	Client = client
	logger.Log.Info(" MongoDB connected")
}
