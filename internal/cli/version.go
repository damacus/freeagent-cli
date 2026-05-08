package cli

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the CLI version",
		Action: action(func(c *cli.Command) error {
			_, err := fmt.Fprintln(c.Root().Writer, c.Root().Version)
			return err
		}),
	}
}
