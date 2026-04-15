package main

import "testing"

func TestBuildVersion_ReleaseBuild(t *testing.T) {
	origVersion, origCommit := version, commit
	t.Cleanup(func() {
		version = origVersion
		commit = origCommit
	})

	version = "0.4.0"
	commit = "36d324c7b8533f637683d1afaae52e87a54bc122"

	if got := buildVersion(); got != "0.4.0" {
		t.Fatalf("buildVersion() = %q, want %q", got, "0.4.0")
	}
}

func TestBuildVersion_DevBuildIncludesCommit(t *testing.T) {
	origVersion, origCommit := version, commit
	t.Cleanup(func() {
		version = origVersion
		commit = origCommit
	})

	version = "dev"
	commit = "36d324c7b8533f637683d1afaae52e87a54bc122"

	if got := buildVersion(); got != "dev (commit 36d324c7b853)" {
		t.Fatalf("buildVersion() = %q, want %q", got, "dev (commit 36d324c7b853)")
	}
}
