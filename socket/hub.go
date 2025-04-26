package socket

import (
	"bytes"
	"log"
	"sync"
)

type Filter struct {
	Clients []*Client
	Message []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Inbound messages from the clients.
	broadcast chan []byte

	filtercast chan *Filter

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Registered clients.
	clients map[*Client]bool

	// Mutex for clients
	mutex sync.Mutex

	// Notify channel
	updateList chan struct{}
}

// NewHub creates new Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		updateList: make(chan struct{}),
	}
}

func (h *Hub) Broadcast(message string) {
	h.broadcast <- []byte(message)
}

// Run handles communication operations with Hub
func (h *Hub) Run() {
	log.Println("HUB RUNNING")
	for {
		select {
		case client := <-h.Register:
			log.Printf("Registering client %s", client.name)

			h.mutex.Lock()
			h.clients[client] = true
			go func() {
				h.updateList <- struct{}{}
			}()
			h.mutex.Unlock()

		case client := <-h.Unregister:

			log.Printf("Unregistering client %s", client.name)

			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			go func() {
				h.updateList <- struct{}{}
			}()
			h.mutex.Unlock()

		case filter := <-h.filtercast:
			h.mutex.Lock()
			for _, client := range filter.Clients {
				client.send <- filter.Message
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				client.send <- message
			}
			h.mutex.Unlock()

		case <-h.updateList:
			h.mutex.Lock()
			b := []byte("<div id=\"users\" class=\"users\">" + h.GetClientList() + "</div>")
			for client := range h.clients {
				client.send <- b
			}
			h.mutex.Unlock()
		}
	}
}

func (h *Hub) GetClientList() string {
	var buf bytes.Buffer
	for client := range h.clients {
		if !client.active {
			buf.WriteString("<div  class=\"user-inactive\">" + client.name + "</i></div>")
		} else {
			buf.WriteString("<div class=\"user-active\">" + client.name + "</b></div>")
		}
	}
	return buf.String()
}
