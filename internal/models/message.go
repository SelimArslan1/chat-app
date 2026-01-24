package models

import "time"

type Message struct {
	ID        string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ChannelID string     `gorm:"type:uuid;index;not null" json:"channel_id"`
	UserID    string     `gorm:"type:uuid;index" json:"user_id"`
	Content   string     `gorm:"not null" json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
