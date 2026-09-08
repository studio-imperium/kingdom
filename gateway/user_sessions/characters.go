package sessions

import (
	"context"
	"database/sql"
	"errors"
	"gateway/data"
)

var ErrCharacterLimit = errors.New("character limit reached")

var lock_character_user_query string = `SELECT id FROM game.users WHERE email = $1 FOR UPDATE`
var count_characters_query string = `SELECT COUNT(*) FROM game."character" WHERE user_id = $1 AND dead IS NOT TRUE`
var create_character_query string = `INSERT INTO game."character" (id, user_id, dead, inventory) VALUES ($1, $2, false, $3)`

func NewCharacter(token SessionToken) error {
	// create a new character
	var ctx context.Context = context.Background()
	var user_id int64
	var character_count int

	email, err := CheckToken(token)
	if err != nil {
		return err
	}
	if email == "" {
		return ErrInvalidCredentials
	}

	tx, err := data.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Serialize character creation for this account before checking the limit.
	err = tx.QueryRowContext(ctx, lock_character_user_query, email).Scan(&user_id)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, count_characters_query, user_id).Scan(&character_count)
	if err != nil {
		return err
	}
	if character_count >= 1 {
		return ErrCharacterLimit
	}

	// Bytes 0-23 are backpack slots; 24 is head, 25 is body, 26 is hand.
	// Zero means an empty backpack slot; omit it when decoding into the map.
	var inventory [27]byte
	character := DefaultCharacter()
	for slot, item := range character.Inventory {
		inventory[slot] = item
	}
	inventory[24], inventory[25], inventory[26] = character.Head, character.Body, character.Hand

	_, err = tx.ExecContext(ctx, create_character_query, character.Id, user_id, inventory[:])
	if err != nil {
		return err
	}
	return tx.Commit()
}
