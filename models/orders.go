package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Cart struct {
	ProductName string `bson:"product_name" json:"product_name"`
	Quantity    int    `bson:"quantity" json:"quantity"`
	Price       int    `bson:"price" json:"price"`
	PImg        string `bson:"pimg" json:"pimg"`
}

type CartSummary struct {
	ProductName string `bson:"product_name" json:"product_name"`
	Quantity    int    `bson:"quantity" json:"quantity"`
	Price       int    `bson:"price" json:"price"`
	PImg        string `bson:"pimg" json:"pimg"`
}

type Orders struct {
	ID       primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Email    string                 `bson:"email" json:"email"`
	Items    map[string]CartSummary `bson:"items" json:"items"`
	DateTime primitive.DateTime     `bson:"datetime" json:"datetime"`
	PayMode  string                 `bson:"paymode" json:"paymode"`
	Amount   float64                `bson:"amount" json:"amount"`
}

func InsertOrder(ctx context.Context, data *Orders) (*mongo.InsertOneResult, error) {
	collection := GetCollection("orders")

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}
	if data.DateTime == 0 {
		data.DateTime = primitive.NewDateTimeFromTime(time.Now())
	}

	return collection.InsertOne(ctx, data)
}

func ConvertCartToSummary(cart []Cart) map[string]CartSummary {
	cartMap := make(map[string]CartSummary)

	for _, item := range cart {
		if summary, exists := cartMap[item.ProductName]; exists {
			summary.Quantity += item.Quantity
			cartMap[item.ProductName] = summary
		} else {
			cartMap[item.ProductName] = CartSummary{
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       item.Price,
				PImg:        item.PImg,
			}
		}
	}

	return cartMap
}

func GetOrdersByEmail(ctx context.Context, email string) ([]Orders, error) {
	collection := GetCollection("orders")

	filter := bson.M{"email": email}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []Orders
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode orders: %w", err)
	}

	return orders, nil
}
