package gameservers

import (
	"context"
	"encoding/hex"
	"fmt"
	"gateway/data"
	sessions "gateway/user_sessions"
	"strconv"
)

var player_activity_query string = `
	if redis.call('TYPE', KEYS[1]).ok == 'string' then
		local email = redis.call('GET', KEYS[1])
		redis.call('DEL', KEYS[1])
		redis.call('HSET', KEYS[1], 'email', email)
	end
	if redis.call('HEXISTS', KEYS[1], 'email') == 0 then return nil end
	return redis.call('HSET', KEYS[1], 'active', ARGV[1])
`
var update_character_query = `
	UPDATE game."character" AS c
	SET inventory = $1
	FROM game.users AS u
	WHERE c.id = $2 AND c."name" = u."name" AND u.email = $3
	RETURNING c.id
`

func ValidSession(token sessions.SessionToken) (string, error) {
	return sessions.CheckToken(token)
}

func PlayerActivity(token sessions.SessionToken, active bool) error {
	return data.Cache.Eval(context.Background(), player_activity_query, []string{"session:token:" + hex.EncodeToString(token[:])}, strconv.FormatBool(active)).Err()
}

func UpdateCharacter(token sessions.SessionToken, character sessions.Character) error {
	email, err := sessions.CheckToken(token)
	if err != nil {
		return err
	}

	// Bytes 0-23 are backpack slots; 24-26 are head, body, and hand.
	var inventory [27]byte
	for slot, item := range character.Inventory {
		if slot >= 24 {
			return fmt.Errorf("invalid inventory slot: %d", slot)
		}
		inventory[slot] = item
	}
	inventory[24], inventory[25], inventory[26] = character.Head, character.Body, character.Hand

	updated_char := data.DB.QueryRowContext(context.Background(), update_character_query, inventory[:], character.Id, email)
	return updated_char.Scan(&character.Id)
}
