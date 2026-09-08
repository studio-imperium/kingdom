package clients

import (
	"encoding/binary"
	"errors"
	"gameserver/engine"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func CreateClient(w http.ResponseWriter, r *http.Request) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	client := &Client{conn: conn, done: make(chan struct{})}
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	messageType, data, err := conn.ReadMessage()
	if err != nil {
		client.Close()
		return nil, err
	}
	// HANDSHAKE + 32-byte session token + little-endian int64 character ID + UTF-8 name (1-255 bytes).
	if messageType != websocket.BinaryMessage || len(data) < 42 || len(data) > 296 || data[0] != engine.HANDSHAKE {
		client.Close()
		return nil, errors.New("invalid handshake")
	}
	copy(client.Token[:], data[1:33])
	client.CharacterID = int64(binary.LittleEndian.Uint64(data[33:]))
	client.Username = strings.TrimSpace(string(data[41:]))
	if client.Username == "" || !utf8.ValidString(client.Username) {
		client.Close()
		return nil, errors.New("invalid player name")
	}
	if client.CharacterID < 0 {
		client.Close()
		return nil, errors.New("invalid character ID")
	}
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	return client, nil
}

func (client *Client) Close() {
	client.close_once.Do(func() { close(client.done); client.conn.Close() })
}

// Run returns only after both socket loops stop. It has no engine or character reference.
func (client *Client) Run(input chan<- engine.Packet, output <-chan []byte) {
	written := make(chan struct{})
	go func() { defer close(written); client.writePackets(output) }()
	client.receivePackets(input)
	client.Close()
	<-written
}

func (client *Client) receivePackets(input chan<- engine.Packet) {
	defer client.Close()
	for {
		messageType, data, err := client.conn.ReadMessage()
		if err != nil || messageType != websocket.BinaryMessage {
			return
		}
		packet, err := ParsePacket(data)
		if err != nil {
			return
		}
		packet.ID = client.ID
		select {
		case input <- packet:
		case <-client.done:
			return
		}
	}
}

func (client *Client) writePackets(output <-chan []byte) {
	defer client.Close()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-client.done:
			return
		case message, okay := <-output:
			if !okay {
				return
			}
			client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := client.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}
			if len(message) > 0 && message[0] == engine.CHARACTER_DEAD {
				return
			}
		case <-ticker.C:
			if err := client.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}
