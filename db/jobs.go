package db

import "go.mongodb.org/mongo-driver/mongo"

func JobsCollection() *mongo.Collection {
	return Client.Database("jobqueue").Collection("jobs")
}
