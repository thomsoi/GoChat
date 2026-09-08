package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gochat/internal/auth"
	"gochat/internal/room"
)

type RoomHandler struct {
	roomService *room.Service
}

func NewRoomHandler(rooms *room.Service) *RoomHandler {
	return &RoomHandler{
		roomService: rooms,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req room.CreateRoomRequest

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

	room, err := h.roomService.CreateRoom(r.Context(), req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	rm, err := h.roomService.GetRoomByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(rm)
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.ListRooms(r.Context())
	if err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) ListRoomsByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Error unauthorized, cannot obtain userID", http.StatusUnauthorized)
		return
	}

	rooms, err := h.roomService.ListRoomsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user's rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) ListMembersByRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	members, err := h.roomService.ListMembersByRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "Failed to get room members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(members)
}

func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("id")
	roomID, err := strconv.ParseInt(roomIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Ensure only the room owner can delete, so obtain the user ID to check
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized could not obtain userID", http.StatusUnauthorized)
		return
	}

	if err := h.roomService.DeleteRoom(r.Context(), roomID, userID); err != nil {
		http.Error(w, "Failed to delete room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddMember adds a user to a room with the specified role.
func (h *RoomHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	var req room.AddMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.roomService.AddMember(r.Context(), roomID, req.UserID, req.Role); err != nil {
		http.Error(w, "Failed to add room member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember removes a user from a room.
func (h *RoomHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.roomService.RemoveMember(
		r.Context(),
		roomID,
		userID,
	); err != nil {
		http.Error(w, "Failed to remove room member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// IsMember checks whether a user is a member of a room.
func (h *RoomHandler) IsMember(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	isMember, err := h.roomService.IsMember(
		r.Context(),
		roomID,
		userID,
	)

	if err != nil {
		http.Error(w, "Failed to check room membership", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]bool{
		"is_member": isMember,
	})
}
