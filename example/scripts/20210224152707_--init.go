package scripts

import (
	"context"
	"mongodb-data-migrate/example/internal"
	"mongodb-data-migrate/migrate"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	migrate.Register(func(db *mongo.Client) error {
		document := struct {
			ID       string `bson:"_id,omitempty"`
			FullName string `bson:"full_name,omitempty"`
		}{
			ID:       primitive.NewObjectID().Hex(),
			FullName: "test",
		}
		_collection := db.Database(internal.DB).Collection("users")
		_, err := _collection.InsertOne(context.Background(), document)
		if err != nil {
			return err
		}
		return nil
	}, func(db *mongo.Client) error {
		return nil
	})
}
