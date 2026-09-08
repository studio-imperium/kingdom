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
	mux.HandleFunc("POST /players/update", update_player)
	mux.HandleFunc("POST /players/verify", verify_token)
	mux.HandleFunc("POST /players/session", get_session)
	mux.HandleFunc("POST /players/active", player_activity)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Browser clients authenticate with bearer tokens, not cookies.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})
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
