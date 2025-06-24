package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Guide_Requests struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Datetime         time.Time          `bson:"datetime"`
	UserEmail        string             `bson:"useremail"`
	GuideEmail       string             `bson:"guideemail"`
	Msg              string             `bson:"msg"`
	Status           string             `bson:"status"`
	Rejection_Reason string             `bson:"rejection_reason"`
}

func CanUserCreateNewRequest(useremail string) (bool, error) {
	coll := GetCollection("guide_requests")
	tenHoursAgo := time.Now().Add(-5 * time.Hour)
	filter := bson.M{
		"useremail": useremail,
		"status": bson.M{
			"$in": []string{"pending", "accepted"},
		},
		"datetime": bson.M{
			"$gte": tenHoursAgo,
		},
	}

	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Println("CanUserCreateNewRequest Error ->", err)
		return false, err
	}

	if count > 0 {
		return false, nil
	}
	return true, nil
}

func InsertGuideRequestData(mydata Guide_Requests) (*mongo.InsertOneResult, error) {
	collection := GetCollection("guide_requests")

	_, err := RejectOldPendingRequests(mydata.UserEmail, 5)
	if err != nil {
		log.Println("Error rejecting old requests:", err)
		return nil, err
	}

	canCreate, err := CanUserCreateNewRequest(mydata.UserEmail)
	if err != nil {
		log.Println("Error checking for active requests:", err)
		return nil, err
	}

	if !canCreate {
		return nil, fmt.Errorf("You already have an active request within last 5 hours. Please wait or cancel existing request.")
	}

	if mydata.ID.IsZero() {
		mydata.ID = primitive.NewObjectID()
	}
	mydata.Datetime = time.Now()

	res, err := collection.InsertOne(context.TODO(), mydata)
	if err != nil {
		log.Println("Error inserting new request:", err)
		return nil, err
	}

	return res, nil
}

func GetGuideBookingRequestData(guideemail string) (Guide_Requests, error) {
	collection := GetCollection("guide_requests")
	var req Guide_Requests
	err := collection.FindOne(context.TODO(), bson.M{"guideemail": guideemail, "status": "pending"}).Decode(&req)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No guide found for email:", guideemail)
			return Guide_Requests{}, nil
		}
		log.Println("Error retrieving guide request data:", err)
		return Guide_Requests{}, err
	}

	return req, nil

}

func RejectApprovalRequest(useremail string) (bool, error) {
	coll := GetCollection("guide_requests")
	filter := bson.M{"useremail": useremail}
	query := bson.M{"$set": bson.M{"status": "rejected"}}
	res, err := coll.UpdateOne(context.TODO(), filter, query)
	if err != nil {
		log.Println("\nError -> ", err)
		return false, nil
	}
	fmt.Println(res)
	return true, nil
}

func AcceptApprovalRequest(useremail string) (bool, error) {
	coll := GetCollection("guide_requests")
	filter := bson.M{"useremail": useremail}
	update := bson.M{"$set": bson.M{"status": "accepted"}}
	opts := options.FindOneAndUpdate().SetSort(bson.M{"datetime": -1})
	res := coll.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if res.Err() != nil {
		log.Println("\nError -> ", res.Err())
		return false, res.Err()
	}
	fmt.Println("Updated latest document of user:", useremail)
	return true, nil
}

func UpdateRejectionReason(useremail string, reason string) (bool, error) {
	coll := GetCollection("guide_requests")
	filter := bson.M{"useremail": useremail}
	query := bson.M{"$set": bson.M{"rejection_reason": reason}}
	res, err := coll.UpdateOne(context.TODO(), filter, query)
	if err != nil {
		log.Println("\nError -> ", err)
		return false, nil
	}
	fmt.Println(res)
	return true, nil
}

func RejectOldPendingRequests(useremail string, hours int) (bool, error) {
	coll := GetCollection("guide_requests")
	timeThreshold := time.Now().Add(-time.Duration(hours) * time.Hour)

	filter := bson.M{
		"useremail": useremail,
		"status":    "pending",
		"datetime":  bson.M{"$lte": timeThreshold},
	}

	update := bson.M{
		"$set": bson.M{
			"status":           "rejected",
			"rejection_reason": "Your request was automatically rejected as the guide did not respond within the expected time.",
		},
	}

	res, err := coll.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		log.Printf("RejectOldPendingRequests Error for user: %s -> %v\n", useremail, err)
		return false, err
	}

	if res.ModifiedCount == 0 {
		log.Printf("No old pending requests to reject for user: %s\n", useremail)
		return false, nil
	}

	log.Printf("Rejected %d old pending request(s) for user: %s\n", res.ModifiedCount, useremail)
	return true, nil
}

func GetPendingRequestUserEmails() ([]string, error) {
	coll := GetCollection("guide_requests")
	timeThreshold := time.Now().Add(-5 * time.Hour)
	filter := bson.M{
		"status":   "pending",
		"datetime": bson.M{"$lte": timeThreshold},
	}
	emails, err := coll.Distinct(context.TODO(), "useremail", filter)
	if err != nil {
		log.Printf("Error fetching pending request emails: %v\n", err)
		return nil, err
	}
	emailList := make([]string, 0, len(emails))
	for _, email := range emails {
		if str, ok := email.(string); ok {
			emailList = append(emailList, str)
		}
	}
	return emailList, nil
}
