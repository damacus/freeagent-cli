package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Missing_ReturnsEmptyConfig(t *testing.T) {
	cfg, path, err := Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil || cfg.Profiles == nil {
		t.Fatal("expected non-nil config with initialised Profiles map")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected empty profiles, got %v", cfg.Profiles)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(f, []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, _, err := Load(f); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSave_Load_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := &Config{
		Profiles: map[string]Profile{
			"default": {
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURI:  "http://localhost/cb",
				UserAgent:    "test/1.0",
				BaseURL:      "https://api.freeagent.com/v2",
			},
		},
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	p := loaded.Profile("default")
	want := original.Profiles["default"]
	if p.ClientID != want.ClientID {
		t.Errorf("ClientID: got %q, want %q", p.ClientID, want.ClientID)
	}
	if p.ClientSecret != want.ClientSecret {
		t.Errorf("ClientSecret: got %q, want %q", p.ClientSecret, want.ClientSecret)
	}
	if p.RedirectURI != want.RedirectURI {
		t.Errorf("RedirectURI: got %q, want %q", p.RedirectURI, want.RedirectURI)
	}
}

func TestSave_CreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "config.json")
	cfg := &Config{Profiles: map[string]Profile{}}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestConfig_Profile_Unknown_ReturnsEmpty(t *testing.T) {
	cfg := &Config{Profiles: map[string]Profile{}}
	p := cfg.Profile("nonexistent")
	if p.ClientID != "" || p.ClientSecret != "" {
		t.Errorf("expected empty profile, got %+v", p)
	}
}

func TestConfig_SetProfile_Then_Profile(t *testing.T) {
	cfg := &Config{}
	want := Profile{ClientID: "id", ClientSecret: "secret"}
	cfg.SetProfile("test", want)
	got := cfg.Profile("test")
	if got.ClientID != want.ClientID || got.ClientSecret != want.ClientSecret {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
