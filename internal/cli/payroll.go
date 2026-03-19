package cli

import (
	"fmt"
	"net/http"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func payrollCommand() *cli.Command {
	return &cli.Command{
		Name:  "payroll",
		Usage: "View payroll data",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get payroll for a tax year",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year (e.g. 2025)"},
				},
				Action: payrollGet,
			},
			{
				Name:  "get-period",
				Usage: "Get payroll for a specific period",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year"},
					&cli.IntFlag{Name: "period", Required: true, Usage: "Period number"},
				},
				Action: payrollGetPeriod,
			},
		},
	}
}

func payrollGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	year := c.Int("year")
	if year == 0 {
		return fmt.Errorf("--year required")
	}

	path := fmt.Sprintf("/payroll/%d", year)
	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func payrollGetPeriod(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	year := c.Int("year")
	period := c.Int("period")
	if year == 0 {
		return fmt.Errorf("--year required")
	}
	if period == 0 {
		return fmt.Errorf("--period required")
	}

	path := fmt.Sprintf("/payroll/%d/%d", year, period)
	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func payrollProfilesCommand() *cli.Command {
	return &cli.Command{
		Name:  "payroll-profiles",
		Usage: "View payroll profiles",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get payroll profiles for a tax year",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year"},
					&cli.StringFlag{Name: "user", Usage: "Filter by user URL"},
				},
				Action: payrollProfilesGet,
			},
		},
	}
}

func payrollProfilesGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	year := c.Int("year")
	if year == 0 {
		return fmt.Errorf("--year required")
	}

	path := fmt.Sprintf("/payroll_profiles/%d", year)
	if v := c.String("user"); v != "" {
		path += "?user=" + v
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
