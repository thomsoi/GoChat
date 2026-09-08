package websocket

// My application's representation of a specific authenticated user's
// live WebSocket connection, subscribed to one or more rooms

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gochat/internal/message"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn // The persistent connection after the HTTP request is upgraded
	hub      *Hub            // Includes a map of all connected clients and their rooms
	userID   int64           // The authenticated user that created the connection
	messages MessageService  // The message service used to persist the incoming messages
	send     chan []byte     // Outgoing message queue for this client, the Hub broadcasts it, does not directly do conn.WriteMessage, Hub takes control of this
}

type MessageService interface {
	CreateMessage(ctx context.Context, req message.CreateMessageRequest, userID int64) (*message.Message, error)
	UpdateMessage(ctx context.Context, messageID int64, userID int64, content string) (*message.Message, error)
	DeleteMessage(ctx context.Context, messageID int64, userID int64) (error)
	GetMessageByID(ctx context.Context, id int64) (*message.Message, error)
}

// Continously pump messages from this application's outgoing
// queue into the WebSocket connection (server -> client)
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		message, ok := <- c.send // Waits for a message to become available, will be sent by the client and then broadcast to the Hub
		if !ok {
			return
		}

		// Writing to the WebSocket
		// Only writePump writes to `conn` so we do not have multiple
		// goroutines doing it, there is only ONE writer
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Websocket write error: %v", err)
			return
		}
	}
}

// Continuously read incoming WebSocket messages (client -> server)
func (c *Client) readPump() {
	// Cleanup code
	defer func() {
		c.hub.unregister <- c // Must unregister from the Hub so they are also removed from the current active clients map so no messages are sent to this client. Again, only Hub's run() goroutine receives this so no mutex needed
		c.conn.Close()
	}()

	// Keep reading messages, then persist it
	for {
		_, message, err := c.conn.ReadMessage() // returns message type, message payload, error
		if err != nil {
			return
		}

		c.handleMessage(message)
	}
}

// Persist and broadcast the message
func (c *Client) handleMessage(message []byte) {
	//log.Printf("Received from user %d: %s", c.userID, message)
	// c.hub.broadcastToRoom(BroadcastMessage{
	// 	RoomID:  c.roomID,
	// 	Message: message,
	// })

	var incoming IncomingMessage

	if err := json.Unmarshal(message, &incoming); err != nil {
		log.Printf("Invalid websocket message from user %d: %v", c.userID, err)
		return
	}

	switch incoming.Type {
		case TypeMessageSend:
			c.handleMessageSend(incoming.Data)
		case TypeMessageEdit:
			c.handleMessageEdit(incoming.Data)
		case TypeMessageDelete:
			c.handleMessageDelete(incoming.Data)
		default:
			log.Printf("Unknown websocket message type %q from user %d", incoming.Type, c.userID)
	}
}

// This will handle the message of type "message.send"
func (c *Client) handleMessageSend(rawData json.RawMessage) {
	var data MessageSendData

	if err := json.Unmarshal(rawData, &data); err != nil {
		log.Printf("Invalid message.send from user %d: %v", c.userID, err)
		return
	}

	// Give the database operation a bounded amount of time to persist the message
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// roomID comes from the client's message payload
	// userID comes from the authenticated JWT
	// so a client cannot pretend to be another user
	msgReq := message.CreateMessageRequest {
		RoomID: data.RoomID,
		Content: data.Content,
	}
	msg, err := c.messages.CreateMessage(ctx, msgReq, c.userID)
	if err != nil {
		log.Printf("Error when creating message for user %d in room %d: %v",
			c.userID,
			data.RoomID,
			err,
		)
		return
	}

	// Incoming is message.send, outgoing is message.created
	outgoing := OutgoingMessage{
		Type: TypeMessageCreated,
		Data: msg, // returned from the database
	}

	payload, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("Error outgoing message for message %d: %v", msg.ID, err)
		return
	}

	// Send to the broadcast channel so h.Run() is the only goroutine accessing h.rooms
	c.hub.broadcast <- BroadcastMessage{
		RoomID:  msg.RoomID,
		Message: payload,
	}
}

// This will handle the message of type "message.edit"
func (c *Client) handleMessageEdit(rawData json.RawMessage) {
	var data MessageEditData

	if err := json.Unmarshal(rawData, &data); err != nil {
		log.Printf("Invalid message.edit from user %d: %v", c.userID, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	msg, err := c.messages.UpdateMessage(
		ctx,
		data.MessageID,
		c.userID,
		data.Content,
	)
	if err != nil {
		log.Printf(
			"Error when updating message %d for user %d: %v",
			data.MessageID,
			c.userID,
			err,
		)
		return
	}

	outgoing := OutgoingMessage{
		Type: TypeMessageUpdated,
		Data: msg,
	}

	payload, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("Error outgoing message updated %d: %v", msg.ID, err)
		return
	}

	// Send through the Hub's broadcast channel
	c.hub.broadcast <- BroadcastMessage{
		RoomID:  msg.RoomID,
		Message: payload,
	}
}

// This will handle the message of type "message.delete"
func (c *Client) handleMessageDelete(rawData json.RawMessage) {
	var data MessageDeleteData

	if err := json.Unmarshal(rawData, &data); err != nil {
		log.Printf("Invalid message.delete from user %d: %v", c.userID, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// We need the message's room ID so that the delete event
	// can be broadcast to the correct room.
	msg, err := c.messages.GetMessageByID(
		ctx,
		data.MessageID,
	)
	if err != nil {
		log.Printf(
			"Error getting message %d for user %d: %v", data.MessageID, c.userID, err)
		return
	}

	err2 := c.messages.DeleteMessage(ctx, data.MessageID, c.userID)
	if err2 != nil {
		log.Printf(
			"Error delete message %d for user %d: %v", data.MessageID, c.userID, err2)
		return
	}

	outgoing := OutgoingMessage{
		Type: TypeMessageDeleted,
		Data: MessageDeleteData{
			MessageID: data.MessageID,
		},
	}

	payload, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("Error outgoing message.deleted for message %d: %v", data.MessageID, err)
		return
	}

	c.hub.broadcast <- BroadcastMessage{
		RoomID:  msg.RoomID,
		Message: payload,
	}
}
