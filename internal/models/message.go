package models

import "time"

type Message struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ChannelID string `gorm:"type:uuid;index"`
	UserID    string `gorm:"type:uuid"`
	Content   string `gorm:"not null"`
	CreatedAt time.Time
}
