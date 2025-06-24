package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Fullname   string             `bson:"fullname"`
	Email      string             `bson:"email"`
	DOB        string             `bson:"dob"`
	Addr       string             `bson:"addr"`
	ContactNo  string             `bson:"contactno"`
	PWD        string             `bson:"pwd"`
	Profilepic string             `bson:"profilepic"`
	Bio        string             `bson:"bio"`
	Status     bool               `bson:"status"`
	SkillCoins int32              `bson:"skillcoin"`
}

func InsertUsers(users Users) (*mongo.InsertOneResult, error) {
	collection := GetCollection("users")
	if users.ID.IsZero() {
		users.ID = primitive.NewObjectID()
	}

	result, err := collection.InsertOne(context.TODO(), users)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetUserByEmail(email string) (*Users, error) {
	collection := GetCollection("users")
	var users Users
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&users)
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func UpdateLoggedInUserStatus(email string) (bool, error) {
	collection := GetCollection("users")
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"status": true}}
	res, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	fmt.Println("Update Matches :", res.MatchedCount)
	return true, nil
}
func UpdateRestLoginStatus(email string) (bool, error) {
	collection := GetCollection("users")
	filter := bson.M{"email": bson.M{"$ne": email}}
	updation := bson.M{"$set": bson.M{"status": false}}
	res, err := collection.UpdateMany(context.TODO(), filter, updation)
	if err != nil {
		return false, err
	}
	fmt.Println("Modified Count: ", res.ModifiedCount)
	return true, nil
}
func GetCurrentUserInfo() (*Users, error) {
	collection := GetCollection("users")
	var user Users
	err := collection.FindOne(context.TODO(), bson.M{"status": true}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserInfo(email string, name string, addr string, contact string, bio string, profilepic string) (bool, error) {
	collection := GetCollection("users")
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"fullname": name, "addr": addr, "contactno": contact, "profilepic": profilepic, "bio": bio}}
	res, err := collection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	fmt.Println("User Info Updated.", res.MatchedCount)
	return true, nil

}

func UpdateSkillCoinsOnUpload() (bool, error) {
	coll := GetCollection("users")
	filter := bson.M{"status": true}
	query := bson.M{"$inc": bson.M{"skillcoin": 5}}
	res, err := coll.UpdateOne(context.TODO(), filter, query)
	if err != nil {
		return false, err
	}
	fmt.Println(res.UpsertedCount)
	return true, nil
}

func UpdateSkillCoinsOnPurchase(mycoin int) (bool, error) {
	coll := GetCollection("users")
	filter := bson.M{"status": true}
	query := bson.M{"$set": bson.M{"skillcoin": mycoin}}
	res, err := coll.UpdateOne(context.TODO(), filter, query)
	if err != nil {
		return false, err
	}
	fmt.Println(res.UpsertedCount)
	return true, nil
}
func UpdateSkillCoinsOnPayment(skillcoins int) (bool, error) {
	coll := GetCollection("users")
	filter := bson.M{"status": true}
	query := bson.M{"$inc": bson.M{"skillcoin": skillcoins}}
	res, err := coll.UpdateOne(context.TODO(), filter, query)
	if err != nil {
		return false, err
	}
	fmt.Println(res.UpsertedCount)
	return true, nil
}
