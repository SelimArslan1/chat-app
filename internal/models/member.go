package models

import "time"

type ServerMember struct {
	UserID   string `gorm:"type:uuid;primaryKey"`
	ServerID string `gorm:"type:uuid;primaryKey"`
	Role     string `gorm:"not null"`
	JoinedAt time.Time
}
