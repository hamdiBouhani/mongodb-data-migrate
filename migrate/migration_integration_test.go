package migrate

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDB         = "testM_db"
	testCollection = "testM_coll"
)

func cleanup(db *mongo.Client) {
	colls, err := db.Database(testDB).ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}

	for _, collection := range colls {
		if _, err := db.Database(testDB).Collection(collection).Indexes().DropAll(context.Background()); err != nil {
			panic(err)
		}

		if err := db.Database(testDB).Collection(collection).Drop(context.Background()); err != nil {
			panic(err)
		}
	}
}

var client *mongo.Client

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer cleanup(client)
	os.Exit(m.Run())
}

func TestSetGetVersion(t *testing.T) {
	defer cleanup(client)

	migrate := NewMigrate(testDB, client)
	if err := migrate.SetVersion(1, "hello"); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	version, description, err := migrate.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if version != 1 || description != "hello" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}

	if err := migrate.SetVersion(2, "world"); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	version, description, err = migrate.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if version != 2 || description != "world" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}

	if err := migrate.SetVersion(1, "hello"); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	version, description, err = migrate.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if version != 1 || description != "hello" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}
}

func TestVersionBeforeSet(t *testing.T) {
	defer cleanup(client)
	migrate := NewMigrate(testDB, client)
	version, _, err := migrate.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if version != 0 {
		t.Errorf("Unexpected version: %v", err)
		return
	}
}

func TestUpMigrations(t *testing.T) {
	defer cleanup(client)
	migrate := NewMigrate(testDB, client,
		Migration{Version: 1, Description: "hello", Up: func(db *mongo.Client) error {
			_collection := db.Database(testDB).Collection(testCollection)
			_, err := _collection.InsertOne(context.Background(), bson.M{"hello": "world"})
			if err != nil {
				return err
			}
			return nil
		}},
		Migration{Version: 2, Description: "world", Up: func(db *mongo.Client) error {
			//return db.C(testCollection).EnsureIndex(mgo.Index{Name: "test_idx", Key: []string{"hello"}})

			_collection := db.Database(testDB).Collection(testCollection)
			_, err := _collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.M{
					"hello": 1, // index in ascending order
				}, Options: nil,
			})
			if err != nil {
				return err
			}

			return nil
		}},
	)
	if err := migrate.Up(AllAvailable); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	version, description, err := migrate.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if version != 2 || description != "world" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}

	doc := bson.M{}
	_collection := client.Database(testDB).Collection(testCollection)
	err = _collection.FindOne(context.Background(), bson.M{"hello": "world"}).Decode(&doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if doc["hello"].(string) != "world" {
		t.Errorf("Unexpected data")
		return
	}
}
