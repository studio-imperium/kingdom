package sessions

type UserData struct {
	id    int8
	name  string
	email string
	data  AccountData
}

type AccountData struct {
	Characters []Character `json:"characters"`
	Graveyard  []Character `json:"graveyard"`
}
type Character struct {
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
		Hand: 0,
		Head: 0,
		Body: 1,
		Inventory: map[uint8]uint8{
			0: 8,
		},
	}
}
