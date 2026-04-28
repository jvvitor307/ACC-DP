package normalizer

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"acc-dp/producer/internal/domain/event"
	"acc-dp/producer/internal/source/acc_shm"
)

const (
	PhysicsSource  = "acpmf_physics"
	GraphicsSource = "acpmf_graphics"
	StaticSource   = "acpmf_static"
)

type Identity struct {
	UsuarioID string
	Username  string
}

type sessionTracker struct {
	mu        sync.Mutex
	sessionID string
	lastPacketID int32
	userID    string
}

func newSessionTracker(userID string) *sessionTracker {
	return &sessionTracker{
		sessionID: generateSessionID(userID),
		userID:    userID,
	}
}

func (t *sessionTracker) update(packetID int32) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if packetID < t.lastPacketID {
		t.sessionID = generateSessionID(t.userID)
	}
	t.lastPacketID = packetID
	return t.sessionID
}

func (t *sessionTracker) current() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sessionID
}

func generateSessionID(userID string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(userID+uuid.New().String())).String()
}

type Service struct {
	now            func() time.Time
	sessionTracker *sessionTracker
}

func New(userID string) *Service {
	return &Service{
		now:            time.Now,
		sessionTracker: newSessionTracker(userID),
	}
}

func (s *Service) NormalizePhysics(page *acc_shm.PhysicsPage, identity Identity) (*event.PhysicsEvent, error) {
	if page == nil {
		return nil, fmt.Errorf("normalize physics: page is nil")
	}

	sessionID := s.sessionTracker.update(page.PacketID)
	meta := s.newMetadata(identity, PhysicsSource, fmt.Sprintf("%d", page.PacketID), sessionID)

	return &event.PhysicsEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
		SessionID:     meta.SessionID,
		UsuarioID:     meta.UsuarioID,
		Username:      meta.Username,
		Source:        meta.Source,
		Payload:       *page,
	}, nil
}

func (s *Service) NormalizeGraphics(page *acc_shm.GraphicsPage, identity Identity) (*event.GraphicsEvent, error) {
	if page == nil {
		return nil, fmt.Errorf("normalize graphics: page is nil")
	}

	sessionID := s.sessionTracker.current()
	meta := s.newMetadata(identity, GraphicsSource, fmt.Sprintf("%d", page.PacketID), sessionID)

	return &event.GraphicsEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
		SessionID:     meta.SessionID,
		UsuarioID:     meta.UsuarioID,
		Username:      meta.Username,
		Source:        meta.Source,
		Payload:       *page,
	}, nil
}

func (s *Service) NormalizeStatic(page *acc_shm.StaticPage, identity Identity) (*event.StaticEvent, error) {
	if page == nil {
		return nil, fmt.Errorf("normalize static: page is nil")
	}

	sessionID := s.sessionTracker.current()
	meta := s.newMetadata(identity, StaticSource, uuid.NewString(), sessionID)

	return &event.StaticEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
		SessionID:     meta.SessionID,
		UsuarioID:     meta.UsuarioID,
		Username:      meta.Username,
		Source:        meta.Source,
		Payload:       *page,
	}, nil
}

type metadata struct {
	EventID       string
	EventTime     int64
	IngestionTime int64
	SessionID     string
	UsuarioID     string
	Username      string
	Source        string
}

func (s *Service) newMetadata(identity Identity, source string, eventID string, sessionID string) metadata {
	nowMillis := s.now().UnixMilli()

	return metadata{
		EventID:       eventID,
		EventTime:     nowMillis,
		IngestionTime: nowMillis,
		SessionID:     sessionID,
		UsuarioID:     identity.UsuarioID,
		Username:      identity.Username,
		Source:        source,
	}
}
