package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
	ws "github.com/SelimArslan1/chat-app/internal/websocket"
	jwtutil "github.com/SelimArslan1/chat-app/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate against allowed origins
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// Default: allow same-origin and localhost in development
			origin := r.Header.Get("Origin")
			return origin == "" ||
				strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "https://localhost")
		}
		origin := r.Header.Get("Origin")
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},
}

type WebSocketHandler struct {
	DB  *gorm.DB
	Hub *ws.Hub
}

func NewWebSocketHandler(db *gorm.DB, hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{DB: db, Hub: hub}
}

func (h *WebSocketHandler) Handle(c *gin.Context) {
	token := c.Query("token")
	claims, err := jwtutil.ParseToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	channelID := c.Query("channel_id")
	if channelID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "channel_id required"})
		return
	}

	var channel models.Channel
	if err := h.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var member models.ServerMember
	if err := h.DB.First(
		&member,
		"user_id = ? AND server_id = ?",
		claims.UserID,
		channel.ServerID,
	).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no access to this channel"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    claims.UserID,
		ChannelID: channelID,
		Hub:       h.Hub,
	}

	h.Hub.Register <- client
	go client.WritePump()

	client.ReadPump(func(c *ws.Client, event ws.ClientEvent) {
		if event.Type != "SEND_MESSAGE" {
			return
		}

		msg := models.Message{
			UserID:    c.UserID,
			ChannelID: c.ChannelID,
			Content:   event.Content,
		}

		if err := h.DB.Create(&msg).Error; err != nil {
			return
		}

		out, _ := json.Marshal(ws.ServerEvent{
			Type:    "NEW_MESSAGE",
			Payload: msg,
		})

		h.Hub.Broadcast <- ws.Broadcast{
			ChannelID: c.ChannelID,
			Message:   out,
		}
	})
}
