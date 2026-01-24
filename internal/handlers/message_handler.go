package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
)

type MessageHandler struct {
	DB *gorm.DB
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{DB: db}
}

// POST /channels/:id/messages
type createMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

func (h *MessageHandler) Create(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("user_id")

	var channel models.Channel
	if err := h.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var member models.ServerMember
	if err := h.DB.First(&member,
		"user_id = ? AND server_id = ?",
		userID,
		channel.ServerID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := models.Message{
		ChannelID: channelID,
		UserID:    userID,
		Content:   req.Content,
	}

	if err := h.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

// GET /channels/:id/messages
func (h *MessageHandler) List(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.GetString("user_id")

	// pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 100 {
		limit = 100
	}

	before := c.Query("before")

	// membership check
	var channel models.Channel
	if err := h.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var member models.ServerMember
	if err := h.DB.First(&member,
		"user_id = ? AND server_id = ?",
		userID,
		channel.ServerID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	query := h.DB.
		Where("channel_id = ? AND deleted_at IS NULL", channelID).
		Order("created_at DESC").
		Limit(limit)

	if before != "" {
		var beforeMsg models.Message
		if err := h.DB.First(&beforeMsg, "id = ?", before).Error; err == nil {
			query = query.Where("created_at < ?", beforeMsg.CreatedAt)
		}
	}

	var messages []models.Message
	query.Find(&messages)

	// Populate usernames
	userIDs := make([]string, 0, len(messages))
	for _, msg := range messages {
		userIDs = append(userIDs, msg.UserID)
	}

	var users []models.User
	h.DB.Where("id IN ?", userIDs).Find(&users)

	userMap := make(map[string]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	for i := range messages {
		messages[i].Username = userMap[messages[i].UserID]
	}

	c.JSON(http.StatusOK, messages)
}

// DELETE /messages/:id
func (h *MessageHandler) Delete(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.GetString("user_id")

	var msg models.Message
	if err := h.DB.First(&msg, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	var channel models.Channel
	h.DB.First(&channel, "id = ?", msg.ChannelID)

	var member models.ServerMember
	if err := h.DB.First(&member,
		"user_id = ? AND server_id = ?",
		userID,
		channel.ServerID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	// permission check
	if msg.UserID != userID && member.Role != "owner" && member.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete message"})
		return
	}

	now := time.Now()
	msg.DeletedAt = &now

	h.DB.Save(&msg)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
