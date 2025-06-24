package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var db *mongo.Database

func Connect() {
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	db = client.Database("culturyus")
}

func GetCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

func Disconnect() {
	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
