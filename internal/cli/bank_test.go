package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitIDs(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"123", []string{"123"}},
		{"123,456", []string{"123", "456"}},
		{"123 456", []string{"123", "456"}},
		{"123\n456\n789", []string{"123", "456", "789"}},
		{"  123 , 456 , ", []string{"123", "456"}},
		{"@123", []string{"123"}}, // strips leading @
	}

	for _, tc := range cases {
		got := splitIDs(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("splitIDs(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitIDs(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestDedupeStrings(t *testing.T) {
	cases := []struct {
		input []string
		want  []string
	}{
		{nil, nil},
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "a", "b"}, []string{"a", "b"}},
		{[]string{"", "a", "", "b"}, []string{"a", "b"}},
		{[]string{"z", "z", "z"}, []string{"z"}},
	}

	for _, tc := range cases {
		got := dedupeStrings(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("dedupeStrings(%v) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("dedupeStrings(%v)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestLooksLikeFile_ExistingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "ids.txt")
	if err := os.WriteFile(f, []byte("123"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !looksLikeFile(f) {
		t.Errorf("looksLikeFile(%q) = false, want true for existing file", f)
	}
}

func TestLooksLikeFile_NonExistent(t *testing.T) {
	if looksLikeFile("/nonexistent/path/ids.txt") {
		t.Error("looksLikeFile returned true for non-existent path")
	}
}

func TestLooksLikeFile_Multiline(t *testing.T) {
	if looksLikeFile("123\n456") {
		t.Error("looksLikeFile returned true for multiline string")
	}
}

func TestParseIDList_InlineIDs(t *testing.T) {
	ids, err := parseIDList("123,456,789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[0] != "123" || ids[2] != "789" {
		t.Errorf("got %v, want [123 456 789]", ids)
	}
}

func TestParseIDList_FilePrefix(t *testing.T) {
	f := filepath.Join(t.TempDir(), "ids.txt")
	if err := os.WriteFile(f, []byte("aaa\nbbb\nccc"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	ids, err := parseIDList("@" + f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[1] != "bbb" {
		t.Errorf("got %v, want [aaa bbb ccc]", ids)
	}
}

func TestParseIDList_Empty(t *testing.T) {
	ids, err := parseIDList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids != nil {
		t.Errorf("expected nil for empty input, got %v", ids)
	}
}
