package main

import (
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/damacus/freeagent-cli/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"
var commit = ""
var date = ""

func buildVersion() string {
	info, _ := debug.ReadBuildInfo()
	return buildVersionFrom(version, commit, info)
}

func buildVersionFrom(buildVersion, buildCommit string, info *debug.BuildInfo) string {
	if buildVersion != "dev" {
		return buildVersion
	}

	if info != nil {
		moduleVersion := strings.TrimSpace(info.Main.Version)
		if moduleVersion != "" && moduleVersion != "(devel)" {
			return moduleVersion
		}
	}

	revision := strings.TrimSpace(buildCommit)
	dirty := false

	if revision == "" && info != nil {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if revision == "" {
					revision = strings.TrimSpace(setting.Value)
				}
			case "vcs.modified":
				dirty = setting.Value == "true"
			}
		}
	}

	if revision == "" {
		return version
	}

	if len(revision) > 12 {
		revision = revision[:12]
	}

	if dirty {
		return "dev (commit " + revision + ", dirty)"
	}

	return "dev (commit " + revision + ")"
}

func main() {
	app := cli.NewApp(buildVersion())
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
