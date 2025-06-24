package models

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PurchasedSkills struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	Email                string             `bson:"email"`
	SkillName            string             `bson:"skillname"`
	SkillPurchased       string             `bson:"skillpurchased"`
	PurchaseDateTime     string             `bson:"purchasedatetime"`
	SkillStatus          string             `bson:"skillstatus"`
	SkillCertificateID   string             `bson:"skillcertificateID"`
	SkillCertificatePath string             `bson:"skillcertificatepath"`
}

func InsertPurchasedSkillsData(ps PurchasedSkills) (*mongo.InsertOneResult, error) {
	cll := GetCollection("purchasedskills")
	if ps.ID.IsZero() {
		ps.ID = primitive.NewObjectID()
	}
	res, err := cll.InsertOne(context.TODO(), ps)
	if err != nil {
		log.Fatal("\nPurchasing skills Error -> ", err)
	}
	return res, nil
}

func RetrievePurchasedSkillData(skillname string) (*PurchasedSkills, error) {
	coll := GetCollection("purchasedskills")
	var data PurchasedSkills
	err := coll.FindOne(context.TODO(), bson.M{"skillname": skillname}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
func UpdateGeneratedCertificateInfo(skillname string, certID string, certPath string) (bool, error) {
	coll := GetCollection("purchasedskills")
	filter := bson.M{"skillname": skillname}
	update := bson.M{"$set": bson.M{"skillstatus": "completed", "skillcertificateID": certID, "skillcertificatepath": certPath}}
	res, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	fmt.Println(res.ModifiedCount)
	return true, nil
}

func CheckIfDataExists(skillname string) (bool, error) {
	coll := GetCollection("purchasedskills")
	var data PurchasedSkills
	err := coll.FindOne(context.TODO(), bson.M{"skillname": skillname}).Decode(&data)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func RetrieveAllPurchasedSkills(ctx context.Context, email string) ([]Skills, error) {
	cll := GetCollection("purchasedskills")
	var myskills []Skills
	filter := bson.M{"email": email}
	projection := bson.M{"skillstatus": "completed"}
	cr, err := cll.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cr.Close(ctx)
	for cr.Next(ctx) {
		var skill Skills
		if err := cr.Decode(&skill); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		myskills = append(myskills, skill)
	}
	if err := cr.Err(); err != nil {
		return nil, err
	}
	return myskills, nil
}
