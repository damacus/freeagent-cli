package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStore_SetGet(t *testing.T) {
	dir := t.TempDir()
	fs := &FileStore{Dir: dir}

	tok := &Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour).Truncate(time.Second),
	}
	if err := fs.Set("profile", tok); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := fs.Get("profile")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, tok.AccessToken)
	}
	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
	if !got.ExpiresAt.Equal(tok.ExpiresAt) {
		t.Errorf("ExpiresAt: got %v, want %v", got.ExpiresAt, tok.ExpiresAt)
	}
}

func TestFileStore_Get_NotFound(t *testing.T) {
	fs := &FileStore{Dir: t.TempDir()}
	_, err := fs.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	fs := &FileStore{Dir: dir}

	tok := &Token{AccessToken: "tok"}
	if err := fs.Set("profile", tok); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := fs.Delete("profile"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := fs.Get("profile")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("after Delete, expected ErrNotFound, got %v", err)
	}
}

func TestFileStore_Delete_NotFound(t *testing.T) {
	fs := &FileStore{Dir: t.TempDir()}
	if err := fs.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFileStore_Set_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	fs := &FileStore{Dir: dir + "/nested/dir"}

	if err := fs.Set("p", &Token{AccessToken: "tok"}); err != nil {
		t.Fatalf("Set with nested dir: %v", err)
	}
	if _, err := fs.Get("p"); err != nil {
		t.Fatalf("Get after Set in nested dir: %v", err)
	}
}

func TestFileStore_Get_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	s := &FileStore{Dir: dir}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.Get("bad")
	if err == nil {
		t.Error("expected error for corrupt JSON")
	}
}
