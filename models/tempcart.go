package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartItem struct {
	PName  string `bson:"pname"`
	PImg   string `bson:"pimg"`
	PPrice int    `bson:"pprice"`
	PQty   int    `bson:"pqty"`
}

type TempCart struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CartID    string             `bson:"cart_id"`
	Items     []CartItem         `bson:"items"`
	CreatedAt time.Time          `bson:"created_at"`
}

func getTempCartCollection() *mongo.Collection {
	return GetCollection("tempcart")
}

func SaveOrUpdateTempCart(ctx context.Context, cartID string, items []CartItem) error {
	coll := getTempCartCollection()
	filter := bson.M{"cart_id": cartID}
	update := bson.M{
		"$set": bson.M{
			"items":      items,
			"created_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func GetTempCart(ctx context.Context, cartID string) (*TempCart, error) {
	coll := getTempCartCollection()
	var cart TempCart
	err := coll.FindOne(ctx, bson.M{"cart_id": cartID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		return nil, nil // no cart found
	}
	return &cart, err
}

func DeleteTempCart(ctx context.Context, cartID string) error {
	coll := getTempCartCollection()
	_, err := coll.DeleteOne(ctx, bson.M{"cart_id": cartID})
	return err
}
