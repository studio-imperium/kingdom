package session

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"kingdoms/protocol"
)

type outbound struct {
	data  []byte
	close bool
}

type client struct {
	server     *Server
	connection *websocket.Conn
	id         uint32
	username   string
	admin      bool
	ready      atomic.Bool
	outbound   chan outbound
	done       chan struct{}
	closeOnce  sync.Once
	viewMu     sync.Mutex
	discovered map[uint16]struct{}
}

func (c *client) run() {
	go c.write()
	c.read()
}

func (c *client) read() {
	defer c.close()
	for {
		messageType, reader, err := c.connection.NextReader()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		command, err := protocol.Decode(reader)
		if err != nil || c.server.handle(c, command) != nil {
			return
		}
	}
}

func (c *client) write() {
	defer c.close()
	for {
		select {
		case <-c.done:
			return
		case message := <-c.outbound:
			c.connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.connection.WriteMessage(websocket.BinaryMessage, message.data); err != nil {
				return
			}
			if message.close {
				return
			}
		}
	}
}

func (c *client) send(data []byte) {
	c.enqueue(data, false)
}

func (c *client) trySend(data []byte) {
	if len(data) == 0 {
		return
	}
	select {
	case <-c.done:
		return
	case c.outbound <- outbound{data: data}:
		return
	default:
		return
	}
}

func (c *client) enqueue(data []byte, closeAfter bool) {
	if len(data) == 0 {
		return
	}
	select {
	case <-c.done:
		return
	case c.outbound <- outbound{data: data, close: closeAfter}:
		return
	default:
		c.close()
	}
}

func (c *client) discover(cellID uint16) bool {
	c.viewMu.Lock()
	defer c.viewMu.Unlock()
	if _, exists := c.discovered[cellID]; exists {
		return false
	}
	c.discovered[cellID] = struct{}{}
	return true
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.connection.Close()
		c.server.unregister(c)
	})
}
