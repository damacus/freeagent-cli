package main

import (
	"runtime/debug"
	"testing"
)

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

func TestBuildVersionFrom_ModuleVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "v0.5.1",
		},
	}

	if got := buildVersionFrom("dev", "", info); got != "v0.5.1" {
		t.Fatalf("buildVersionFrom() = %q, want %q", got, "v0.5.1")
	}
}

func TestBuildVersionFrom_DevelModuleVersionFallsBackToVCSRevision(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "(devel)",
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "36d324c7b8533f637683d1afaae52e87a54bc122"},
		},
	}

	if got := buildVersionFrom("dev", "", info); got != "dev (commit 36d324c7b853)" {
		t.Fatalf("buildVersionFrom() = %q, want %q", got, "dev (commit 36d324c7b853)")
	}
}

func TestBuildVersionFrom_DirtyVCSRevision(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{
			Version: "(devel)",
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "36d324c7b8533f637683d1afaae52e87a54bc122"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	if got := buildVersionFrom("dev", "", info); got != "dev (commit 36d324c7b853, dirty)" {
		t.Fatalf("buildVersionFrom() = %q, want %q", got, "dev (commit 36d324c7b853, dirty)")
	}
}
