package scripts

import (
	"context"
	"mongodb-data-migrate/example/internal"
	"mongodb-data-migrate/migrate"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	migrate.Register(func(db *mongo.Client) error {
		document := struct {
			ID    string `bson:"_id,omitempty"`
			model string `bson:"phone_model,omitempty"`
		}{
			ID:    primitive.NewObjectID().Hex(),
			model: "bmw",
		}
		_collection := db.Database(internal.DB).Collection("cars")
		_, err := _collection.InsertOne(context.Background(), document)
		if err != nil {
			return err
		}
		return nil
	}, func(db *mongo.Client) error {
		_collection := db.Database(internal.DB).Collection("cars")
		_, err := _collection.DeleteOne(context.Background(), bson.M{"phone_model": "bmw"})
		if err != nil {
			return err
		}
		return nil
	})
}
