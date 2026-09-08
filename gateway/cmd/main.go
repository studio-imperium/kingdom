package main

import (
	"gateway/data"
	"log"
	"net/http"
)

func createApiSurface() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", auth(false))
	mux.HandleFunc("POST /logout", logout)
	mux.HandleFunc("POST /player/guest", guest)
	mux.HandleFunc("POST /register", auth(true))
	mux.HandleFunc("POST /character/new", create_character)

	mux.HandleFunc("GET /gameservers", get_gameservers)
	mux.HandleFunc("GET /leaderboard", get_leaderboard)
	mux.HandleFunc("POST /players/update", update_player)
	mux.HandleFunc("POST /players/verify", verify_token)
	mux.HandleFunc("POST /players/session", get_session)
	mux.HandleFunc("POST /players/active", player_activity)

	return http.HandlerFunc(cors(mux))
}

func main() {
	if err := data.Open(); err != nil {
		log.Print(err)
		return
	}
	log.Print("Running on port 8080")
	log.Print(http.ListenAndServe(":8080", createApiSurface()))
	data.Close()
}
