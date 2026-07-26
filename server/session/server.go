package session

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"kingdoms/engine"
	"kingdoms/protocol"
)

type Server struct {
	world    *engine.World
	upgrader websocket.Upgrader
	mu       sync.RWMutex
	clients  map[uint32]*client
	guests   atomic.Uint64
}

func New(world *engine.World) *Server {
	return &Server{
		world: world,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		},
		clients: make(map[uint32]*client),
	}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	connection, err := s.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return
	}

	client := &client{
		server:     s,
		connection: connection,
		username:   fmt.Sprintf("Guest_%d", s.guests.Add(1)),
		admin:      true,
		outbound:   make(chan outbound, 64),
		done:       make(chan struct{}),
		discovered: make(map[uint16]struct{}),
	}
	client.run()
}

func (s *Server) Run(ctx context.Context) {
	worldTicker := time.NewTicker(50 * time.Millisecond)
	snapshotTicker := time.NewTicker(200 * time.Millisecond)
	cellTicker := time.NewTicker(500 * time.Millisecond)
	defer worldTicker.Stop()
	defer snapshotTicker.Stop()
	defer cellTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.close()
			return
		case <-worldTicker.C:
			s.deliver(s.world.Tick(50 * time.Millisecond))
		case <-snapshotTicker.C:
			s.sendSnapshots()
		case <-cellTicker.C:
			s.sendCells()
		}
	}
}

func (s *Server) handle(client *client, command protocol.Command) error {
	if !client.ready.Load() {
		handshake, valid := command.(protocol.HandshakeCommand)
		if !valid {
			return fmt.Errorf("handshake required")
		}
		return s.handshake(client, handshake.ID)
	}

	id := client.id
	switch command := command.(type) {
	case protocol.MoveCommand:
		if !s.world.MoveCharacter(id, command.X, command.Y, command.Angle) {
			return fmt.Errorf("character not found")
		}
	case protocol.AttackCommand:
		s.deliver(s.world.Attack(
			id,
			command.X,
			command.Y,
			command.TargetX,
			command.TargetY,
			command.Angle,
		))
	case protocol.SelectSlotCommand:
		if character, changed := s.world.SelectSlot(id, command.Slot); changed {
			client.send(protocol.EncodeCharacter(character))
		}
	case protocol.ChangeInventoryCommand:
		if character, changed := s.world.ChangeInventory(id, command.To, command.From); changed {
			client.send(protocol.EncodeCharacter(character))
		}
	case protocol.DropItemCommand:
		if character, dropped := s.world.DropItem(id, command.Slot); dropped {
			client.send(protocol.EncodeCharacter(character))
		}
	case protocol.ChatCommand:
		s.handleChat(client, command.Message)
	case protocol.HandshakeCommand:
		return fmt.Errorf("already connected")
	}
	return nil
}

func (s *Server) handshake(client *client, id uint32) error {
	if id == 0 {
		return fmt.Errorf("invalid character id")
	}
	s.mu.Lock()
	if _, exists := s.clients[id]; exists {
		s.mu.Unlock()
		return fmt.Errorf("character %d already connected", id)
	}
	character, err := s.world.AddCharacter(id)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	client.id = id
	client.ready.Store(true)
	s.clients[id] = client
	s.mu.Unlock()

	client.send(protocol.EncodeCharacter(character))
	s.sendClientCells(client)
	return nil
}

func (s *Server) unregister(client *client) {
	if !client.ready.Load() {
		return
	}

	s.mu.Lock()
	if s.clients[client.id] == client {
		delete(s.clients, client.id)
		s.mu.Unlock()
		s.world.RemoveCharacter(client.id)
		return
	}
	s.mu.Unlock()
}

func (s *Server) deliver(events []engine.Event) {
	for _, event := range events {
		packet := protocol.EncodeMessage(event.Message)
		if len(packet) == 0 {
			continue
		}
		_, closes := event.Message.(engine.DeathMessage)
		for _, id := range event.Recipients {
			if client := s.client(id); client != nil {
				client.enqueue(packet, closes)
			}
		}
	}
}

func (s *Server) sendSnapshots() {
	for _, client := range s.clientList() {
		snapshot, valid := s.world.SnapshotFor(client.id)
		if valid {
			client.trySend(protocol.EncodeSnapshot(snapshot))
		}
	}
}

func (s *Server) sendCells() {
	for _, client := range s.clientList() {
		s.sendClientCells(client)
	}
}

func (s *Server) sendClientCells(client *client) {
	cells, valid := s.world.CellsFor(client.id)
	if !valid {
		return
	}
	for _, cell := range cells {
		if client.discover(cell.ID) {
			client.send(protocol.EncodeCell(cell))
		}
	}
}

func (s *Server) client(id uint32) *client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[id]
}

func (s *Server) clientList() []*client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clients := make([]*client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *Server) close() {
	for _, client := range s.clientList() {
		client.close()
	}
}
