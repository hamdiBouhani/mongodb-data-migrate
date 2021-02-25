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
