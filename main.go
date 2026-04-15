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
	if version != "dev" {
		return version
	}

	revision := commit
	dirty := false

	if revision == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if revision == "" {
						revision = setting.Value
					}
				case "vcs.modified":
					dirty = setting.Value == "true"
				}
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

	return "dev (commit " + strings.TrimSpace(revision) + ")"
}

func main() {
	app := cli.NewApp(buildVersion())
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
