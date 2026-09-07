package gateway

import "crypto/sha256"

type SessionToken [sha256.Size]byte

type Character struct {
	Hand      uint8           `json:"hand"`
	Head      uint8           `json:"head"`
	Body      uint8           `json:"body"`
	Inventory map[uint8]uint8 `json:"inventory"`
}
