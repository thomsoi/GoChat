package websocket

// Links HTTP and WebSocket transport, and the application logic.
// This file is responsible for handling clients who want to establish
// a WebSocket connection to a specific room.

import (
	"context"
	"net/http"

	"gochat/internal/auth"
	"gochat/internal/room"

	"github.com/gorilla/websocket"
)

// The handler needs the room service layer to check if a user that wants to send
// a message to a particular room is actually a member of that room
// and it also needs the hub to register the client to the rooms they belong to
type Handler struct {
	rooms    RoomService
	hub      *Hub
	messages MessageService
}

type RoomService interface {
	ListRoomsByUser(ctx context.Context, userID int64) ([]room.Room, error)
	//IsMember(ctx context.Context, roomID int64, userID int64) (bool, error)
}

// The HTTP upgrader to WebSocket
var upgrader = websocket.Upgrader{}

func NewHandler(rooms RoomService, hub *Hub, messages MessageService) *Handler {
	return &Handler{rooms: rooms, hub: hub, messages: messages}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	
	// Gets the user ID of the authenticated user
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get every room that the authenticated user belongs to
	rooms, err := h.rooms.ListRoomsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// // NOTE: To implement other access rules, such as archived room
	// // or locked room, then another service function will have to be created
	// isMember, err := h.rooms.IsMember(
	// 	r.Context(),
	// 	roomID,
	// 	userID,
	// )
	// if !isMember {
	// 	http.Error(w, "Forbidden", http.StatusForbidden)
	// 	return
	// }

	// Extract just the room IDs because the Hub only needs the IDs
	// to create the room -> client subscriptions
	roomIDs := make([]int64, 0, len(rooms))
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.ID)
	}

	// Establish the WebSocket connection by upgrading the HTTP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create the client (application's representation of this one live connection)
	client := &Client{
		conn:     conn,
		hub:      h.hub,
		userID:   userID,
		messages: h.messages,
		send:     make(chan []byte, 256), // 256 []byte messages to be waiting to be processed is the maximum before a sender has to wait/block
	}

	// Register the client
	h.hub.register <- Registration{Client: client, RoomIDs: roomIDs}

	// Start the two read and write goroutines for server->client and client->server communication
	go client.writePump()
	go client.readPump()
}
