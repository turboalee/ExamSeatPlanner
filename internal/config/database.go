package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

type MongoDBConfig struct {
	URI string
}

func NewMongoDBConfig() *MongoDBConfig {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("DB uri not set")
	}
	return &MongoDBConfig{URI: uri}
}

type MongoDBClient struct {
	Client *mongo.Client
}

func NewMongoDBClient(lc fx.Lifecycle, config *MongoDBConfig) (*MongoDBClient, error) {
	clientOptions := options.Client().ApplyURI(config.URI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Errorf("Failed to connect to MongoDB")
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Errorf("Failed to ping MongoDB")
	}

	log.Println("Connected to MongoDB")

	lc.Append(fx.Hook{
		OnStart: func(Startctx context.Context) error {
			log.Println("MongoDB connection verified on startup")
			return nil
		},
		OnStop: func(Stopctx context.Context) error {
			log.Println("Closing MongoDB connection ...")
			return client.Disconnect(Stopctx)
		},
	})
	return &MongoDBClient{Client: client}, nil
}

func (c *MongoDBClient) GetCollection(collectionName string) *mongo.Collection {
	return c.Client.Database("exam_seat_planner").Collection(collectionName)
}

var MongoModule = fx.Module("mongo",
	fx.Provide(NewMongoDBConfig, NewMongoDBClient))
