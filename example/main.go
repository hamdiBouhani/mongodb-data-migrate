package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mongodb-data-migrate/example/internal"
	_ "mongodb-data-migrate/example/scripts"
	"mongodb-data-migrate/migrate"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var commandMigrate *cobra.Command
var (
	argDsn       string
	description  string
	up           bool
	down         bool
	newMigration bool
)

func init() {
	commandMigrate = &cobra.Command{
		Use:   "migrate",
		Short: "Connect to the storage and begin serving requests.",
		Long:  ``,
		Run: func(commandMigrate *cobra.Command, args []string) {
			if err := migrateDB(commandMigrate, args); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		},
	}
	commandMigrate.Flags().StringVar(&argDsn, "dsn", "mongodb://localhost:27017", "db url")
	commandMigrate.Flags().StringVar(&description, "desc", "latest", "migration description")
	commandMigrate.Flags().BoolVar(&up, "up", false, "migrate up")
	commandMigrate.Flags().BoolVar(&down, "down", false, "migrate down")
	commandMigrate.Flags().BoolVar(&newMigration, "new", false, "New migration")

}

func migrateDB(cmd *cobra.Command, args []string) error {
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(argDsn)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	migrate.SetDatabase(internal.DB, client)
	migrate.SetMigrationsCollection("migrations")
	migrate.SetLogger(log.New(os.Stdout, "INFO: ", 0))

	for index, v := range migrate.GetMigrations() {
		log.Printf("migration :%d\t description :%s\tmigrations version :%d\n", index, v.Description, v.Version)
	}

	if newMigration {

		fName := fmt.Sprintf("./scripts/%s_%s.go", time.Now().Format("20060102150405"), description)
		from, err := os.Open("./scripts/latest_migration.go")
		if err != nil {
			log.Fatal("Should be: new description-of-migration")
		}
		defer from.Close()

		to, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("New migration created: %s\n", fName)
	} else if up {
		fmt.Println("up")
		err := migrate.Up(migrate.AllAvailable)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if down {
		fmt.Println("down")
		err := migrate.Down(migrate.AllAvailable)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	return nil
}

func main() {

	rootCmd := &cobra.Command{
		Use: "mongo-bd-migrate",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(2)
		},
	}

	rootCmd.AddCommand(commandMigrate)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}
