package gateway

var gateway_address string = "localhost:8080"
var character_update string = gateway_address + "/players/update"
var player_activity string = gateway_address + "/players/active"
var player_verify string = gateway_address + "/players/verify"

func UpdatePlayerActivity(token SessionToken, activity bool) {
	// send an https request to [player_activity] to update their activity
	// false = offline (logged out/died) true = online (they just joined)
}

func UpdateCharacter(token SessionToken, hand uint8, head uint8, body uint8, inventory map[uint8]uint8) {
	// once player has either exited the game or died, send their new character so the db is updated accordingly
	// via https request to [character_update], it will expect packed character bytes
}
