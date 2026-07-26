package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"kingdoms/engine"
	"kingdoms/engine/assets"
	"kingdoms/session"
)

func main() {
	catalog, err := assets.Load()
	if err != nil {
		log.Fatal(err)
	}

	world, err := engine.NewWorld(catalog, "desertonly")
	if err != nil {
		log.Fatal(err)
	}

	sessions := session.New(world)
	go sessions.Run(context.Background())

	mux := http.NewServeMux()
	mux.Handle("/connect", sessions)
	mux.Handle(
		"/assets/",
		withCORS(http.StripPrefix("/assets/", http.FileServer(http.FS(assets.Files())))),
	)

	fmt.Println("Listening on 8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
