package gameservers

import "gateway/data"

type Gameserver struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	MaxPlayers int    `json:"max_players"`
}

var get_gameservers_query string = "SELECT name, address, max_players FROM architecture.servers"

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

		res = append(res, gameserver)
	}

	return res, rows.Err()
}
