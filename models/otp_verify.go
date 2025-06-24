package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Otp_Verify struct {
	ID    string `bson:"_id,omitempty"`
	Email string `bson:"email"`
	Otp   int    `bson:"otp"`
	DT    string `bson:"dt"`
}

func InsertOTP_Verify(otp Otp_Verify) (*mongo.InsertOneResult, error) {
	collection := GetCollection("otp_verify")
	result, err := collection.InsertOne(context.TODO(), otp)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetOTPByEmail(email string) (int, error) {
	collection := GetCollection("otp_verify")
	var result Otp_Verify
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Otp, nil
}
