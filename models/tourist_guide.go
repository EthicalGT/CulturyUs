package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Tourist_Guide struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Fullname         string             `bson:"fullname"`
	Email            string             `bson:"email"`
	DOB              string             `bson:"dob"`
	Addr             string             `bson:"addr"`
	ContactNo        string             `bson:"contactno"`
	Profilepic       string             `bson:"profilepic"`
	Experience       string             `bson:"experience"`
	Languages        string             `bson:"languages"`
	Preffered_States string             `bson:"preffered_states"`
	Charges          string             `bson:"charges"`
	Identity_Proof   string             `bson:"indentity_proof"`
	Bio              string             `bson:"bio"`
	Active           bool               `bson:"active"`
}

func InsertGuide(guide Tourist_Guide) (*mongo.InsertOneResult, error) {
	coll := GetCollection("tourist_guide")
	if guide.ID.IsZero() {
		guide.ID = primitive.NewObjectID()
	}

	result, err := coll.InsertOne(context.TODO(), guide)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func RetrieveAllGuideData(email string) ([]Tourist_Guide, error) {
	coll := GetCollection("tourist_guide")
	var guides []Tourist_Guide

	cursor, err := coll.Find(context.TODO(), bson.M{"email": email})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var guide Tourist_Guide
		if err := cursor.Decode(&guide); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		guides = append(guides, guide)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return guides, nil
}
func RetrieveGuideData(email string) (Tourist_Guide, error) {
	coll := GetCollection("tourist_guide")
	var guide Tourist_Guide

	err := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&guide)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No guide found for email:", email)
			return Tourist_Guide{}, nil
		}
		log.Println("Error retrieving guide data:", err)
		return Tourist_Guide{}, err
	}

	return guide, nil
}

func RetrieveAllGuides(state string) ([]Tourist_Guide, error) {
	coll := GetCollection("tourist_guide")
	var guides []Tourist_Guide
	filter := bson.M{"preffered_states": bson.M{"$in": []string{state}}}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var guide Tourist_Guide
		if err := cursor.Decode(&guide); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		guides = append(guides, guide)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return guides, nil
}

func CheckStatewiseGuide(state string) (bool, error) {
	coll := GetCollection("tourist_guide")
	cursor, err := coll.Find(context.TODO(), bson.M{"preffered_states": state})
	if err != nil {
		return false, err
	}
	defer cursor.Close(context.TODO())
	return true, nil
}

func CheckIfGuide(email string) (bool, error) {
	coll := GetCollection("tourist_guide")
	cursor, err := coll.Find(context.TODO(), bson.M{"email": email})
	if err != nil {
		return false, nil
	}
	defer cursor.Close(context.TODO())
	return true, nil
}
