package user

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore_CreatesEmptyCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	users := store.List()
	if len(users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(users))
	}

	if _, err := store.Active(); err == nil {
		t.Fatal("expected error for no active user")
	}
}

func TestUpsertUser_AndSetActive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u := User{
		ID:          "user-1",
		Email:       "alice@example.com",
		DisplayName: "Alice",
		CreatedAt:   time.Now(),
	}

	if err := store.UpsertUser(u); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	if err := store.SetActive("user-1"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	active, err := store.Active()
	if err != nil {
		t.Fatalf("Active: %v", err)
	}

	if active.ID != "user-1" {
		t.Fatalf("expected active user user-1, got %s", active.ID)
	}

	if active.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", active.Email)
	}
}

func TestUpsertUser_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u1 := User{ID: "user-1", Email: "old@example.com", DisplayName: "Old", CreatedAt: time.Now()}
	if err := store.UpsertUser(u1); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	u2 := User{ID: "user-1", Email: "new@example.com", DisplayName: "New", CreatedAt: time.Now()}
	if err := store.UpsertUser(u2); err != nil {
		t.Fatalf("UpsertUser update: %v", err)
	}

	users := store.List()
	if len(users) != 1 {
		t.Fatalf("expected 1 user after update, got %d", len(users))
	}

	if users[0].Email != "new@example.com" {
		t.Fatalf("expected updated email, got %s", users[0].Email)
	}
}

func TestSetActive_UnknownUser(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if err := store.SetActive("nonexistent"); err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func TestPersistence_AcrossRestart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store1, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore 1: %v", err)
	}

	u1 := User{ID: "user-1", Email: "alice@example.com", DisplayName: "Alice", CreatedAt: time.Now()}
	u2 := User{ID: "user-2", Email: "bob@example.com", DisplayName: "Bob", CreatedAt: time.Now()}

	if err := store1.UpsertUser(u1); err != nil {
		t.Fatalf("UpsertUser u1: %v", err)
	}
	if err := store1.UpsertUser(u2); err != nil {
		t.Fatalf("UpsertUser u2: %v", err)
	}
	if err := store1.SetActive("user-2"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	store2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore 2 (restart): %v", err)
	}

	active, err := store2.Active()
	if err != nil {
		t.Fatalf("Active after restart: %v", err)
	}

	if active.ID != "user-2" {
		t.Fatalf("expected active user-2 after restart, got %s", active.ID)
	}

	users := store2.List()
	if len(users) != 2 {
		t.Fatalf("expected 2 users after restart, got %d", len(users))
	}
}

func TestReload_DetectsChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store1, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore 1: %v", err)
	}

	u1 := User{ID: "user-1", Email: "alice@example.com", DisplayName: "Alice", CreatedAt: time.Now()}
	if err := store1.UpsertUser(u1); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}
	if err := store1.SetActive("user-1"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	store2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore 2: %v", err)
	}

	u2 := User{ID: "user-2", Email: "bob@example.com", DisplayName: "Bob", CreatedAt: time.Now()}
	if err := store2.UpsertUser(u2); err != nil {
		t.Fatalf("UpsertUser u2: %v", err)
	}
	if err := store2.SetActive("user-2"); err != nil {
		t.Fatalf("SetActive user-2: %v", err)
	}

	user, changed, err := store1.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}

	if !changed {
		t.Fatal("expected changed=true after external modification")
	}

	if user.ID != "user-2" {
		t.Fatalf("expected reloaded user user-2, got %s", user.ID)
	}
}

func TestReload_NoChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u := User{ID: "user-1", Email: "alice@example.com", DisplayName: "Alice", CreatedAt: time.Now()}
	if err := store.UpsertUser(u); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}
	if err := store.SetActive("user-1"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	_, changed, err := store.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}

	if changed {
		t.Fatal("expected changed=false when no external modification")
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u := User{ID: "user-1", Email: "test@example.com", DisplayName: "Test", CreatedAt: time.Now()}
	if err := store.UpsertUser(u); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("temp file should not exist after atomic write")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("users.json should exist after write")
	}
}
