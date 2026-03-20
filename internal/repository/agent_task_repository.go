package repository

import (
	"officeworker/models"

	"gorm.io/gorm"
)

type AgentTaskRepository struct {
	db *gorm.DB
}

func NewAgentTaskRepository(db *gorm.DB) *AgentTaskRepository {
	return &AgentTaskRepository{db: db}
}

func (r *AgentTaskRepository) Create(task *models.AgentTask) error {
	return r.db.Create(task).Error
}
