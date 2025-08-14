package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"os"
)

type Database struct {
	db *sql.DB
}

const (
	DB_USER     = "DB_USER"
	DB_PASSWORD = "DB_PASSWORD"
	DB_HOST     = "DB_HOST"
	DB_PORT     = "DB_PORT"
	DB_NAME     = "DB_NAME"
)

func NewDatabase() (*Database, error) {
	username := os.Getenv(DB_USER)
	pwd := os.Getenv(DB_PASSWORD)
	host := os.Getenv(DB_HOST)
	port := os.Getenv(DB_PORT)
	dbName := os.Getenv(DB_NAME)

	db, err := sql.Open("postgres",
		"host="+host+" port="+port+" user="+username+" password="+pwd+" dbname="+dbName+" sslmode=disable",
	)

	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}
