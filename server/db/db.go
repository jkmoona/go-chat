package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"os"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	dbURL := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", dbURL)

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
