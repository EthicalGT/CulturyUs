package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Artpieces struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Datetime      time.Time          `bson:"datetime"`
	Email         string             `bson:"useremail"`
	Artpiece_name string             `bson:"artpeice_name"`
	Artpiece_desc string             `bson:"artpiece_desc"`
	Artpiece_img  string             `bson:"artpiece_img"`
	Price         string             `bson:"price"`
}

func InsertArtpeices(product Artpieces) (*mongo.InsertOneResult, error) {
	coll := GetCollection("artpieces")
	if product.ID.IsZero() {
		product.ID = primitive.NewObjectID()
	}
	res, err := coll.InsertOne(context.TODO(), product)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RetrieveArtpiecesData() ([]Artpieces, error) {
	cll := GetCollection("artpieces")
	var artpieces []Artpieces
	cr, err := cll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cr.Close(context.TODO())
	for cr.Next(context.TODO()) {
		var product Artpieces
		if err := cr.Decode(&product); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		artpieces = append(artpieces, product)
	}
	if err := cr.Err(); err != nil {
		return nil, err
	}

	return artpieces, nil
}
