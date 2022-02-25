package storage

import (
	"context"
	"github.com/explabs/ad-ctf-paas-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func connect() (*mongo.Client, error) {
	credential := options.Credential{
		Username: "admin",
		Password: os.Getenv("ADMIN_PASS"),
	}
	clientOpts := options.Client().ApplyURI("mongodb://mongo:27017").
		SetAuth(credential)
	return mongo.Connect(context.TODO(), clientOpts)

}

func scoreboard() (*mongo.Collection, error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	coll := client.Database("ad").Collection("scoreboard")
	return coll, nil
}
func services() *mongo.Collection {
	client, err := connect()
	if err != nil {
		return nil
	}
	return client.Database("ad").Collection("services")
}

func UpdateScore(score models.Score) (*mongo.UpdateResult, error) {
	coll, err := scoreboard()
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"name": score.Name}
	update := bson.M{
		"$set": score,
	}
	opts := options.Update().SetUpsert(true)
	return coll.UpdateOne(context.Background(), filter, update, opts)
}
