package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the CLI version",
		Action: func(c *cli.Context) error {
			_, err := fmt.Fprintln(c.App.Writer, c.App.Version)
			return err
		},
	}
}
