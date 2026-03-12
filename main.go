package main

import (
	"log"
	"os"

	"github.com/damacus/freeagent-cli/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	app := cli.NewApp(version)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
