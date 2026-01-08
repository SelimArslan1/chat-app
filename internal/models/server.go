package models

import "time"

type Server struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string `gorm:"not null"`
	OwnerID   string `gorm:"type:uuid;not null"`
	CreatedAt time.Time
}
