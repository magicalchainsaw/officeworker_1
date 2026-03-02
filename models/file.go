package models

import (
	"time"
)

type File struct {
	BaseModel
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	SessionID   *uint     `gorm:"index" json:"session_id,omitempty"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Path        string    `gorm:"type:varchar(500);not null" json:"path"`
	Size        int64     `gorm:"not null" json:"size"`
	Type        string    `gorm:"type:varchar(100)" json:"type"`
	Hash        string    `gorm:"type:varchar(64);index" json:"hash"`
	Status      string    `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (File) TableName() string {
	return "files"
}
