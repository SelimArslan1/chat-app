package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/SelimArslan1/chat-app/internal/db"
	"github.com/SelimArslan1/chat-app/internal/handlers"
	"github.com/SelimArslan1/chat-app/internal/middleware"
	"github.com/SelimArslan1/chat-app/internal/models"
	"github.com/SelimArslan1/chat-app/internal/websocket"
)

func main() {
	_ = godotenv.Load()

	database := db.Connect()
	log.Println("DB connected:", database != nil)

	err := database.AutoMigrate(
		&models.User{},
		&models.Server{},
		&models.Channel{},
		&models.Message{},
		&models.ServerMember{},
		&models.ServerInvite{},
	)

	if err != nil {
		log.Fatal("auto-migrate failed:", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(database)
	serverHandler := handlers.NewServerHandler(database)
	channelHandler := handlers.NewChannelHandler(database)
	messageHandler := handlers.NewMessageHandler(database)
	inviteHandler := handlers.NewInviteHandler(database)

	hub := websocket.NewHub()
	go hub.Run()

	wsHandler := handlers.NewWebSocketHandler(database, hub)

	r.GET("/ws", wsHandler.Handle)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.GET("/me", middleware.AuthRequired(), authHandler.Me)
	}

	servers := r.Group("/servers", middleware.AuthRequired())
	{
		servers.POST("", serverHandler.Create)
		servers.GET("", serverHandler.List)
		servers.GET("/:id", serverHandler.Get)

		servers.POST("/:id/channels", channelHandler.Create)
		servers.GET("/:id/channels", channelHandler.List)

		// Invite routes
		servers.POST("/:id/invites", inviteHandler.Create)
		servers.GET("/:id/invites", inviteHandler.List)
		servers.DELETE("/:id/invites/:invite_id", inviteHandler.Delete)
	}

	// Join server with invite code
	invites := r.Group("/invites", middleware.AuthRequired())
	{
		invites.POST("/join", inviteHandler.Join)
	}

	channels := r.Group("/channels", middleware.AuthRequired())
	{
		channels.POST("/:id/messages", messageHandler.Create)
		channels.GET("/:id/messages", messageHandler.List)
	}

	messages := r.Group("/messages", middleware.AuthRequired())
	{
		messages.DELETE("/:id", messageHandler.Delete)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
