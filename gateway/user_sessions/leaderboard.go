package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/data"
	"time"

	"github.com/redis/go-redis/v9"
)

var leaderboard_query = `SELECT id, level, exp, COALESCE(dead, false), COALESCE(username, ''), inventory FROM game."character" ORDER BY level DESC, exp DESC, id ASC LIMIT 100`

type LeaderboardCharacter struct {
	Id       int64  `json:"id,string"`
	Level    int64  `json:"level,string"`
	Exp      int64  `json:"exp,string"`
	Dead     bool   `json:"dead"`
	Username string `json:"username"`
	Head     uint8  `json:"head"`
	Body     uint8  `json:"body"`
	Weapon   uint8  `json:"weapon"`
}

func GetLeaderboard() (json.RawMessage, error) {
	ctx := context.Background()
	cached, err := data.Cache.Get(ctx, "leaderboard:all_time:v2").Bytes()
	if err == nil {
		return cached, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	rows, err := data.DB.QueryContext(ctx, leaderboard_query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	leaderboard := []LeaderboardCharacter{}
	for rows.Next() {
		var character LeaderboardCharacter
		var inventory []byte
		if err := rows.Scan(&character.Id, &character.Level, &character.Exp, &character.Dead, &character.Username, &inventory); err != nil {
			return nil, err
		}
		if len(inventory) != 27 || inventory[26] >= 24 {
			return nil, fmt.Errorf("invalid inventory for character %d", character.Id)
		}
		character.Head, character.Body, character.Weapon = inventory[24], inventory[25], inventory[inventory[26]]
		leaderboard = append(leaderboard, character)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(leaderboard)
	if err != nil {
		return nil, err
	}
	if err := data.Cache.Set(ctx, "leaderboard:all_time:v2", encoded, 10*time.Second).Err(); err != nil {
		return nil, err
	}
	return encoded, nil
}
