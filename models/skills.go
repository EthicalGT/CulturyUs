package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Skills struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Email            string             `bson:"email"`
	Uploader         string             `bson:"uploader"`
	Skillname        string             `bson:"skillname"`
	Skilltype        string             `bson:"skilltype"`
	Skilldesc        string             `bson:"skilldesc"`
	Mediatype        string             `bson:"mediatype"`
	Mediapath        string             `bson:"mediapath"`
	LanguageInstruct string             `bson:"languageinstruct"`
	UploadDateTime   string             `bson:"datetime"`
}

func InsertSkills(skills Skills) (*mongo.InsertOneResult, error) {
	coll := GetCollection("skills")
	if skills.ID.IsZero() {
		skills.ID = primitive.NewObjectID()
	}

	result, err := coll.InsertOne(context.TODO(), skills)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CheckSkillAvailablity(skilltype string) (*Skills, error) {
	cll := GetCollection("skills")
	var skills Skills
	err := cll.FindOne(context.TODO(), bson.M{"skilltype": skilltype}).Decode(&skills)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &skills, nil
}

func RetrieveSkillData(skilltype string) ([]Skills, error) {
	cll := GetCollection("skills")
	var skills []Skills
	cr, err := cll.Find(context.TODO(), bson.M{"skilltype": skilltype})
	if err != nil {
		return nil, err
	}
	defer cr.Close(context.TODO())
	for cr.Next(context.TODO()) {
		var skill Skills
		if err := cr.Decode(&skill); err != nil {
			log.Println("Decode error:", err)
			continue
		}
		skills = append(skills, skill)
	}
	if err := cr.Err(); err != nil {
		return nil, err
	}

	return skills, nil
}
