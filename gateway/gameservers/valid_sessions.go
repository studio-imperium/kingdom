package gameservers

import (
	"context"
	"encoding/hex"
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

func ValidSession(token sessions.SessionToken) (string, error) {
	return sessions.CheckToken(token)
}

func PlayerActivity(token sessions.SessionToken, active bool) error {
	// Check and update atomically so a logged-out token cannot be recreated.
	return data.Cache.Eval(context.Background(), player_activity_query, []string{"session:token:" + hex.EncodeToString(token[:])}, strconv.FormatBool(active)).Err()
}
