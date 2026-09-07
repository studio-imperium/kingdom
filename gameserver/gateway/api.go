package gateway

import (
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}
var gateway_address string = "http://localhost:8080"
var character_update string = "/players/update"
var player_activity string = "/players/active"
var player_verify string = "/players/verify"

func VerifyPlayer(token SessionToken) bool {
	return post(token, player_verify, "application/json", nil) == nil
}

func UpdatePlayerActivity(token SessionToken, activity bool) error {
	return post(token, player_activity, "application/json", []byte(`{"active":`+strconv.FormatBool(activity)+`}`))
}

func UpdateCharacter(token SessionToken, character_packed []byte) error {
	return post(token, "/players/update", "application/octet-stream", character_packed)
}
