package main

import (
	"context"
	"log"
	"net/http"

	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"gochat/internal/auth"
	"gochat/internal/cache"
	"gochat/internal/database"
	"gochat/internal/handler"
	"gochat/internal/message"
	"gochat/internal/room"
	"gochat/internal/store"
	"gochat/internal/user"
	"gochat/internal/websocket"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	// Redis
	redisCache := cache.NewRedis("redis:6379")
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	defer redisCache.Close()

	// Database connection pool
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Concrete database store
	dbStore := store.New(pool, redisCache)

	// JWT manager
	jwtManager := auth.NewJWTManager(
		jwtSecret,
		time.Hour,
	)

	// Authentication middleware
	authMiddleware := auth.NewMiddleware(jwtManager)

	// Application services
	userService := user.NewService(dbStore)
	roomService := room.NewService(dbStore)
	messageService := message.NewService(dbStore)
	authService := auth.NewService(dbStore, jwtManager)

	// HTTP handlers
	userHandler := handler.NewUserHandler(userService)
	roomHandler := handler.NewRoomHandler(roomService)
	messageHandler := handler.NewMessageHandler(messageService)
	authHandler := handler.NewAuthHandler(authService)

	// WebSocket
	hub := websocket.NewHub()
	go hub.Run()                                                        // create the hub and start its background event loop
	wsHandler := websocket.NewHandler(roomService, hub, messageService) // create the handler

	// Router
	mux := mount(userHandler, roomHandler, messageHandler, authHandler, authMiddleware, wsHandler)

	// HTTP server
	if err := run(mux); err != nil {
		log.Fatal(err)
	}
}

func mount(userHandler *handler.UserHandler,
	roomHandler *handler.RoomHandler,
	messageHandler *handler.MessageHandler,
	authHandler *handler.AuthHandler,
	authMiddleware *auth.Middleware,
	wsHandler *websocket.Handler,
) http.Handler {
	//mux := http.NewServeMux()

	r := chi.NewRouter()

	// Do not require authentication
	r.Post("/api/v1/users/register", userHandler.CreateUser)
	r.Post("/api/v1/users/login", authHandler.Login)

	// Require authentication
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Users
		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/{id}", userHandler.GetUserByID)
			r.Get("/username/{username}", userHandler.GetUserByUsername)
		})

		// Rooms
		r.Route("/api/v1/rooms", func(r chi.Router) {
			r.Post("/", roomHandler.CreateRoom)
			r.Get("/all", roomHandler.ListRooms)
			r.Get("/users", roomHandler.ListRoomsByUser)
			r.Get("/{id}", roomHandler.GetRoomByID)
			r.Delete("/{id}", roomHandler.DeleteRoom)
			r.Post("/{id}/members", roomHandler.AddMember)
			r.Get("/{id}/members", roomHandler.ListMembersByRoom)
			r.Get("/{id}/members/{userID}", roomHandler.IsMember)
			r.Delete("/{id}/members/{userID}", roomHandler.RemoveMember)
		})

		// Messages
		r.Route("/api/v1/messages", func(r chi.Router) {
			r.Post("/", messageHandler.CreateMessage)
			r.Get("/{id}", messageHandler.GetMessageByID)
			r.Get("/rooms/{roomID}", messageHandler.ListMessagesByRoom)
			r.Patch("/{id}", messageHandler.UpdateMessage)
			r.Delete("/{id}", messageHandler.DeleteMessage)
		})

		// WebSocket
		r.Get("/api/v1/ws", wsHandler.ServeWS)

	})

	return r
}

func run(mux http.Handler) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return server.ListenAndServe()
}
