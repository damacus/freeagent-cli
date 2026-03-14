package cli

import "testing"

func TestExtractCode_BareCode(t *testing.T) {
	got, err := extractCode("myauthcode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myauthcode" {
		t.Errorf("got %q, want %q", got, "myauthcode")
	}
}

func TestExtractCode_FromRedirectURL(t *testing.T) {
	input := "http://127.0.0.1:8797/callback?code=abc123&state=xyz"
	got, err := extractCode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestExtractCode_Empty(t *testing.T) {
	if _, err := extractCode(""); err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestRandomState_ReturnsNonEmptyString(t *testing.T) {
	s, err := randomState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) == 0 {
		t.Error("expected non-empty state string")
	}
}

func TestRandomState_Unique(t *testing.T) {
	a, _ := randomState()
	b, _ := randomState()
	if a == b {
		t.Error("expected different state values on successive calls")
	}
}
