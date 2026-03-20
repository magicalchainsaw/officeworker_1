package service

import (
	"errors"
	"officeworker/internal/repository"
	"officeworker/models"
	"strings"
	"time"
)

const (
	defaultSessionAgentID   = "default-agent"
	defaultSessionStatus    = "active"
	inactiveSessionStatus   = "inactive"
	defaultSessionTTL       = 2 * time.Hour
	pendingSessionContainer = "pending"
)

type SessionService struct {
	sessionRepo   sessionRepository
	agentTaskRepo agentTaskRepository
}

type sessionRepository interface {
	Create(session *models.Session) error
	ListByUserID(userID uint) ([]models.Session, error)
	ListActiveByUserID(userID uint) ([]models.Session, error)
	FindByIDAndUserID(id, userID uint) (*models.Session, error)
	Delete(session *models.Session) error
	DeleteWithRelations(session *models.Session) error
	Update(session *models.Session) error
}

type agentTaskRepository interface {
	Create(task *models.AgentTask) error
}

func NewSessionService(sessionRepo *repository.SessionRepository, agentTaskRepo *repository.AgentTaskRepository) *SessionService {
	return &SessionService{
		sessionRepo:   sessionRepo,
		agentTaskRepo: agentTaskRepo,
	}
}

type CreateSessionRequest struct {
	Title   string `json:"title" binding:"required,max=255"`
	AgentID string `json:"agent_id" binding:"omitempty,max=100"`
}

type UpdateSessionRequest struct {
	Title string `json:"title" binding:"required,max=255"`
}

type SendMessageRequest struct {
	Message string `json:"message" binding:"required,max=5000"`
}

type SessionInfo struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	AgentID     string    `json:"agent_id"`
	ContainerID string    `json:"container_id"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AgentTaskInfo struct {
	ID          uint       `json:"id"`
	SessionID   uint       `json:"session_id"`
	Type        string     `json:"type"`
	Payload     string     `json:"payload"`
	Status      string     `json:"status"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SendMessageResponse struct {
	Session *SessionInfo   `json:"session"`
	Task    *AgentTaskInfo `json:"task"`
	Reply   string         `json:"reply"`
}

func (s *SessionService) Create(userID uint, req *CreateSessionRequest) (*SessionInfo, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}

	agentID := strings.TrimSpace(req.AgentID)
	if agentID == "" {
		agentID = defaultSessionAgentID
	}

	session := &models.Session{
		UserID:      userID,
		AgentID:     agentID,
		ContainerID: pendingSessionContainer,
		Status:      defaultSessionStatus,
		Title:       title,
		ExpiresAt:   time.Now().UTC().Add(defaultSessionTTL),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return toSessionInfo(session), nil
}

func (s *SessionService) List(userID uint) ([]*SessionInfo, error) {
	sessions, err := s.sessionRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	items := make([]*SessionInfo, 0, len(sessions))
	for i := range sessions {
		items = append(items, toSessionInfo(&sessions[i]))
	}

	return items, nil
}

func (s *SessionService) ListActive(userID uint) ([]*SessionInfo, error) {
	sessions, err := s.sessionRepo.ListActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	items := make([]*SessionInfo, 0, len(sessions))
	for i := range sessions {
		items = append(items, toSessionInfo(&sessions[i]))
	}

	return items, nil
}

func (s *SessionService) Get(userID, sessionID uint) (*SessionInfo, error) {
	session, err := s.sessionRepo.FindByIDAndUserID(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	return toSessionInfo(session), nil
}

func (s *SessionService) Delete(userID, sessionID uint) error {
	session, err := s.sessionRepo.FindByIDAndUserID(sessionID, userID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	return s.sessionRepo.DeleteWithRelations(session)
}

func (s *SessionService) Update(userID, sessionID uint, req *UpdateSessionRequest) (*SessionInfo, error) {
	session, err := s.sessionRepo.FindByIDAndUserID(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}

	session.Title = title
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	return toSessionInfo(session), nil
}

func (s *SessionService) Activate(userID, sessionID uint) (*SessionInfo, error) {
	return s.updateStatus(userID, sessionID, defaultSessionStatus)
}

func (s *SessionService) Deactivate(userID, sessionID uint) (*SessionInfo, error) {
	return s.updateStatus(userID, sessionID, inactiveSessionStatus)
}

func (s *SessionService) SendMessage(userID, sessionID uint, req *SendMessageRequest) (*SendMessageResponse, error) {
	session, err := s.sessionRepo.FindByIDAndUserID(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}
	if session.Status != defaultSessionStatus {
		return nil, errors.New("session is not active")
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, errors.New("message is required")
	}

	now := time.Now().UTC()
	task := &models.AgentTask{
		SessionID:   session.ID,
		Type:        "message",
		Payload:     message,
		Status:      "completed",
		Result:      "收到消息了",
		StartedAt:   &now,
		CompletedAt: &now,
	}
	if err := s.agentTaskRepo.Create(task); err != nil {
		return nil, err
	}

	session.ExpiresAt = now.Add(defaultSessionTTL)
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	return &SendMessageResponse{
		Session: toSessionInfo(session),
		Task:    toAgentTaskInfo(task),
		Reply:   task.Result,
	}, nil
}

func (s *SessionService) updateStatus(userID, sessionID uint, status string) (*SessionInfo, error) {
	session, err := s.sessionRepo.FindByIDAndUserID(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	session.Status = status
	if status == defaultSessionStatus {
		session.ExpiresAt = time.Now().UTC().Add(defaultSessionTTL)
	}
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	return toSessionInfo(session), nil
}

func toSessionInfo(session *models.Session) *SessionInfo {
	return &SessionInfo{
		ID:          session.ID,
		UserID:      session.UserID,
		AgentID:     session.AgentID,
		ContainerID: session.ContainerID,
		Status:      session.Status,
		Title:       session.Title,
		ExpiresAt:   session.ExpiresAt,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}
}

func toAgentTaskInfo(task *models.AgentTask) *AgentTaskInfo {
	return &AgentTaskInfo{
		ID:          task.ID,
		SessionID:   task.SessionID,
		Type:        task.Type,
		Payload:     task.Payload,
		Status:      task.Status,
		Result:      task.Result,
		Error:       task.Error,
		StartedAt:   task.StartedAt,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
