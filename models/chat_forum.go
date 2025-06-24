package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Chat_forum struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Datetime   time.Time          `bson:"datetime"`
	Email      string             `bson:"useremail"`
	Msg        string             `bson:"msg"`
	ProfilePic string             `bson:"profilepic"`
}

func InsertChatInForum(chat Chat_forum) (*mongo.InsertOneResult, error) {
	coll := GetCollection("chat_forum")
	if chat.ID.IsZero() {
		chat.ID = primitive.NewObjectID()
	}
	res, err := coll.InsertOne(context.TODO(), chat)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetForumMessages(limit int64) ([]Chat_forum, error) {
	coll := GetCollection("chat_forum")

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"datetime", 1}})
	findOptions.SetLimit(limit)

	cursor, err := coll.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var messages []Chat_forum
	if err = cursor.All(context.TODO(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
