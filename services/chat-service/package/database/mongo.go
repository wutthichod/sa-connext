package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDB(ctx context.Context, mongoURI string) *MongoStore {
	databaseName := "chat-db"

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

func (m *MongoStore) RunMigrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Chat Collection
	chatCollection := m.database.Collection("chat")
	_, err := chatCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	// Messages Collection
	msgCollection := m.database.Collection("messages")
	_, err = msgCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}, {Key: "created_at", Value: 1}},
	})
	if err != nil {
		return err
	}

	// Participants Collection
	participantsCollection := m.database.Collection("participants")
	// Composite unique index: chat_id + user_id
	_, err = participantsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	log.Println("MongoDB migration completed successfully")
	return nil
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
