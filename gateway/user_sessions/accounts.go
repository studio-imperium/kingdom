package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"gateway/data"

	"github.com/redis/go-redis/v9"
)

type SessionToken [sha256.Size]byte

var ErrInvalidCredentials = errors.New("invalid email or password")

var create_user_query string = `INSERT INTO game.users ("name", email, password) VALUES ($1, $2, $3)`
var find_user_query string = "SELECT EXISTS (SELECT 1 FROM game.users WHERE email = $1 AND password = $2)"
var replace_session_query string = `
	local old = redis.call('GET', KEYS[1])
	if old then redis.call('DEL', old) end
	redis.call('SET', KEYS[1], KEYS[2])
	redis.call('HSET', KEYS[2], 'email', ARGV[1], 'active', 'false')
	return 1
`

func Signup(name, email, password string) (SessionToken, error) {
	var hashed_password SessionToken = sha256.Sum256([]byte(password))

	_, err := data.DB.ExecContext(context.Background(), create_user_query, name, email, hashed_password[:])
	if err != nil {
		return SessionToken{}, err
	}

	return Login(email, password)
}

func Login(email, password string) (SessionToken, error) {
	var valid_user bool
	var token SessionToken = SessionToken{}
	var hashed_password SessionToken = sha256.Sum256([]byte(password))
	var ctx context.Context = context.Background()

	user_row := data.DB.QueryRowContext(ctx, find_user_query, email, hashed_password[:])
	err := user_row.Scan(&valid_user)

	if err != nil {
		return token, err
	}

	if !valid_user {
		return token, ErrInvalidCredentials
	}

	rand.Read(token[:])
	err = data.Cache.Eval(ctx, replace_session_query, []string{"session:email:" + email, "session:token:" + hex.EncodeToString(token[:])}, email).Err()
	if err != nil {
		return SessionToken{}, err
	}

	return token, nil
}

func Logout(token SessionToken) error {
	return data.Cache.Del(context.Background(), "session:token:"+hex.EncodeToString(token[:])).Err()
}

func CheckToken(token SessionToken) (string, error) {
	var key string = "session:token:" + hex.EncodeToString(token[:])
	email, err := data.Cache.HGet(context.Background(), key, "email").Result()
	// Older sessions store the email directly instead of using a hash.
	if redis.HasErrorPrefix(err, "WRONGTYPE") {
		return data.Cache.Get(context.Background(), key).Result()
	}
	return email, err
}
