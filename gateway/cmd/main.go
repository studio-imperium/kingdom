package main

import (
	"database/sql"
	"gateway/gameservers"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=/var/run/postgresql port=5432 dbname=kingdoms")

	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	gameservers.GetGameservers(db)
	db.Close()
}
