package service

import (
	"testing"
	"time"

	"officeworker/models"
)

type fakeSessionRepo struct {
	sessions          map[uint]*models.Session
	nextID            uint
	deleteWithRelHits int
}

func newFakeSessionRepo(items ...*models.Session) *fakeSessionRepo {
	repo := &fakeSessionRepo{
		sessions: make(map[uint]*models.Session),
		nextID:   1,
	}

	for _, item := range items {
		cp := *item
		repo.sessions[item.ID] = &cp
		if item.ID >= repo.nextID {
			repo.nextID = item.ID + 1
		}
	}

	return repo
}

func (r *fakeSessionRepo) Create(session *models.Session) error {
	cp := *session
	cp.ID = r.nextID
	cp.CreatedAt = time.Now().UTC()
	cp.UpdatedAt = cp.CreatedAt
	r.sessions[cp.ID] = &cp
	r.nextID++
	*session = cp
	return nil
}

func (r *fakeSessionRepo) ListByUserID(userID uint) ([]models.Session, error) {
	items := make([]models.Session, 0)
	for _, session := range r.sessions {
		if session.UserID == userID {
			items = append(items, *session)
		}
	}
	return items, nil
}

func (r *fakeSessionRepo) ListActiveByUserID(userID uint) ([]models.Session, error) {
	items := make([]models.Session, 0)
	for _, session := range r.sessions {
		if session.UserID == userID && session.Status == defaultSessionStatus {
			items = append(items, *session)
		}
	}
	return items, nil
}

func (r *fakeSessionRepo) FindByIDAndUserID(id, userID uint) (*models.Session, error) {
	session, ok := r.sessions[id]
	if !ok || session.UserID != userID {
		return nil, nil
	}
	cp := *session
	return &cp, nil
}

func (r *fakeSessionRepo) Delete(session *models.Session) error {
	delete(r.sessions, session.ID)
	return nil
}

func (r *fakeSessionRepo) DeleteWithRelations(session *models.Session) error {
	r.deleteWithRelHits++
	delete(r.sessions, session.ID)
	return nil
}

func (r *fakeSessionRepo) Update(session *models.Session) error {
	cp := *session
	cp.UpdatedAt = time.Now().UTC()
	r.sessions[cp.ID] = &cp
	*session = cp
	return nil
}

type fakeAgentTaskRepo struct {
	tasks  []*models.AgentTask
	nextID uint
}

func (r *fakeAgentTaskRepo) Create(task *models.AgentTask) error {
	r.nextID++
	cp := *task
	cp.ID = r.nextID
	cp.CreatedAt = time.Now().UTC()
	cp.UpdatedAt = cp.CreatedAt
	r.tasks = append(r.tasks, &cp)
	*task = cp
	return nil
}

func TestSessionServiceCreate(t *testing.T) {
	repo := newFakeSessionRepo()
	service := &SessionService{
		sessionRepo:   repo,
		agentTaskRepo: &fakeAgentTaskRepo{},
	}

	before := time.Now().UTC()
	resp, err := service.Create(7, &CreateSessionRequest{
		Title: "  First Session  ",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if resp.Title != "First Session" {
		t.Fatalf("expected trimmed title, got %q", resp.Title)
	}
	if resp.AgentID != defaultSessionAgentID {
		t.Fatalf("expected default agent id, got %q", resp.AgentID)
	}
	if resp.Status != defaultSessionStatus {
		t.Fatalf("expected status %q, got %q", defaultSessionStatus, resp.Status)
	}
	if !resp.ExpiresAt.After(before) {
		t.Fatalf("expected expires_at after now, got %v", resp.ExpiresAt)
	}
}

func TestSessionServiceListActive(t *testing.T) {
	repo := newFakeSessionRepo(
		&models.Session{BaseModel: models.BaseModel{ID: 1}, UserID: 7, Status: defaultSessionStatus, Title: "active"},
		&models.Session{BaseModel: models.BaseModel{ID: 2}, UserID: 7, Status: inactiveSessionStatus, Title: "inactive"},
		&models.Session{BaseModel: models.BaseModel{ID: 3}, UserID: 9, Status: defaultSessionStatus, Title: "other"},
	)
	service := &SessionService{
		sessionRepo:   repo,
		agentTaskRepo: &fakeAgentTaskRepo{},
	}

	items, err := service.ListActive(7)
	if err != nil {
		t.Fatalf("ListActive returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(items))
	}
	if items[0].ID != 1 {
		t.Fatalf("expected session 1, got %d", items[0].ID)
	}
}

func TestSessionServiceDeleteUsesDeleteWithRelations(t *testing.T) {
	repo := newFakeSessionRepo(
		&models.Session{BaseModel: models.BaseModel{ID: 1}, UserID: 7, Status: defaultSessionStatus},
	)
	service := &SessionService{
		sessionRepo:   repo,
		agentTaskRepo: &fakeAgentTaskRepo{},
	}

	if err := service.Delete(7, 1); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if repo.deleteWithRelHits != 1 {
		t.Fatalf("expected DeleteWithRelations to be called once, got %d", repo.deleteWithRelHits)
	}
}

func TestSessionServiceSendMessage(t *testing.T) {
	repo := newFakeSessionRepo(
		&models.Session{
			BaseModel: models.BaseModel{ID: 1},
			UserID:    7,
			Status:    defaultSessionStatus,
			Title:     "chat",
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		},
	)
	taskRepo := &fakeAgentTaskRepo{}
	service := &SessionService{
		sessionRepo:   repo,
		agentTaskRepo: taskRepo,
	}

	resp, err := service.SendMessage(7, 1, &SendMessageRequest{
		Message: "  hello  ",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if resp.Reply != "收到消息了" {
		t.Fatalf("expected canned reply, got %q", resp.Reply)
	}
	if resp.Task == nil || resp.Task.Type != "message" {
		t.Fatalf("expected message task, got %#v", resp.Task)
	}
	if resp.Task.Payload != "hello" {
		t.Fatalf("expected trimmed payload, got %q", resp.Task.Payload)
	}
	if len(taskRepo.tasks) != 1 {
		t.Fatalf("expected 1 task record, got %d", len(taskRepo.tasks))
	}

	stored := repo.sessions[1]
	if !stored.ExpiresAt.After(time.Now().UTC()) {
		t.Fatalf("expected session ttl to be extended, got %v", stored.ExpiresAt)
	}
}

func TestSessionServiceSendMessageRequiresActiveSession(t *testing.T) {
	repo := newFakeSessionRepo(
		&models.Session{BaseModel: models.BaseModel{ID: 1}, UserID: 7, Status: inactiveSessionStatus},
	)
	service := &SessionService{
		sessionRepo:   repo,
		agentTaskRepo: &fakeAgentTaskRepo{},
	}

	_, err := service.SendMessage(7, 1, &SendMessageRequest{Message: "hello"})
	if err == nil || err.Error() != "session is not active" {
		t.Fatalf("expected inactive session error, got %v", err)
	}
}
