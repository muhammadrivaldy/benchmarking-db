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

	wg.Add(2)

	n := 5000
	d := 50

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

				fmt.Printf("✅ PostgreSQL inserted %d users.\n", n)
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

				fmt.Printf("✅ MySQL inserted %d users.\n", n)
			}(childWG)
		}

		childWG.Wait()
		wg.Done()
	}(wg)

	wg.Wait()
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

	fmt.Println("✅ Connected to PostgreSQL!")
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

	fmt.Println("✅ Connected to MySQL!")
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
