package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/SelimArslan1/chat-app/internal/models"
	jwtutil "github.com/SelimArslan1/chat-app/pkg/jwt"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	accessToken, err := jwtutil.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Create refresh token in DB
	refreshTokenStr := generateSecureToken()
	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	h.DB.Create(&refreshToken)

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshTokenStr,
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := jwtutil.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Create refresh token in DB
	refreshTokenStr := generateSecureToken()
	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	h.DB.Create(&refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshTokenStr,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find refresh token in DB
	var storedToken models.RefreshToken
	if err := h.DB.Where("token = ?", req.RefreshToken).First(&storedToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Check if token is valid
	if !storedToken.IsValid() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired or revoked"})
		return
	}

	// Revoke old token
	now := time.Now()
	h.DB.Model(&storedToken).Update("revoked_at", now)

	// Verify user still exists
	var user models.User
	if err := h.DB.First(&user, "id = ?", storedToken.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Generate new tokens
	accessToken, err := jwtutil.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Create new refresh token
	refreshTokenStr := generateSecureToken()
	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	h.DB.Create(&refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshTokenStr,
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Revoke the refresh token
	now := time.Now()
	result := h.DB.Model(&models.RefreshToken{}).
		Where("token = ? AND revoked_at IS NULL", req.RefreshToken).
		Update("revoked_at", now)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "already logged out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// LogoutAll revokes all refresh tokens for the user
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := c.GetString("user_id")

	now := time.Now()
	h.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)

	c.JSON(http.StatusOK, gin.H{"status": "logged out from all devices"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := h.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}
