package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
)

type ChannelHandler struct {
	DB *gorm.DB
}

func NewChannelHandler(db *gorm.DB) *ChannelHandler {
	return &ChannelHandler{DB: db}
}

/* -------- CREATE CHANNEL -------- */

type createChannelRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (h *ChannelHandler) Create(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.GetString("user_id")

	// 1. Check membership + role
	var member models.ServerMember
	if err := h.DB.First(
		&member,
		"user_id = ? AND server_id = ?",
		userID,
		serverID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	if member.Role != "owner" && member.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// 2. Validate request
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel := models.Channel{
		ServerID: serverID,
		Name:     req.Name,
	}

	if err := h.DB.Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

/* -------- LIST CHANNELS -------- */

func (h *ChannelHandler) List(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.GetString("user_id")

	// Check membership
	var member models.ServerMember
	if err := h.DB.First(
		&member,
		"user_id = ? AND server_id = ?",
		userID,
		serverID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	var channels []models.Channel
	h.DB.Where("server_id = ?", serverID).Find(&channels)

	c.JSON(http.StatusOK, channels)
}
