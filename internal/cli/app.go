package cli

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func NewApp() *cli.Command {
	app := &cli.Command{
		Name:  "freegant",
		Usage: "CLI for the FreeAgent API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				EnvVars: []string{"FREEGANT_CONFIG"},
				Usage:   "Path to config file",
			},
			&cli.StringFlag{
				Name:    "profile",
				EnvVars: []string{"FREEGANT_PROFILE"},
				Value:   "default",
				Usage:   "Credential profile name",
			},
			&cli.BoolFlag{
				Name:  "sandbox",
				Usage: "Use FreeAgent sandbox API",
			},
			&cli.StringFlag{
				Name:    "base-url",
				EnvVars: []string{"FREEGANT_BASE_URL"},
				Usage:   "Override API base URL",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output raw JSON",
			},
		},
		Commands: []*cli.Command{
			authCommand(),
			invoiceCommand(),
			rawCommand(),
		},
	}

	cli.AppHelpTemplate = strings.ReplaceAll(cli.AppHelpTemplate, "GLOBAL OPTIONS", "GLOBAL FLAGS")
	return app
}

func runtimeFrom(cmd *cli.Command) (Runtime, error) {
	rt := Runtime{
		ConfigPath: cmd.String("config"),
		Profile:    cmd.String("profile"),
		Sandbox:    cmd.Bool("sandbox"),
		BaseURL:    cmd.String("base-url"),
		JSONOutput: cmd.Bool("json"),
	}

	if rt.Profile == "" {
		return rt, fmt.Errorf("profile cannot be empty")
	}

	if rt.BaseURL == "" {
		if rt.Sandbox {
			rt.BaseURL = "https://api.sandbox.freeagent.com/v2"
		} else {
			rt.BaseURL = "https://api.freeagent.com/v2"
		}
	}

	if !strings.HasSuffix(rt.BaseURL, "/v2") {
		return rt, fmt.Errorf("base-url must include /v2 (got %s)", rt.BaseURL)
	}

	return rt, nil
}
