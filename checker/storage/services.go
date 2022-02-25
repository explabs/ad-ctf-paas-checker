package storage

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

type Service struct {
	Name   string `bson:"name"`
	Cost   int    `bson:"cost"`
	HP     int    `bson:"hp"`
	Flags  int    `bson:"flags"`
	Script string `bson:"script"`
}

var ctx = context.TODO()

func GetServices() (s []*Service, err error) {
	cur, err := services().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		var service Service
		err := cur.Decode(&service)
		if err != nil {
			return s, err
		}

		s = append(s, &service)
	}

	if err := cur.Err(); err != nil {
		return s, err
	}
	return s, nil
}

func UploadServices(s []*Service) {
	var si []interface{}
	for _, elem := range s {
		si = append(si, *elem)
	}
	services().DeleteMany(ctx, bson.M{})
	insertManyResults, err := services().InsertMany(ctx, si)
	if err != nil {
		log.Println(err)
	}
	log.Println(insertManyResults)
}
