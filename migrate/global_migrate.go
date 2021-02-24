package migrate

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

var globalMigrate = NewMigrate("", nil)

func extractVersionDescription(name string) (uint64, string, error) {
	base := filepath.Base(name)

	if ext := filepath.Ext(base); ext != ".go" {
		return 0, "", fmt.Errorf("can not extract version from %q", base)
	}

	idx := strings.IndexByte(base, '_')
	if idx == -1 {
		return 0, "", fmt.Errorf("can not extract version from %q", base)
	}

	version, err := strconv.ParseUint(base[:idx], 10, 64)
	if err != nil {
		return 0, "", err
	}

	description := base[idx+1 : len(base)-len(".go")]

	return version, description, nil
}

func internalRegister(up, down MigrationFunc, skip int) error {
	_, file, _, _ := runtime.Caller(skip)
	version, description, err := extractVersionDescription(file)
	if err != nil {
		return err
	}
	if hasVersion(globalMigrate.migrations, version) {
		return fmt.Errorf("migration with version %v already registered", version)
	}
	globalMigrate.migrations = append(globalMigrate.migrations, Migration{
		Version:     version,
		Description: description,
		Up:          up,
		Down:        down,
	})
	return nil
}

// Register performs migration registration.
// Use case of this function:
//
// - Create a file called like "1_setup_indexes.go" ("<version>_<comment>.go").
//
// - Use the following template inside:
//
// 	package migrations
// 	import (
// 		"github.com/globalsign/mgo"
//		"github.com/xakep666/mongo-migrate"
// 	)
//
// 	func init() {
//		migrate.Register(func (db *mgo.Database) error {
//			return db.C(collection).EnsureIndex(index)
//		}, func (db *mgo.Database) error {
//			return db.C(collection).DropIndexName(index.Name)
//		})
//	 }
func Register(up, down MigrationFunc) error {
	fmt.Printf("up: %v down: %v \n", up, down)
	return internalRegister(up, down, 2)
}

// MustRegister acts like Register but panics on errors.
func MustRegister(up, down MigrationFunc) {
	if err := internalRegister(up, down, 2); err != nil {
		panic(err)
	}
}

// RegisteredMigrations returns all registered migrations.
func RegisteredMigrations() []Migration {
	ret := make([]Migration, len(globalMigrate.migrations))
	copy(ret, globalMigrate.migrations)
	return ret
}

// SetDatabase sets database for global migrate.
func SetDatabase(name string, db *mongo.Client) {
	globalMigrate.dbName = name
	globalMigrate.db = db
}

// SetMigrationsCollection changes default collection name for migrations history.
func SetMigrationsCollection(name string) {
	globalMigrate.SetMigrationsCollection(name)
}

// SetLogger set a logger
func SetLogger(l *log.Logger) {
	globalMigrate.SetLogger(l)
}

// Version returns current database version.
func Version() (uint64, string, error) {
	return globalMigrate.Version()
}

// Up performs "up" migration using registered migrations.
// Detailed description available in Migrate.Up().
func Up(n int) error {
	return globalMigrate.Up(n)
}

// Down performs "down" migration using registered migrations.
// Detailed description available in Migrate.Down().
func Down(n int) error {
	return globalMigrate.Down(n)
}

func GetMigrations() []Migration {
	return globalMigrate.migrations
}
