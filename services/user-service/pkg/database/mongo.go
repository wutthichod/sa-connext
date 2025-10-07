package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoURI = ""
)

type MongoStore struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDB(ctx context.Context) *MongoStore {
	databaseName := "user-db"

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(databaseName)

	log.Println("MongoDB migration completed")

	return &MongoStore{
		client:   client,
		database: db,
	}
}

func (m *MongoStore) DB() *mongo.Database {
	return m.database
}

func (m *MongoStore) Client() *mongo.Client {
	return m.client
}

func (m *MongoStore) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
