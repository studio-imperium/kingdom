package clients

import (
	"gameserver/gateway"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID          uint32
	Token       gateway.SessionToken
	CharacterID int64
	conn        *websocket.Conn
	done        chan struct{}
	close_once  sync.Once
}
