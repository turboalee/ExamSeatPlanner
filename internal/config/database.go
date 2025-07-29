package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDBClient(lc fx.Lifecycle, config *MongoDBConfig) (*MongoDBClient, *mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(config.URI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
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
	db := client.Database("exam_seat_planner")
	return &MongoDBClient{Client: client, Database: db}, db, nil
}

func UniqueCMSIndex(collection *mongo.Collection) {
	indexmodel := mongo.IndexModel{
		Keys:    bson.M{"cms_id": 1},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, indexmodel)
	if err != nil {
		log.Fatal("Failed to create unique index of CMS ID:", err)
	}

	log.Println("Unique Index on CMS ID created successfully")
}

func (c *MongoDBClient) GetCollection(collectionName string) *mongo.Collection {
	return c.Client.Database("exam_seat_planner").Collection(collectionName)
}
