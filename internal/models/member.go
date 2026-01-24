package models

import "time"

type ServerMember struct {
	UserID   string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	ServerID string    `gorm:"type:uuid;primaryKey" json:"server_id"`
	Role     string    `gorm:"not null" json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
