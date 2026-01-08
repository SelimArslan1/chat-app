package models

import "time"

type Channel struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ServerID  string `gorm:"type:uuid;not null"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
}
