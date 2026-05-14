package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func New() (*sql.DB, error) {
	connStr := "postgres://chat:chat@localhost:5432/chat?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}