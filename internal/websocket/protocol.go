package websocket

/*
The WebSocket connection is effectively a pipe carrying bytes between the browser
and the server. It is a permanent connection. For HTTP requests, the URL gives information
about what the request means (the HTTP verb, the path, and any query parameters or path variables)
But a WebSocket does not, so this file is here so we can
provide some context to a WebSocket message
*/

import "encoding/json"

const (
	TypeMessageSend   = "message.send" // Client -> Server, user is sending a message
	TypeMessageEdit   = "message.edit"
	TypeMessageDelete = "message.delete"

	TypeMessageCreated = "message.created" // Server -> Client, message has been created
	TypeMessageUpdated = "message.updated"
	TypeMessageDeleted = "message.deleted"

	TypeError = "error"
)

// Represents a message sent by the client
type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Represents a message sent by the server
type OutgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// The payload for message.send
type MessageSendData struct {
	RoomID int64 `json:"room_id"`
	Content string `json:"content"`
}

// The payload for message.edit
type MessageEditData struct {
	MessageID int64  `json:"message_id"`
	Content   string `json:"content"`
}

// The payload for message.delete
type MessageDeleteData struct {
	MessageID int64  `json:"message_id"`
}
