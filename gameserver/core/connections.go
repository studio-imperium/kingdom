package core

import (
	"context"
	"gameserver/core/clients"
	"gameserver/engine"
	"gameserver/gateway"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var port = 8082
var connected_players = make(map[gateway.SessionToken]*clients.Client)
var players_mutex sync.Mutex
var world *engine.Engine
var next_player_id uint32

func ProcessConnection(w http.ResponseWriter, r *http.Request) {
	client, err := clients.CreateClient(w, r)
	if err != nil {
		log.Print(err)
		return
	}
	var token gateway.SessionToken = client.Token
	defer client.Close()
	if !gateway.VerifyPlayer(token) {
		return
	}

	players_mutex.Lock()
	if _, exists := connected_players[token]; exists {
		players_mutex.Unlock()
		return
	}
	next_player_id++
	client.ID = next_player_id
	connected_players[token] = client
	players_mutex.Unlock()

	output := make(chan []byte, 256)
	world.Input <- engine.Packet{Type: engine.JOIN, ID: client.ID, CharacterID: client.CharacterID, Username: "Guest_" + strconv.FormatUint(uint64(client.ID), 10), Send: output}
	client.Run(world.Input, output)
	world.Input <- engine.Packet{Type: engine.LEAVE, ID: client.ID}

	players_mutex.Lock()
	delete(connected_players, token)
	players_mutex.Unlock()
}

func PlayersOnline() int {
	players_mutex.Lock()
	defer players_mutex.Unlock()
	return len(connected_players)
}

func Start() error {
	if err := engine.InitAssets(); err != nil {
		return err
	}
	var err error
	world, err = engine.CreateIsland()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go world.Run(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /connect", ProcessConnection)
	assets := http.StripPrefix("/assets/", http.FileServer(http.FS(engine.JSONAssets())))
	mux.HandleFunc("GET /assets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		assets.ServeHTTP(w, r)
	})
	return http.ListenAndServe(":"+strconv.Itoa(port), mux)
}
