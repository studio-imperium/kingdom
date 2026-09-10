package gateway

import (
	"encoding/json"
	"fmt"
	"gameserver/engine"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}
var gateway_address string = "http://localhost:8080"
var character_update string = "/players/update"
var player_activity string = "/players/active"
var player_verify string = "/players/verify"
var server_address string = "server.kingdomcrushers.io"

func AssetsURL() string { return gateway_address + "/assets/" }

type Session struct {
	Valid bool `json:"valid"`
	Guest bool `json:"guest"`
	Data  struct {
		Characters []engine.CharacterData `json:"characters"`
	} `json:"data"`
}

func VerifyPlayer(token SessionToken, players_online int) (Session, error) {
	var session Session
	err := post(token, player_verify, "application/json", []byte(`{"address":`+strconv.Quote(server_address)+`,"players_online":`+strconv.Itoa(players_online)+`}`), &session)
	if err == nil && !session.Valid {
		err = fmt.Errorf("invalid session")
	}
	return session, err
}

func UpdatePlayerActivity(token SessionToken, activity bool) error {
	return post(token, player_activity, "application/json", []byte(`{"active":`+strconv.FormatBool(activity)+`}`), nil)
}

func UpdateCharacter(token SessionToken, character engine.CharacterData) error {
	data, err := json.Marshal(character)
	if err != nil {
		return err
	}
	return post(token, character_update, "application/json", data, nil)
}
