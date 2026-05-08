package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

var appVersion = "dev"

func NewApp(version string) *cli.Command {
	appVersion = version
	app := &cli.Command{
		Name:                  "freeagent",
		Usage:                 "CLI for the FreeAgent API",
		Version:               version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Sources: cli.EnvVars("FREEAGENT_CONFIG"),
				Usage:   "Path to config file",
			},
			&cli.StringFlag{
				Name:    "profile",
				Sources: cli.EnvVars("FREEAGENT_PROFILE"),
				Value:   "default",
				Usage:   "Credential profile name",
			},
			&cli.BoolFlag{
				Name:  "sandbox",
				Usage: "Use FreeAgent sandbox API",
			},
			&cli.StringFlag{
				Name:    "base-url",
				Sources: cli.EnvVars("FREEAGENT_BASE_URL"),
				Usage:   "Override API base URL",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output raw JSON",
			},
		},
		Before: initRuntime,
		Commands: []*cli.Command{
			accountingCommand(),
			accountManagersCommand(),
			authCommand(),
			bankAccountsCommand(),
			bankCommand(),
			billsCommand(),
			capitalAssetsCommand(),
			capitalAssetTypesCommand(),
			cashflowCommand(),
			categoriesCommand(),
			cisBandsCommand(),
			clientsCommand(),
			companyCommand(),
			contactsCommand(),
			creditNoteReconciliationsCommand(),
			creditNotesCommand(),
			emailAddressesCommand(),
			estimatesCommand(),
			expensesCommand(),
			invoiceCommand(),
			journalSetsCommand(),
			notesCommand(),
			payrollCommand(),
			payrollProfilesCommand(),
			priceListItemsCommand(),
			projectsCommand(),
			propertiesCommand(),
			rawCommand(),
			recurringInvoicesCommand(),
			salesTaxPeriodsCommand(),
			stockItemsCommand(),
			tasksCommand(),
			timeslipsCommand(),
			usersCommand(),
			versionCommand(),
			completionCommand(),
		},
	}

	cli.RootCommandHelpTemplate = strings.ReplaceAll(cli.RootCommandHelpTemplate, "GLOBAL OPTIONS", "GLOBAL FLAGS")
	return app
}

func action(fn func(*cli.Command) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Metadata == nil {
			c.Metadata = map[string]any{}
		}
		c.Metadata["context"] = ctx
		return fn(c)
	}
}

func commandContext(c *cli.Command) context.Context {
	if c.Metadata != nil {
		if ctx, ok := c.Metadata["context"].(context.Context); ok {
			return ctx
		}
	}
	return context.Background()
}

func initRuntime(ctx context.Context, c *cli.Command) (context.Context, error) {
	rt := Runtime{
		ConfigPath: c.String("config"),
		Profile:    c.String("profile"),
		Sandbox:    c.Bool("sandbox"),
		BaseURL:    c.String("base-url"),
		JSONOutput: c.Bool("json"),
	}

	if rt.Profile == "" {
		return nil, errors.New("profile cannot be empty")
	}

	if rt.BaseURL == "" {
		if rt.Sandbox {
			rt.BaseURL = "https://api.sandbox.freeagent.com/v2"
		} else {
			rt.BaseURL = "https://api.freeagent.com/v2"
		}
	}

	c.Root().Metadata = map[string]interface{}{
		"runtime": rt,
	}

	if !strings.HasSuffix(rt.BaseURL, "/v2") {
		return nil, fmt.Errorf("base-url must include /v2 (got %s)", rt.BaseURL)
	}

	return ctx, nil
}

func runtimeFrom(c *cli.Command) (Runtime, error) {
	if c.Root().Metadata == nil {
		return Runtime{}, errors.New("runtime not initialized")
	}
	if v, ok := c.Root().Metadata["runtime"]; ok {
		if rt, ok := v.(Runtime); ok {
			return rt, nil
		}
	}
	return Runtime{}, errors.New("runtime missing")
}
