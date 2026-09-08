package room

import (
	"context"
	"time"
	"errors"
)

var ErrNotRoomOwner = errors.New("Deletion denied. You are not the owner of the roon.")

type Room struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// For simplicty, membership will not have its own file
// and will be put with room instead.
type Membership struct {
	RoomID   int64
	UserID   int64
	Role     string
	JoinedAt time.Time
}

// This represents a member of a room
type RoomMember struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// DTOs
type CreateRoomRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type Service struct {
	store Store
}

type Store interface {
	CreateRoom(
		ctx context.Context,
		room *Room,
	) (*Room, error)

	GetRoomByID(
		ctx context.Context,
		id int64,
	) (*Room, error)

	ListRooms(
		ctx context.Context,
	) ([]Room, error)

	ListRoomsByUser(
		ctx context.Context,
		userID int64,
	) ([]Room, error)

	ListMembersByRoom(
		ctx context.Context,
		roomID int64,
	) ([]RoomMember, error)

	DeleteRoom(
		ctx context.Context,
		id int64,
	) error

	AddMember(
		ctx context.Context,
		roomID int64,
		userID int64,
		role string,
	) error

	RemoveMember(
		ctx context.Context,
		roomID int64,
		userID int64,
	) error

	IsMember(
		ctx context.Context,
		roomID int64,
		userID int64,
	) (bool, error)
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// Creates a new room
func (s *Service) CreateRoom(ctx context.Context, r CreateRoomRequest, userID int64) (*Room, error) {
	// Map the CreateRoomRequest to a Room
	room := &Room{
		Name:      r.Name,
		CreatedBy: userID,
	}

	rm, err := s.store.CreateRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	if err := s.store.AddMember(ctx, rm.ID, userID, "owner"); err != nil {
		return nil, err
	}

	return rm, nil
}

// Obtains a room by its ID
func (s *Service) GetRoomByID(ctx context.Context, id int64) (*Room, error) {
	return s.store.GetRoomByID(ctx, id)
}

// Obtains all rooms
func (s *Service) ListRooms(ctx context.Context) ([]Room, error) {
	return s.store.ListRooms(ctx)
}

// Deletes a room by its ID
func (s *Service) DeleteRoom(ctx context.Context, roomID int64, userID int64) error {
	rm, err := s.store.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if rm.CreatedBy != userID {
		return ErrNotRoomOwner
	}
	
	return s.store.DeleteRoom(ctx, roomID)
}

func (s *Service) ListRoomsByUser(ctx context.Context, id int64) ([]Room, error) {
	return s.store.ListRoomsByUser(ctx, id)
}

func (s *Service) ListMembersByRoom(ctx context.Context, id int64) ([]RoomMember, error) {
	return s.store.ListMembersByRoom(ctx, id)
}

// Adds a user to a room with the specified role
func (s *Service) AddMember(
	ctx context.Context,
	roomID int64,
	userID int64,
	role string,
) error {
	return s.store.AddMember(ctx, roomID, userID, role)
}

// Removes a user from a room
func (s *Service) RemoveMember(
	ctx context.Context,
	roomID int64,
	userID int64,
) error {
	return s.store.RemoveMember(ctx, roomID, userID)
}

// Checks whether a user is a member of a room
func (s *Service) IsMember(
	ctx context.Context,
	roomID int64,
	userID int64,
) (bool, error) {
	return s.store.IsMember(ctx, roomID, userID)
}
