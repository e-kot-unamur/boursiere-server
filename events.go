package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// Broker represents a server-sent events (SSE) route.
//
// Once created with NewBroker, use ServeHTTP as a route handler and Broadcast
// to send events.
type Broker struct {
	clients   map[chan string]struct{}
	openChan  chan chan string
	closeChan chan chan string
	sendChan  chan string
}

// NewBroker creates a ready-to-use broker. It starts a new goroutine that
// handles incoming messages and connections.
func NewBroker() Broker {
	b := Broker{
		clients:   make(map[chan string]struct{}),
		openChan:  make(chan chan string),
		closeChan: make(chan chan string),
		sendChan:  make(chan string, 4),
	}

	go b.listen()
	return b
}

// ServeHTTP is a SSE handler function that sends any broadcasted message.
func (b *Broker) ServeHTTP(c *gin.Context) {
	msgs := b.Subscribe()
	defer b.Unsubscribe(msgs)

	w := c.Writer
	closed := w.CloseNotify()

	// Don't use c.Stream(): it would block the connection until a new message
	// is received whereas we should listen for disconnection simultaneously.
	for {
		select {
		case msg := <-msgs:
			c.SSEvent("", msg)
			w.Flush()
		case <-closed:
			return
		}
	}
}

// Subscribe creates, registers and returns a new channel that can be used to
// receive incoming messages.
func (b *Broker) Subscribe() chan string {
	c := make(chan string)
	b.openChan <- c
	return c
}

// Unsubscribe closes a channel that was created with Subscribe. It is
// mandatory to unsubscribe when a client disconnects.
func (b *Broker) Unsubscribe(c chan string) {
	b.closeChan <- c
}

// Broadcast sends a message to every connected client. The sent event is
// marshalled into JSON beforehand.
func (b *Broker) Broadcast(event interface{}) {
	msg, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	b.sendChan <- string(msg)
}

// listen handles client's connection and disconnection, and broadcasts any
// incoming message to all clients.
func (b *Broker) listen() {
	for {
		select {
		case c := <-b.openChan:
			b.clients[c] = struct{}{}

		case c := <-b.closeChan:
			close(c)
			delete(b.clients, c)

		case msg := <-b.sendChan:
			for client := range b.clients {
				b.send(client, msg)
			}
		}
	}
}

// send writes a message to a client's channel. It returns when the message is
// sent or if the client disconnected.
func (b *Broker) send(client chan string, msg string) {
	// A client could disconnect while we're trying to send a message.
	// So we MUST listen to the closeChan to avoid a (sneaky) deadlock!
	for {
		select {
		case client <- msg:
			return

		case c := <-b.closeChan:
			close(c)
			delete(b.clients, c)
			if c == client {
				return
			}
		}
	}
}
