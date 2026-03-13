package storage

import (
	"errors"
	"testing"
	"time"
)

// mockTokenStore is a simple in-memory TokenStore for unit tests.
type mockTokenStore struct {
	token *Token
	err   error
}

func (m *mockTokenStore) Get(_ string) (*Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.token == nil {
		return nil, ErrNotFound
	}
	return m.token, nil
}

func (m *mockTokenStore) Set(_ string, t *Token) error {
	if m.err != nil {
		return m.err
	}
	m.token = t
	return nil
}

func (m *mockTokenStore) Delete(_ string) error {
	if m.err != nil {
		return m.err
	}
	m.token = nil
	return nil
}

var errStoreDown = errors.New("store unavailable")

func TestStore_Get_Primary(t *testing.T) {
	tok := &Token{AccessToken: "primary"}
	s := NewStore(&mockTokenStore{token: tok}, nil)

	got, err := s.Get("p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccessToken != "primary" {
		t.Errorf("got %q, want %q", got.AccessToken, "primary")
	}
}

func TestStore_Get_FallsBackToSecondary(t *testing.T) {
	tok := &Token{AccessToken: "fallback"}
	s := NewStore(
		&mockTokenStore{err: ErrNotFound},
		&mockTokenStore{token: tok},
	)

	got, err := s.Get("p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccessToken != "fallback" {
		t.Errorf("got %q, want %q", got.AccessToken, "fallback")
	}
}

func TestStore_Get_BothMissing(t *testing.T) {
	s := NewStore(
		&mockTokenStore{err: ErrNotFound},
		&mockTokenStore{err: ErrNotFound},
	)
	_, err := s.Get("p")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_Set_Primary(t *testing.T) {
	primary := &mockTokenStore{}
	s := NewStore(primary, nil)

	tok := &Token{AccessToken: "tok"}
	if err := s.Set("p", tok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primary.token == nil || primary.token.AccessToken != "tok" {
		t.Error("token not written to primary")
	}
}

func TestStore_Set_PrimaryFail_FallbackSucceeds(t *testing.T) {
	fallback := &mockTokenStore{}
	s := NewStore(
		&mockTokenStore{err: errStoreDown},
		fallback,
	)

	tok := &Token{AccessToken: "tok"}
	if err := s.Set("p", tok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fallback.token == nil || fallback.token.AccessToken != "tok" {
		t.Error("token not written to fallback")
	}
}

// TestStore_Set_BothFail ensures our fix is in place: Store.Set must return
// an error when both primary and fallback stores fail (previously returned nil).
func TestStore_Set_BothFail_ReturnsError(t *testing.T) {
	s := NewStore(
		&mockTokenStore{err: errStoreDown},
		&mockTokenStore{err: errStoreDown},
	)
	if err := s.Set("p", &Token{AccessToken: "tok"}); err == nil {
		t.Error("expected error when both stores fail, got nil")
	}
}

func TestStore_Set_NilStores_ReturnsError(t *testing.T) {
	s := NewStore(nil, nil)
	if err := s.Set("p", &Token{AccessToken: "tok"}); err == nil {
		t.Error("expected error with no stores configured, got nil")
	}
}

func TestStore_Delete(t *testing.T) {
	tok := &Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)}
	primary := &mockTokenStore{token: tok}
	fallback := &mockTokenStore{token: tok}
	s := NewStore(primary, fallback)

	if err := s.Delete("p"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primary.token != nil {
		t.Error("primary token not deleted")
	}
	if fallback.token != nil {
		t.Error("fallback token not deleted")
	}
}
