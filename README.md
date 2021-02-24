# mongodb-data-migrate
# Versioned migrations for MongoDB

This package allows to perform versioned migrations on your MongoDB using [go.mongodb.org/mongo-driver](https://github.com/mongodb/mongo-go-driver).
It depends only on standard library and MongoDB Go Driver.
Inspired by [mongo-migrate](https://github.com/eminetto/mongo-migrate).


## Usage
### Migrations in files.

```go
package scripts

import (
	"context"
	"mongodb-data-migrate/internal"
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
```

* Import it in your application.
```go
import (
    ...
    migrate "mongodb-data-migrate/migrate"
    _ "path/to/migrations_package" // database migrations
    ...
)
```

* Run migrations.
```shell
go run main.go migrate --init
go run main.go migrate --up
go run main.go migrate --down
```
* example.
example [main.go](https://github.com/hamdiBouhani/mongodb-data-migrate/tree/main/example).