package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"real-time-chat/internal/adapters/postgres"
	"real-time-chat/internal/adapters/redis"
	"real-time-chat/internal/config"
	"real-time-chat/internal/delivery/http_delivery"
	"real-time-chat/internal/delivery/ws_delivery"
	"real-time-chat/internal/services"
	"real-time-chat/internal/utils"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	dbPool, err := postgres.NewDBPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Unable to connect to redis: %v\n", err)
	}

	// Email Sender
	emailSender := utils.NewGomailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	// Repositories
	userRepo := postgres.NewPostgresUserRepository(dbPool)
	tokenRepo := redis.NewRedisTokenRepository(redisClient)
	convoRepo := postgres.NewPostgresConversationRepository(dbPool)
	messageRepo := postgres.NewPostgresMessageRepository(dbPool)
	friendshipRepo := postgres.NewPostgresFriendshipRepository(dbPool)
	groupRepo := postgres.NewPostgresGroupRepository(dbPool)
	gameRepo := postgres.NewPostgresGameRepository(dbPool)
	eventRepo := postgres.NewPostgresEventRepository(dbPool) // New event repository

	// Services
	tokenService := services.NewTokenService(userRepo, tokenRepo, cfg.JWTSecret, time.Hour*8, time.Hour*24*7, time.Minute*30) // OTP expiry 30 mins
	eventService := services.NewEventService(eventRepo, userRepo) // New event service
	userService := services.NewUserService(userRepo, tokenService, emailSender)
	convoService := services.NewConversationService(convoRepo, userRepo, groupRepo)
	messageService := services.NewMessageService(messageRepo, convoRepo, userRepo)
	friendshipService := services.NewFriendshipService(friendshipRepo, userRepo, convoService, eventService) // Pass eventService
	groupService := services.NewGroupService(groupRepo, userRepo, convoService, eventService) // Pass eventService
	gameService := services.NewGameService(gameRepo, userRepo, convoService, eventService) // Pass eventService

	// WebSocket Hub
	hub := ws_delivery.NewHub(messageService, convoService, gameService, eventService) // Pass eventService to Hub
	go hub.Run()

	// HTTP Handlers
	userHandler := http_delivery.NewUserHandler(userService, tokenService, cfg.JWTSecret, cfg.UploadDir)
	convoHandler := http_delivery.NewConversationHandler(convoService, messageService)
	friendshipHandler := http_delivery.NewFriendshipHandler(friendshipService)
	groupHandler := http_delivery.NewGroupHandler(groupService, convoService)
	gameHandler := http_delivery.NewGameHandler(gameService, hub)
	longPollingHandler := http_delivery.NewLongPollingHandler(eventService, userService) // Use eventService
	wsHandler := ws_delivery.NewWSHandler(hub, tokenService)

	// Create upload directory if it doesn't exist
	if _, err := os.Stat(cfg.UploadDir); os.IsNotExist(err) {
		os.Mkdir(cfg.UploadDir, 0755)
	}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Serve static files (profile pictures)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, cfg.UploadDir))
	http_delivery.FileServer(r, "/uploads", filesDir)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Real-Time Chat API"))
	})

	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/register", userHandler.Register)
		r.Get("/verify-email", userHandler.VerifyEmail) // GET for direct link click
		r.Post("/verify-email", userHandler.VerifyEmail) // POST for manual token entry
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.RefreshToken)
		r.Post("/request-password-reset", userHandler.RequestPasswordReset)
		r.Post("/reset-password", userHandler.ResetPassword)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(http_delivery.AuthMiddleware(cfg.JWTSecret, userService))
			r.Get("/me", userHandler.GetMe)
			r.Post("/me/avatar", userHandler.UploadProfilePicture) // Profile picture upload

			// Conversation & Message Routes
			r.Get("/conversations", convoHandler.GetUserConversations)
			r.Get("/conversations/{conversationID}/messages", convoHandler.GetMessages)
			r.Post("/conversations/{conversationID}/read", convoHandler.MarkAsRead)
			r.Delete("/conversations/{conversationID}", convoHandler.DeleteOneToOneConversation) // Delete 1-1 chat

			// Friendship Routes
			r.Post("/friends/requests", friendshipHandler.SendRequest)
			r.Get("/friends/requests", friendshipHandler.GetRequests)
			r.Put("/friends/requests/{requestID}", friendshipHandler.RespondToRequest)
			r.Get("/friends", friendshipHandler.GetFriends)

			// Group Routes
			r.Post("/groups", groupHandler.CreateGroup)
			r.Get("/groups/{groupID}", groupHandler.GetGroupDetails)
			r.Post("/groups/{groupID}/join", groupHandler.JoinGroup)
			r.Post("/groups/{groupID}/leave", groupHandler.LeaveGroup)
			r.Delete("/groups/{groupID}/members/{memberID}", groupHandler.RemoveGroupMember)
			r.Delete("/groups/{groupID}", groupHandler.DeleteGroup)

			// Game Routes
			r.Post("/games/invite", gameHandler.InviteToGame)
			r.Post("/games/{gameID}/respond", gameHandler.RespondToGameInvite)
			r.Get("/games/{gameID}", gameHandler.GetGameState)
			r.Post("/games/{gameID}/move", gameHandler.MakeMove)
		})
	})

	// WebSocket route
	r.Get("/ws", wsHandler.HandleWebSocket)

	// Long polling fallback endpoint
	r.Get("/api/events", longPollingHandler.HandleLongPolling)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}
	log.Println("Server exited properly")
}
