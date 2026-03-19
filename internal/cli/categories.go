package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func categoriesCommand() *cli.Command {
	return &cli.Command{
		Name:  "categories",
		Usage: "Manage categories",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List categories", Action: categoriesList},
			{Name: "get", Usage: "Get a category by ID or URL", ArgsUsage: "<id|url>", Action: categoriesGet},
			{
				Name: "create", Usage: "Create a category",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "description", Required: true, Usage: "Description"},
					&cli.StringFlag{Name: "nominal-code", Usage: "Nominal code"},
					&cli.StringFlag{Name: "category-group", Usage: "Category group URL"},
					&cli.StringFlag{Name: "tax-reporting-name", Usage: "Tax reporting name"},
				},
				Action: categoriesCreate,
			},
			{
				Name: "update", Usage: "Update a category", ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "description", Usage: "Description"},
					&cli.StringFlag{Name: "tax-reporting-name", Usage: "Tax reporting name"},
				},
				Action: categoriesUpdate,
			},
			{Name: "delete", Usage: "Delete a category", ArgsUsage: "<id|url>", Action: categoriesDelete},
		},
	}
}

func categoriesList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/categories", nil, "")
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CategoriesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if len(result.Categories) == 0 {
		fmt.Fprintln(os.Stdout, "No categories found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Code\tDescription\tURL")
	for _, cat := range result.Categories {
		fmt.Fprintf(w, "%v\t%v\t%v\n", cat.NominalCode, cat.Description, cat.URL)
	}
	_ = w.Flush()
	return nil
}

func categoriesGet(c *cli.Context) error {
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

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("category id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func categoriesCreate(c *cli.Context) error {
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

	input := fa.CategoryInput{
		Description: c.String("description"),
	}
	if v := c.String("nominal-code"); v != "" {
		input.NominalCode = v
	}
	if v := c.String("category-group"); v != "" {
		input.CategoryGroup = v
	}
	if v := c.String("tax-reporting-name"); v != "" {
		input.TaxReportingName = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/categories", fa.CreateCategoryRequest{Category: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CategoryResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Created category %v (%v)\n", result.Category.Description, result.Category.URL)
	return nil
}

func categoriesUpdate(c *cli.Context) error {
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

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("category id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil {
		return err
	}

	input := fa.CategoryInput{}
	if v := c.String("description"); v != "" {
		input.Description = v
	}
	if v := c.String("tax-reporting-name"); v != "" {
		input.TaxReportingName = v
	}

	if input.Description == "" && input.TaxReportingName == "" {
		return fmt.Errorf("at least one of --description or --tax-reporting-name is required")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateCategoryRequest{Category: input})
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func categoriesDelete(c *cli.Context) error {
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

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("category id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Category deleted")
	return nil
}
