package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func completionCommand() *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Generate shell completion scripts",
		Commands: []*cli.Command{
			{
				Name:  "fish",
				Usage: "Generate fish shell completions",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "install",
						Usage: "Install completions to ~/.config/fish/completions/",
					},
				},
				Action: action(func(c *cli.Command) error {
					script, err := c.Root().ToFishCompletion()
					if err != nil {
						return err
					}
					if c.Bool("install") {
						dir := os.ExpandEnv("$HOME/.config/fish/completions")
						if err := os.MkdirAll(dir, 0755); err != nil {
							return err
						}
						path := dir + "/freeagent-cli.fish"
						if err := os.WriteFile(path, []byte(script), 0644); err != nil {
							return err
						}
						fmt.Fprintf(os.Stdout, "Installed fish completions to %s\n", path)
						return nil
					}
					fmt.Print(script)
					return nil
				}),
			},
		},
	}
}
