package sessions

import "math/rand/v2"

type UserData struct {
	Id    int64       `json:"id"`
	Email string      `json:"email"`
	Data  AccountData `json:"data"`
}

type Session struct {
	UserData
	Valid  bool `json:"valid"`
	Active bool `json:"active"`
	Guest  bool `json:"guest"`
}

type AccountData struct {
	Characters []Character `json:"characters"`
	Graveyard  []Character `json:"graveyard"`
}
type Character struct {
	Id        int64           `json:"id,string"`
	Dead      bool            `json:"dead"`
	Hand      uint8           `json:"hand"`
	Head      uint8           `json:"head"`
	Body      uint8           `json:"body"`
	Inventory map[uint8]uint8 `json:"inventory"`
}

func DefaultAccountData() AccountData {
	return AccountData{
		Characters: []Character{},
		Graveyard:  []Character{},
	}
}

func DefaultCharacter() Character {
	return Character{
		Id:   rand.Int64(),
		Hand: 0,
		Head: 0,
		Body: 1,
		Inventory: map[uint8]uint8{
			0: 8,
		},
	}
}
