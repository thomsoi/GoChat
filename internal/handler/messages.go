package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gochat/internal/auth"
	"gochat/internal/message"
)

const getMessagesByRoomLimit = 100

type MessageHandler struct {
	messageService *message.Service
}

func NewMessageHandler(messages *message.Service) *MessageHandler {
	return &MessageHandler{messageService: messages}
}

// Creates a new message
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req message.CreateMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract the user ID that made the request
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Error unauthorized cannot obtain userID", http.StatusUnauthorized)
		return
	}

	msg, err := h.messageService.CreateMessage(r.Context(), req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// Gets a message by its ID
func (h *MessageHandler) GetMessageByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	msg, err := h.messageService.GetMessageByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(msg)
}

// Gets the messages (bounded by a limit) that belong to a specified room
func (h *MessageHandler) ListMessagesByRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("roomID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	/* This block is if the client specifies the limit */
	// if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
	// 	limit, err = strconv.Atoi(limitParam)
	// 	if err != nil || limit <= 0 {
	// 		http.Error(w, "Invalid limit", http.StatusBadRequest)
	// 		return
	// 	}

	// 	if limit > 100 {
	// 		limit = 100
	// 	}
	// }

	// Extract the user ID that made the request (cannot read other rooms messages)
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Error unauthorized cannot obtain userID", http.StatusUnauthorized)
		return
	}

	messages, err := h.messageService.ListMessagesByRoom(r.Context(), roomID, userID, getMessagesByRoomLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(messages)
}

// Updates the content of an existing message
func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Body of this request
	// TODO: Make this a DTO
	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract the user ID that made the request to check the message's sender ID matches (i.e. update only their own message)
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Error unauthorized cannot obtain userID", http.StatusUnauthorized)
		return
	}

	msg, err := h.messageService.UpdateMessage(r.Context(), id, userID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(msg)
}

// DeleteMessage soft-deletes a message.
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Ensure only the message sender can delete, so obtain the user ID to check
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized could not obtain userID", http.StatusUnauthorized)
		return
	}

	if err := h.messageService.DeleteMessage(r.Context(), roomID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
