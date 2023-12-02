package utils

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"time"
)

func MongoDBConnection() (client *mongo.Client, database *mongo.Database, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			log.Print("[mongoDB] ", evt.Command)
		},
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")).SetMonitor(cmdMonitor))
	log.Printf("Initialize mongoDB connection... | url: %s", os.Getenv("MONGO_URL"))

	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("PING: %v", err)
		return
	}

	return client, client.Database(os.Getenv("MONGODB_NAME")), nil
}
