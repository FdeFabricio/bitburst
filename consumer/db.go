package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	username   = os.Getenv("POSTGRES_USER")
	password   = os.Getenv("POSTGRES_PASSWORD")
	host       = os.Getenv("POSTGRES_HOST")
	port       = os.Getenv("POSTGRES_PORT")
	database   = os.Getenv("POSTGRES_DB")
	connString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
)

func InsertObject(ID int, ts time.Time) error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return extendError(err, "db: failed to open database")
	}

	err = db.Ping()
	if err != nil {
		return extendError(err, "db: failed to establish connection")
	}

	sqlStatement := `
		INSERT INTO object (id, last_seen) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET last_seen = EXCLUDED.last_seen;
	`

	_, err = db.Exec(sqlStatement, ID, ts)
	if err != nil {
		return extendError(err, "db: failed to insert record")
	}

	err = db.Close()
	if err != nil {
		return extendError(err, "db: failed to close connection")
	}

	return nil
}
