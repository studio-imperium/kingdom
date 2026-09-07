package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gateway/data"
)

type SessionToken [sha256.Size]byte

var ErrInvalidCredentials = errors.New("invalid email or password")

var create_user_query string = `INSERT INTO game.users ("name", email, password, data) VALUES ($1, $2, $3, $4)`
var find_user_query string = "SELECT EXISTS (SELECT 1 FROM game.users WHERE email = $1 AND password = $2)"
var replace_session_query string = `
	local old = redis.call('GET', KEYS[1])
	if old then redis.call('DEL', old) end
	redis.call('SET', KEYS[1], KEYS[2])
	redis.call('SET', KEYS[2], ARGV[1])
	return 1
`

func Signup(name, email, password string) (SessionToken, error) {
	var hashed_password SessionToken = sha256.Sum256([]byte(password))
	account_data, _ := json.Marshal(DefaultAccountData())

	_, err := data.DB.ExecContext(context.Background(), create_user_query, name, email, hashed_password[:], string(account_data))
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
	return data.Cache.Get(context.Background(), "session:token:"+hex.EncodeToString(token[:])).Result()
}
