// Package migrate allows to perform versioned migrations in your MongoDB.
package migrate

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// migrate allows to perform versioned migrations in your MongoDB.
type versionRecord struct {
	Version     uint64
	Description string `bson:",omitempty"`
	Timestamp   time.Time
}

const defaultMigrationsCollection = "migrations"

// AllAvailable used in "Up" or "Down" methods to run all available migrations.
const AllAvailable = -1

// Migrate is type for performing migrations in provided database.
// Database versioned using dedicated collection.
// Each migration applying ("up" and "down") adds new document to collection.
// This document consists migration version, migration description and timestamp.
// Current database version determined as version in latest added document (biggest "_id") from collection mentioned above.
type Migrate struct {
	dbName               string
	db                   *mongo.Client
	migrations           []Migration
	migrationsCollection string
	logger               *log.Logger
}

func NewMigrate(dbName string, db *mongo.Client, migrations ...Migration) *Migrate {
	internalMigrations := make([]Migration, len(migrations))
	copy(internalMigrations, migrations)
	return &Migrate{
		dbName:               dbName,
		db:                   db,
		migrations:           internalMigrations,
		migrationsCollection: defaultMigrationsCollection,
	}
}

// SetMigrationsCollection replaces name of collection for storing migration information.
// By default it is "migrations".
func (m *Migrate) SetMigrationsCollection(name string) {
	m.migrationsCollection = name
}

// SetLogger set a logger
func (m *Migrate) SetLogger(l *log.Logger) {
	m.logger = l
}

func (m *Migrate) isCollectionExist(name string) (bool, error) {
	colls, err := m.db.Database(m.dbName).ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return false, err
	}
	for _, v := range colls {
		if name == v {
			return true, nil
		}
	}
	return false, nil
}

func (m *Migrate) createCollectionIfNotExist(name string) error {
	exist, err := m.isCollectionExist(name)
	if err != nil {
		return err
	}
	if !exist {
		quickstartDatabase := m.db.Database(m.dbName)
		_ = quickstartDatabase.Collection(name)
	}
	return nil
}

// Version returns current database version and comment.
func (m *Migrate) Version() (uint64, string, error) {
	if err := m.createCollectionIfNotExist(m.migrationsCollection); err != nil {
		return 0, "", err
	}

	var recs []versionRecord
	// find record with greatest id (assuming it`s latest also)
	ctx := context.Background()
	findOptions := options.Find()
	// Sort by `_id` field descending
	findOptions.SetSort(bson.D{{"_id", -1}})
	findOptions.SetLimit(1)

	res, err := m.db.Database(m.dbName).Collection("migrations").Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return 0, "", err
	}

	err = res.All(ctx, &recs)
	if err != nil {
		return 0, "", err
	}
	if len(recs) > 0 {
		return recs[0].Version, recs[0].Description, nil
	}
	return 0, "", nil
}

// SetVersion forcibly changes database version to provided.
func (m *Migrate) SetVersion(version uint64, description string) error {
	_collection := m.db.Database(m.dbName).Collection(m.migrationsCollection)
	_, err := _collection.InsertOne(context.Background(), versionRecord{
		Version:     version,
		Timestamp:   time.Now().UTC(),
		Description: description,
	})
	if err != nil {
		return err
	}
	return nil
}

// Up performs "up" migrations to latest available version.
// If n<=0 all "up" migrations with newer versions will be performed.
// If n>0 only n migrations with newer version will be performed.
func (m *Migrate) Up(n int) error {
	currentVersion, _, err := m.Version()
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := 0, 0; i < len(m.migrations) && p < n; i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Up == nil {
			continue
		}
		p++
		if err := migration.Up(m.db); err != nil {
			return err
		}
		if m.logger != nil {
			m.logger.Printf("MIGRATED UP: %d %s\n", migration.Version, migration.Description)
		}
		if err := m.SetVersion(migration.Version, migration.Description); err != nil {
			return err
		}
	}
	return nil
}

// Down performs "down" migration to oldest available version.
// If n<=0 all "down" migrations with older version will be performed.
// If n>0 only n migrations with older version will be performed.
func (m *Migrate) Down(n int) error {
	currentVersion, _, err := m.Version()
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := len(m.migrations)-1, 0; i >= 0 && p < n; i-- {
		migration := m.migrations[i]
		if migration.Version > currentVersion || migration.Down == nil {
			continue
		}
		p++
		if err := migration.Down(m.db); err != nil {
			return err
		}

		var prevMigration Migration
		if i == 0 {
			prevMigration = Migration{Version: 0}
		} else {
			prevMigration = m.migrations[i-1]
		}
		if m.logger != nil {
			m.logger.Printf("MIGRATED DOWN: %d %s\n", migration.Version, migration.Description)
		}
		if err := m.SetVersion(prevMigration.Version, prevMigration.Description); err != nil {
			return err
		}
	}
	return nil
}
