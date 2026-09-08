package websocket

// Represents all live connections. Owns collections of connected clients
// and decides which clients receive each broadcast
// Only one goroutine owns `rooms`, other goroutines communicate with it
// through channels

type Hub struct {
	rooms map[int64]map[*Client]bool // Maps the room ID to a set of clients
	clients map[*Client]bool         // All currently registered WebSocket connections
	register   chan Registration     // Channel where the WebSocket handler sends a client to add to the `rooms` map
	unregister chan *Client          // Unregister a client from `rooms`
	broadcast  chan BroadcastMessage // Send the incoming message to every connected to a certain room, the client's `writePump()` does the actual network write
}

// Represents a request to subscribe a client to their rooms they are members of
type Registration struct {
	Client  *Client
	RoomIDs []int64
}

// The object/value that travels through the `brodcast` channel to the hub
type BroadcastMessage struct {
	RoomID  int64
	Message []byte
}

// Must initialise the map before assigning into it
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		clients:    make(map[*Client]bool),
		register:   make(chan Registration),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage),
	}
}

// Register a WebSocket connection and subscribe it to all of the user's current rooms
func (h *Hub) registerClient(registration Registration) {
	client := registration.Client
	
	// Register the live WebSocket connection
	h.clients[client] = true

	// Register the client to every room they are in
	for _, roomID := range registration.RoomIDs {
		clients, ok := h.rooms[roomID]

		// If nobody is currently registered to this room,
		// create the room's client set (and register the current user)
		if !ok {
			clients = make(map[*Client]bool)
			h.rooms[roomID] = clients
		}

		clients[client] = true
	}
}

// Remove a client from every room they are subscribed tp (no longer listening for messages)
func (h *Hub) unregisterClient(client *Client) {
	// If the client has already been unregistered, return
	if !h.clients[client] {
		return
	}

	delete(h.clients, client)

	for roomID, clients := range h.rooms {
		if _, exists := clients[client]; !exists {
			continue
		}

		delete(clients, client)

		// If nobody is subscribed to this room anymore,
		// remove the room from the map.
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}

	// TODO: Check idempotency and sync.Once
	close(client.send) // client's writePump can stop listening
}

// Sends the WebSocket message to every connected client
func (h *Hub) broadcastToRoom(message BroadcastMessage) {
	// Get clients connected to the room
	clients, ok := h.rooms[message.RoomID]
	if !ok {
		return
	}

	// For every client, send the message by sending it to their send channel
	// TODO: May have to check if they are banned during the WebSocket connection
	for client := range clients {
		select {
		case client.send <- message.Message:

		default: // prevents client.send <- message.Message from blocking (e.g. if Alice's queue is full) for example due to poor network connection or something
			h.unregisterClient(client) // so simply disconnects the client so they can no longer listen for messages. The message is not lost, it is in PostgreSQL and can be loaded
		}
	}
}

// Run starts the Hub's event loop
// The goroutine running this method is the only goroutine that directly
// accesses h.rooms, pther goroutines communicate with the Hub through
// register, unregister, and broadcast channels
func (h *Hub) Run() {
	for {
		select {
		case registration := <-h.register:
			h.registerClient(registration)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}
