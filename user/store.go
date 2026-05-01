package user

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu       sync.Mutex
	filePath string
	cache    *Cache
}

func NewStore(storagePath string) (*Store, error) {
	s := &Store{
		filePath: storagePath,
		cache:    &Cache{},
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("load user store: %w", err)
	}

	return s, nil
}

func (s *Store) UpsertUser(u User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.cache.Users {
		if existing.ID == u.ID {
			s.cache.Users[i] = u
			return s.save()
		}
	}

	s.cache.Users = append(s.cache.Users, u)
	return s.save()
}

func (s *Store) SetActive(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, u := range s.cache.Users {
		if u.ID == userID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user %s not found in cache", userID)
	}

	s.cache.ActiveUserID = userID
	return s.save()
}

func (s *Store) Active() (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache.ActiveUserID == "" {
		return User{}, fmt.Errorf("no active user")
	}

	for _, u := range s.cache.Users {
		if u.ID == s.cache.ActiveUserID {
			return u, nil
		}
	}

	return User{}, fmt.Errorf("active user %s not found in cache", s.cache.ActiveUserID)
}

func (s *Store) List() []User {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]User, len(s.cache.Users))
	copy(result, s.cache.Users)
	return result
}

func (s *Store) Reload() (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldActive := s.cache.ActiveUserID

	if err := s.load(); err != nil {
		return User{}, false, fmt.Errorf("reload: %w", err)
	}

	if s.cache.ActiveUserID == "" {
		return User{}, false, fmt.Errorf("no active user")
	}

	for _, u := range s.cache.Users {
		if u.ID == s.cache.ActiveUserID {
			changed := oldActive != s.cache.ActiveUserID
			return u, changed, nil
		}
	}

	return User{}, false, fmt.Errorf("active user %s not found in cache", s.cache.ActiveUserID)
}

func (s *Store) FilePath() string {
	return s.filePath
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.cache = &Cache{}
			return nil
		}
		return fmt.Errorf("read %s: %w", s.filePath, err)
	}

	if len(data) == 0 {
		s.cache = &Cache{}
		return nil
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("parse %s: %w", s.filePath, err)
	}

	s.cache = &cache
	return nil
}

func (s *Store) save() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(s.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", tmpFile, err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpFile, s.filePath, err)
	}

	return nil
}
