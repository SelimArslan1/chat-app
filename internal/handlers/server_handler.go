package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
)

type ServerHandler struct {
	DB *gorm.DB
}

func NewServerHandler(db *gorm.DB) *ServerHandler {
	return &ServerHandler{DB: db}
}

/* -------- CREATE SERVER -------- */

type createServerRequest struct {
	Name string `json:"name" binding:"required,min=3"`
}

func (h *ServerHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req createServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server := models.Server{
		Name:    req.Name,
		OwnerID: userID,
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&server).Error; err != nil {
			return err
		}

		member := models.ServerMember{
			UserID:   userID,
			ServerID: server.ID,
			Role:     "owner",
		}

		return tx.Create(&member).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, server)
}

/* -------- LIST USER SERVERS -------- */

func (h *ServerHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	var servers []models.Server
	h.DB.
		Joins("JOIN server_members ON server_members.server_id = servers.id").
		Where("server_members.user_id = ?", userID).
		Find(&servers)

	c.JSON(http.StatusOK, servers)
}

/* -------- GET SERVER -------- */

func (h *ServerHandler) Get(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.GetString("user_id")

	var member models.ServerMember
	if err := h.DB.First(&member,
		"user_id = ? AND server_id = ?", userID, serverID,
	).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var server models.Server
	h.DB.First(&server, "id = ?", serverID)

	c.JSON(http.StatusOK, server)
}
