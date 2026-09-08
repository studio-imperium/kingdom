package gameservers

import (
	"gateway/data"
	"sync"
)

type Gameserver struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	MaxPlayers    int    `json:"max_players"`
	PlayersOnline int    `json:"players_online"`
}

var get_gameservers_query string = "SELECT name, address, max_players FROM architecture.servers"
var player_counts sync.Map

func ReportPlayers(address string, count int) { player_counts.Store(address, count) }

func GetGameservers() ([]Gameserver, error) {
	rows, err := data.DB.Query(get_gameservers_query)

	var res []Gameserver = []Gameserver{}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		gameserver := Gameserver{}

		if err := rows.Scan(&gameserver.Name, &gameserver.Address, &gameserver.MaxPlayers); err != nil {
			return nil, err
		}
		if count, ok := player_counts.Load(gameserver.Address); ok {
			gameserver.PlayersOnline = count.(int)
		}

		res = append(res, gameserver)
	}

	return res, rows.Err()
}
