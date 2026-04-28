package normalizer

import (
	"fmt"
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

type Service struct {
	now func() time.Time
}

func New() *Service {
	return &Service{now: time.Now}
}

func (s *Service) NormalizePhysics(page *acc_shm.PhysicsPage, identity Identity) (*event.PhysicsEvent, error) {
	if page == nil {
		return nil, fmt.Errorf("normalize physics: page is nil")
	}

	meta := s.newMetadata(identity, PhysicsSource, fmt.Sprintf("%d", page.PacketID))

	return &event.PhysicsEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
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

	meta := s.newMetadata(identity, GraphicsSource, fmt.Sprintf("%d", page.PacketID))

	return &event.GraphicsEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
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

	meta := s.newMetadata(identity, StaticSource, uuid.NewString())

	return &event.StaticEvent{
		EventID:       meta.EventID,
		EventTime:     meta.EventTime,
		IngestionTime: meta.IngestionTime,
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
	UsuarioID     string
	Username      string
	Source        string
}

func (s *Service) newMetadata(identity Identity, source string, eventID string) metadata {
	nowMillis := s.now().UnixMilli()

	return metadata{
		EventID:       eventID,
		EventTime:     nowMillis,
		IngestionTime: nowMillis,
		UsuarioID:     identity.UsuarioID,
		Username:      identity.Username,
		Source:        source,
	}
}
