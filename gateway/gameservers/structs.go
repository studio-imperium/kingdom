package gameservers

import (
	"database/sql"
	"log"
)

type Gameserver struct {
	name              string
	address           string
	max_players       int
	connected_players int
}

var get_gameservers_query string = "SELECT name, address, max_players FROM architecture.servers"

func GetGameservers(db *sql.DB) {
	rows, err := db.Query(get_gameservers_query)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		gameserver := Gameserver{}

		if err := rows.Scan(&gameserver.name, &gameserver.address, &gameserver.max_players); err != nil {
			log.Fatal(err)
		}
		log.Printf("name is %s\n", gameserver.name)
	}
	rows.Close()
}
