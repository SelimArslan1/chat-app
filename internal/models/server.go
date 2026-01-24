package models

import "time"

type Server struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	OwnerID   string    `gorm:"type:uuid;not null" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}
