package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
)

var client *mongo.Client

func Connect() {
	config := config.GetConfig()

	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI(config.GetString("database.uri")))
	if err != nil {
		log.Fatalf("Error while creating MongoDB client %+v", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed while connecting to MongoDB %+v", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed while pinging to MongoDB %+v", err)
	}

	log.Println("Successfully connected to MongoDB!")
}

func Client() *mongo.Client {
	return client
}
