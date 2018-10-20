package main

type message struct {
	data []byte
	room string
}

type subscription struct {
	conn *Client
	room string
}

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan message

	// Register requests from the clients.
	register chan subscription

	// Unregister requests from clients.
	unregister chan subscription
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			connections := h.rooms[client.room]
			if connections == nil {
				connections = make(map[*Client]bool)
				h.rooms[client.room] = connections
			}
			h.rooms[client.room][client.conn] = true
		case client := <-h.unregister:
			connections := h.rooms[client.room]
			if connections != nil {
				if _, ok := connections[client.conn]; ok {
					delete(connections, client.conn)
					close(client.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, client.room)
					}
				}
			}
		case message := <-h.broadcast:
			connections := h.rooms[message.room]
			for client := range connections {
				select {
				case client.send <- message.data:
				default:
					close(client.send)
					delete(connections, client)
					if len(connections) == 0 {
						delete(h.rooms, message.room)
					}
				}
			}
		}
	}
}
