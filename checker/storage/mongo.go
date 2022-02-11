package storage

import (
	"context"
	"github.com/explabs/ad-ctf-paas-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func collection() (*mongo.Collection, error) {
	credential := options.Credential{
		Username: "admin",
		Password: "admin",
	}
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017").
		SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return nil, err
	}
	coll := client.Database("ad").Collection("scoreboard")
	return coll, nil
}

func UpdateScore(score models.Score) (*mongo.UpdateResult, error) {
	coll, err := collection()
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"name": score.Name}
	update := bson.M{
		"$set": bson.M{
			"round":         score.Round,
			"services":      score.Services,
			"last_services": score.LastServices,
			"sla":           score.SLA,
			"last_sla":      score.LastSLA,
			"score":         score.Score,
			"last_score":    score.LastScore,
		},
	}
	opts := options.Update().SetUpsert(true)
	return coll.UpdateOne(context.Background(), filter, update, opts)
}
