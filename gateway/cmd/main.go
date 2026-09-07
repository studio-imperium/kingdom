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
	mux.HandleFunc("POST /register", auth(true))
	mux.HandleFunc("POST /character/new", create_character)

	mux.HandleFunc("GET /gameservers", get_gameservers)
	mux.HandleFunc("POST /token/verify", verify_token)
	mux.HandleFunc("POST /token/active", player_activity)
	return mux
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
