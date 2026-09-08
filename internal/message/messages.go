package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotMessageSender = errors.New("Deletion denied. You did not send the message.")
var ErrEmptyMessage = errors.New("This message is blank. Please enter a message.")
var ErrMessageDeleted = errors.New("This message has already been deleted.")
var ErrMessagePermission = errors.New("You do not belong to this room.")

type Message struct {
	ID        int64      `json:"id"`
	RoomID    int64      `json:"room_id"`
	SenderID  int64      `json:"sender_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// DTO
type CreateMessageRequest struct {
	RoomID   int64  `json:"room_id"`
	Content  string `json:"content"`
}

type Service struct {
	store Store
}

type Store interface {
	CreateMessage(
		ctx context.Context,
		message *Message,
	) (*Message, error)

	GetMessageByID(
		ctx context.Context,
		id int64,
	) (*Message, error)

	ListMessagesByRoom(
		ctx context.Context,
		roomID int64,
		limit int,
	) ([]Message, error)

	UpdateMessage(
		ctx context.Context,
		messageID int64,
		content string,
	) (*Message, error)

	DeleteMessage(
		ctx context.Context,
		id int64,
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

func (s *Service) CreateMessage(ctx context.Context, m CreateMessageRequest, userID int64) (*Message, error) {

	// Check message is not blank
	if m.Content == "" {
		return nil, ErrEmptyMessage
	}

	// Remove extra whitespace
	content := strings.TrimSpace(m.Content)

	// Check the user is actually a member of the room
	inRoom, err := s.store.IsMember(ctx, m.RoomID, userID)
	if err != nil {
		return nil, fmt.Errorf("Error when checking room membership: %w", err)
	}
	if !inRoom {
		return nil, ErrMessagePermission
	}

	// Map the CreateMessageRequest to a Message
	message := &Message{
		RoomID:   m.RoomID,
		SenderID: userID,
		Content:  content,
	}

	return s.store.CreateMessage(ctx, message)
}

func (s *Service) GetMessageByID(ctx context.Context, id int64) (*Message, error) {
	return s.store.GetMessageByID(ctx, id)
}

func (s *Service) ListMessagesByRoom(ctx context.Context, roomID int64, userID int64, limit int) ([]Message, error) {
	inRoom, err := s.store.IsMember(ctx, roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("Error when checking room membership: %w", err)
	}
	if !inRoom {
		return nil, ErrMessagePermission
	}

	return s.store.ListMessagesByRoom(ctx, roomID, limit)
}

func (s *Service) UpdateMessage(ctx context.Context, messageID int64, userID int64, content string) (*Message, error) {
	if content == "" {
		return nil, ErrEmptyMessage
	}

	c := strings.TrimSpace(content)
	
	msg, err := s.store.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if msg.SenderID != userID {
		return nil, ErrNotMessageSender
	}
	
	return s.store.UpdateMessage(ctx, messageID, c)
}

func (s *Service) DeleteMessage(ctx context.Context, messageID int64, userID int64) error {
	msg, err := s.store.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return ErrNotMessageSender
	}

	if msg.DeletedAt != nil {
		return ErrMessageDeleted
	}
	
	return s.store.DeleteMessage(ctx, messageID)
}
