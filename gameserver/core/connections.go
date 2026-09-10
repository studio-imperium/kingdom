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
	session, err := gateway.VerifyPlayer(token, PlayersOnline()+1)
	if err != nil {
		return
	}
	var saved *engine.CharacterData
	var finished chan engine.CharacterData
	if session.Guest {
		client.CharacterID = 0 // Guests always join with an unsaved default character.
	} else {
		for i := range session.Data.Characters {
			if session.Data.Characters[i].Id == client.CharacterID {
				saved = &session.Data.Characters[i]
				break
			}
		}
		if saved == nil {
			return
		}
		finished = make(chan engine.CharacterData, 1)
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
	world.Input <- engine.Packet{Type: engine.JOIN, ID: client.ID, CharacterID: client.CharacterID, Username: client.Username, Send: output, Character: saved, Finished: finished}
	client.Run(world.Input, output)
	world.Input <- engine.Packet{Type: engine.LEAVE, ID: client.ID}
	if finished != nil {
		if err := gateway.UpdateCharacter(token, <-finished); err != nil {
			log.Print(err)
		}
	}

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
	if err := engine.InitAssets(gateway.AssetsURL()); err != nil {
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
	mux.HandleFunc("GET /assets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Redirect(w, r, "https://gateway.kingdomcrushers.io"+r.URL.Path, http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
	})
	return http.ListenAndServe(":"+strconv.Itoa(port), mux)
}
