package models

import (
	"time"
)

type Session struct {
	BaseModel
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	AgentID     string    `gorm:"type:varchar(100);not null;index" json:"agent_id"`
	ContainerID string    `gorm:"type:varchar(100);not null;index" json:"container_id"`
	Status      string    `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	Title       string    `gorm:"type:varchar(255)" json:"title"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
}

func (Session) TableName() string {
	return "sessions"
}
