package main

import (
	"database/sql"
	"gateway/gameservers"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	gameservers.GetGameservers(db)
	db.Close()
}
