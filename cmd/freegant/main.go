package main

import (
	"context"
	"log"
	"os"

	"freegant-cli/internal/cli"
)

func main() {
	app := cli.NewApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
