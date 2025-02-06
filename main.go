package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
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

	n := 10
	d := 5
	iteration := 1

	for {

		wg.Add(2)

		go func(wg *sync.WaitGroup) {

			childWG := new(sync.WaitGroup)

			for i := 0; i < d; i++ {

				childWG.Add(1)

				go func(childWG *sync.WaitGroup) {
					defer childWG.Done()
					db := openPostgresDB()
					defer db.Close()

					for i := 0; i < n; i++ {
						name := fmt.Sprintf("User_%d", i)
						address := fmt.Sprintf("Address_%d", i)
						_, err := db.Exec("INSERT INTO mst_user (name, address) VALUES ($1, $2)", name, address)
						if err != nil {
							log.Fatalf("Insert failed: %v", err)
						}
					}
				}(childWG)
			}

			childWG.Wait()
			wg.Done()
		}(wg)

		go func(wg *sync.WaitGroup) {

			childWG := new(sync.WaitGroup)

			for i := 0; i < d; i++ {

				childWG.Add(1)

				go func(childWG *sync.WaitGroup) {
					defer childWG.Done()
					db := openMySQLDB()
					defer db.Close()

					for i := 0; i < n; i++ {
						name := fmt.Sprintf("User_%d", i)
						address := fmt.Sprintf("Address_%d", i)
						_, err := db.Exec("INSERT INTO mst_user (name, address) VALUES (?, ?)", name, address)
						if err != nil {
							log.Fatalf("Insert failed: %v", err)
						}
					}
				}(childWG)
			}

			childWG.Wait()
			wg.Done()
		}(wg)

		wg.Wait()

		fmt.Printf("✅ Iteration %d is done!\n", iteration)
		fmt.Printf("✅ %d records inserted to PostgreSQL\n", (n * d))
		fmt.Printf("✅ %d records inserted to MySQL\n", (n * d))

		n += 500
		d += 2
		iteration++
		if iteration > 20 {
			break
		}
	}
}

func openPostgresDB() *sql.DB {

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

	return db
}

func openMySQLDB() *sql.DB {

	const host = "localhost"
	const port = 3306
	const user = "admin"
	const password = "admin"
	const dbname = "mydb"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	return db
}

func runMigration() {
	migrationPostgresDB()
	migrationMySQLDB()
}

func migrationPostgresDB() {

	db := openPostgresDB()
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

	fmt.Println("✅ PostgreSQL migrations applied successfully!")
}

func migrationMySQLDB() {

	db := openMySQLDB()
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
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

	fmt.Println("✅ MySQL migrations applied successfully!")
}
