package repository

import (
	"errors"
	"officeworker/models"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) ListByUserID(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) ListActiveByUserID(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("updated_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) FindByIDAndUserID(id, userID uint) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Delete(session *models.Session) error {
	return r.db.Delete(session).Error
}

func (r *SessionRepository) Update(session *models.Session) error {
	return r.db.Save(session).Error
}

func (r *SessionRepository) DeleteWithRelations(session *models.Session) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", session.ID).Delete(&models.AgentTask{}).Error; err != nil {
			return err
		}

		if err := tx.Where("session_id = ?", session.ID).Delete(&models.File{}).Error; err != nil {
			return err
		}

		return tx.Delete(session).Error
	})
}
