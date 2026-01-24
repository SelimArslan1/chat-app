package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
)

type InviteHandler struct {
	DB *gorm.DB
}

func NewInviteHandler(db *gorm.DB) *InviteHandler {
	return &InviteHandler{DB: db}
}

// Generate a random invite code
func generateInviteCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

/* -------- CREATE INVITE -------- */

type createInviteRequest struct {
	MaxUses   int    `json:"max_uses"`   // 0 = unlimited
	ExpiresIn int    `json:"expires_in"` // hours, 0 = never
}

func (h *InviteHandler) Create(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.GetString("user_id")

	// Check membership + role
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

	var req createInviteRequest
	c.ShouldBindJSON(&req) // optional params

	invite := models.ServerInvite{
		ServerID:  serverID,
		Code:      generateInviteCode(),
		CreatedBy: userID,
		MaxUses:   req.MaxUses,
	}

	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		invite.ExpiresAt = &expiresAt
	}

	if err := h.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invite"})
		return
	}

	c.JSON(http.StatusCreated, invite)
}

/* -------- LIST SERVER INVITES -------- */

func (h *InviteHandler) List(c *gin.Context) {
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

	var invites []models.ServerInvite
	h.DB.Where("server_id = ?", serverID).Find(&invites)

	c.JSON(http.StatusOK, invites)
}

/* -------- JOIN SERVER WITH INVITE CODE -------- */

type joinRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *InviteHandler) Join(c *gin.Context) {
	userID := c.GetString("user_id")

	var req joinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find invite
	var invite models.ServerInvite
	if err := h.DB.Where("code = ?", req.Code).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid invite code"})
		return
	}

	// Check if expired
	if invite.ExpiresAt != nil && time.Now().After(*invite.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "invite has expired"})
		return
	}

	// Check if max uses reached
	if invite.MaxUses > 0 && invite.Uses >= invite.MaxUses {
		c.JSON(http.StatusGone, gin.H{"error": "invite has reached max uses"})
		return
	}

	// Check if already a member
	var existingMember models.ServerMember
	if err := h.DB.First(
		&existingMember,
		"user_id = ? AND server_id = ?",
		userID,
		invite.ServerID,
	).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already a member of this server"})
		return
	}

	// Add member and increment uses in transaction
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		member := models.ServerMember{
			UserID:   userID,
			ServerID: invite.ServerID,
			Role:     "member",
			JoinedAt: time.Now(),
		}

		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		return tx.Model(&invite).Update("uses", invite.Uses+1).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join server"})
		return
	}

	// Return the server info
	var server models.Server
	h.DB.First(&server, "id = ?", invite.ServerID)

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully joined server",
		"server":  server,
	})
}

/* -------- DELETE INVITE -------- */

func (h *InviteHandler) Delete(c *gin.Context) {
	inviteID := c.Param("invite_id")
	userID := c.GetString("user_id")

	var invite models.ServerInvite
	if err := h.DB.First(&invite, "id = ?", inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}

	// Check if user has permission
	var member models.ServerMember
	if err := h.DB.First(
		&member,
		"user_id = ? AND server_id = ?",
		userID,
		invite.ServerID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a server member"})
		return
	}

	if member.Role != "owner" && member.Role != "admin" && invite.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	h.DB.Delete(&invite)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
