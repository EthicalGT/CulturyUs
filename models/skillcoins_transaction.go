package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Skillcoins_Transactions struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `bson:"email"`
	Payment_Status string             `bson:"payment_status"`
	Payment_Mode   string             `bson:"payment_mode"`
	Amount         string             `bson:"amount"`
	DateTime       string             `bson:"datetime"`
}

func InsertSkillcoinTransactionRec(skT Skillcoins_Transactions) (*mongo.InsertOneResult, error) {
	cll := GetCollection("skillcoins_transaction")
	if skT.ID.IsZero() {
		skT.ID = primitive.NewObjectID()
	}
	res, err := cll.InsertOne(context.TODO(), skT)
	if err != nil {
		log.Fatal("\nPurchasing skills Error -> ", err)
	}
	return res, nil
}
