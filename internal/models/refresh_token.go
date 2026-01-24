package models

import "time"

type RefreshToken struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"unique;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && time.Now().Before(rt.ExpiresAt)
}
