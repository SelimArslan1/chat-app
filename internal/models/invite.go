package models

import "time"

type ServerInvite struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ServerID  string    `gorm:"type:uuid;not null;index" json:"server_id"`
	Code      string    `gorm:"unique;not null;size:8" json:"code"`
	CreatedBy string    `gorm:"type:uuid;not null" json:"created_by"`
	MaxUses   int       `gorm:"default:0" json:"max_uses"`  // 0 = unlimited
	Uses      int       `gorm:"default:0" json:"uses"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
