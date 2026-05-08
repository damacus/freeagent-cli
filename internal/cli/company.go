package cli

import (
	"net/http"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v3"
)

func companyCommand() *cli.Command {
	return &cli.Command{
		Name:  "company",
		Usage: "View company information",
		Commands: []*cli.Command{
			{Name: "get", Usage: "Get company details", Action: action(companyGet)},
			{Name: "business-categories", Usage: "List business categories", Action: action(companyBusinessCategories)},
			{Name: "tax-timeline", Usage: "Get tax timeline", Action: action(companyTaxTimeline)},
		},
	}
}

func companyGet(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/company", nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func companyBusinessCategories(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/company/business_categories", nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func companyTaxTimeline(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/company/tax_timeline", nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}
