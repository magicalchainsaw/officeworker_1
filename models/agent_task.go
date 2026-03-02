package models

import (
	"time"
)

type AgentTask struct {
	BaseModel
	SessionID uint      `gorm:"not null;index" json:"session_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"`
	Payload   string    `gorm:"type:text" json:"payload"`
	Status    string    `gorm:"type:varchar(20);default:'pending';not null;index" json:"status"`
	Result    string    `gorm:"type:text" json:"result,omitempty"`
	Error     string    `gorm:"type:text" json:"error,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (AgentTask) TableName() string {
	return "agent_tasks"
}
