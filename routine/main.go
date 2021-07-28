package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var (
	username   = os.Getenv("POSTGRES_USER")
	password   = os.Getenv("POSTGRES_PASSWORD")
	host       = os.Getenv("POSTGRES_HOST")
	port       = os.Getenv("POSTGRES_PORT")
	database   = os.Getenv("POSTGRES_DB")
	connString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
)

func init() { log.SetLevel(log.DebugLevel) }

// this service works as a routine that performs a deletion queries every interval defined
// to remove entities older than 30 seconds. (default on .env file is every 3 seconds)
func main() {
	interval, err := time.ParseDuration(os.Getenv("ROUTINE_INTERVAL"))
	if err != nil {
		log.Fatal(extendError(err, "invalid ROUTINE_INTERVAL env variable"))
	}

	for {
		err := deleteOld()
		if err != nil {
			log.Error(err)
		}

		log.Debug("Query executed")

		time.Sleep(interval)
	}
}

func deleteOld() error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return extendError(err, "db: failed to open database")
	}

	err = db.Ping()
	if err != nil {
		return extendError(err, "db: failed to establish connection")
	}

	sqlStatement := `DELETE FROM object WHERE last_seen < (now()-'30 seconds'::interval)`

	_, err = db.Exec(sqlStatement)
	if err != nil {
		return extendError(err, "db: failed to delete old records")
	}

	err = db.Close()
	if err != nil {
		return extendError(err, "db: failed to close connection")
	}

	return nil
}
