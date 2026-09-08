package sessions

import (
	"context"
	"encoding/hex"
	"fmt"
	"gateway/data"

	"github.com/redis/go-redis/v9"
)

var session_user_query string = `SELECT id FROM game.users WHERE email = $1`
var session_characters_query string = `SELECT id, COALESCE(dead, false), inventory, level, exp, username FROM game."character" WHERE user_id = $1 ORDER BY id`

func GetSession(token SessionToken) (Session, error) {
	var session Session
	var ctx context.Context = context.Background()
	email, err := CheckToken(token)
	if err != nil {
		return session, err
	}
	session.Email, session.Data = email, DefaultAccountData()
	session.Active, err = data.Cache.HGet(ctx, "session:token:"+hex.EncodeToString(token[:]), "active").Bool()
	// Legacy string sessions have no active flag.
	if err != nil && err != redis.Nil && !redis.HasErrorPrefix(err, "WRONGTYPE") {
		return session, err
	}
	if email == "" {
		character := DefaultCharacter()
		character.Id = 0
		session.Valid, session.Guest = true, true
		session.Data.Characters = append(session.Data.Characters, character)
		return session, nil
	}
	if err := data.DB.QueryRowContext(ctx, session_user_query, email).Scan(&session.Id); err != nil {
		return session, err
	}
	rows, err := data.DB.QueryContext(ctx, session_characters_query, session.Id)
	if err != nil {
		return session, err
	}
	defer rows.Close()
	for rows.Next() {
		var character Character
		var inventory []byte
		if err := rows.Scan(&character.Id, &character.Dead, &inventory, &character.Level, &character.Exp, &character.Username); err != nil {
			return session, err
		}
		if len(inventory) != 27 {
			return session, fmt.Errorf("invalid inventory length for character %d: %d", character.Id, len(inventory))
		}
		character.Inventory = map[uint8]uint8{}
		for slot, item := range inventory[:24] {
			if item != 0 {
				character.Inventory[uint8(slot)] = item
			}
		}
		character.Head, character.Body, character.Hand = inventory[24], inventory[25], inventory[26]
		if character.Dead {
			session.Data.Graveyard = append(session.Data.Graveyard, character)
		} else {
			session.Data.Characters = append(session.Data.Characters, character)
		}
	}
	if err := rows.Err(); err != nil {
		return session, err
	}
	session.Valid = true
	return session, nil
}
