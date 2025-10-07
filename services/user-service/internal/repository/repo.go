package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	Createuser(ctx context.Context, name string) error
}

type repository struct {
	db *mongo.Database
}

func NewRepo(db *mongo.Database) Repository {
	return &repository{db}
}

func (r *repository) Createuser(ctx context.Context, name string) error {

	data := bson.M{
		"name": name,
	}
	_, err := r.db.Collection("user").InsertOne(ctx, data)
	return err
}
