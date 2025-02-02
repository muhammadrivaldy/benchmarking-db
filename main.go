package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {

	purpose := os.Args[1]

	if purpose == "run-service" {
		runService()
	} else if purpose == "run-migration" {
		runMigration()
	}
}

func runService() {

	wg := new(sync.WaitGroup)

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			db := openDB()
			defer db.Close()

			for i := 0; i < 1000000; i++ {
				name := fmt.Sprintf("User_%d", i)
				address := fmt.Sprintf("Address_%d", i)
				_, err := db.Exec("INSERT INTO mst_user (name, address) VALUES ($1, $2)", name, address)
				if err != nil {
					log.Fatalf("Insert failed: %v", err)
				}
			}

			fmt.Println("✅ Inserted 1000000 users.")
		}(wg)
	}

	wg.Wait()
}

func openDB() *sql.DB {

	const host = "localhost"
	const port = 5432
	const user = "admin"
	const password = "admin"
	const dbname = "mydb"

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	fmt.Println("✅ Connected to PostgreSQL!")
	return db
}

func runMigration() {

	db := openDB()
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mydb",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations applied successfully!")
}
