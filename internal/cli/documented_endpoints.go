package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// endpointCommand keeps dedicated endpoint wrappers on the shared authenticated client.
// Callers provide a fixed documented route; user input never selects an arbitrary route.
func endpointCommand(name, usage, method string, route func(*cli.Command) (string, error)) *cli.Command {
	cmd := &cli.Command{Name: name, Usage: usage}
	if method != http.MethodGet {
		cmd.Flags = []cli.Flag{&cli.BoolFlag{Name: "dry-run", Usage: "Show the request without changing FreeAgent"}}
	}
	cmd.Action = action(func(c *cli.Command) error {
		endpoint, err := route(c)
		if err != nil {
			return err
		}
		return runDocumentedEndpoint(c, method, endpoint, nil)
	})
	return cmd
}

func runDocumentedEndpoint(c *cli.Command, method, endpoint string, body any) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	if method != http.MethodGet && c.Bool("dry-run") {
		data, err := json.Marshal(map[string]any{"method": method, "path": endpoint, "body": body})
		if err != nil {
			return err
		}
		return renderEndpointResponse(data, rt.JSONOutput)
	}
	response, err := requestDocumentedEndpoint(c, method, endpoint, body)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(response))) == 0 {
		response = []byte(`{"success":true}`)
		// Statement upload acknowledgement does not confirm that the import completed.
		if method == http.MethodPost && strings.HasPrefix(endpoint, "/bank_transactions/statement?") {
			response = []byte(`{"uploaded":true,"import_verified":false}`)
		}
	}
	return renderEndpointResponse(response, rt.JSONOutput)
}

func requestDocumentedEndpoint(c *cli.Command, method, endpoint string, body any) ([]byte, error) {
	rt, err := runtimeFrom(c)
	if err != nil {
		return nil, err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return nil, err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return nil, err
	}
	var response []byte
	if body == nil {
		response, _, _, err = client.Do(commandContext(c), method, endpoint, nil, "")
	} else {
		response, _, _, err = client.DoJSON(commandContext(c), method, endpoint, body)
	}
	return response, err
}

func renderEndpointResponse(data []byte, raw bool) error {
	if raw {
		return writeJSONOutput(data)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode API response: %w", err)
	}
	return printEndpointValue(os.Stdout, "", value, "")
}

// Recursive labels preserve report breakdowns and decimal strings without dumping JSON
// into the human output. Sorting object keys makes output stable between runs.
func printEndpointValue(w io.Writer, label string, value any, indent string) error {
	write := func(format string, args ...any) error { _, err := fmt.Fprintf(w, format, args...); return err }
	switch v := value.(type) {
	case map[string]any:
		if label != "" {
			if err := write("%s%s:\n", indent, label); err != nil {
				return err
			}
			indent += "  "
		}
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := printEndpointValue(w, humanFieldLabel(key), v[key], indent); err != nil {
				return err
			}
		}
	case []any:
		if len(v) == 0 {
			return write("%s%s: none\n", indent, label)
		}
		for i, item := range v {
			if err := printEndpointValue(w, fmt.Sprintf("%s %d", label, i+1), item, indent); err != nil {
				return err
			}
		}
	case nil:
		return write("%s%s: —\n", indent, label)
	default:
		return write("%s%s: %s\n", indent, label, cleanTerminalValue(fmt.Sprint(v)))
	}
	return nil
}

func cleanTerminalValue(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}

func humanFieldLabel(field string) string {
	field = cleanTerminalValue(strings.ReplaceAll(field, "_", " "))
	if field == "url" {
		return "URL"
	}
	if field == "" {
		return field
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

func fixedEndpoint(endpoint string) func(*cli.Command) (string, error) {
	return func(c *cli.Command) (string, error) {
		if c.Args().Len() != 0 {
			return "", fmt.Errorf("%s takes no arguments", c.Name)
		}
		return endpoint, nil
	}
}

func datedArgument(c *cli.Command) (string, error) {
	if c.Args().Len() != 1 {
		return "", fmt.Errorf("exactly one period end date (YYYY-MM-DD) is required")
	}
	date := c.Args().First()
	return date, validateEndpointDate("period end", date)
}

func validateEndpointDate(name, value string) error {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("%s must be a valid date in YYYY-MM-DD format (got %q)", name, value)
	}
	return nil
}

func resourceArgument(c *cli.Command, resource string) (string, error) {
	if c.Args().Len() != 1 {
		return "", fmt.Errorf("exactly one %s ID or URL is required", resource)
	}
	return documentedResourceID(c, resource, c.Args().First())
}

func documentedResourceID(c *cli.Command, resource, value string) (string, error) {
	if strings.Contains(value, "/") {
		rt, err := runtimeFrom(c)
		if err != nil {
			return "", err
		}
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("invalid %s URL", resource)
		}
		cfg, _, err := loadConfig(rt)
		if err != nil {
			return "", err
		}
		profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
		base, err := url.Parse(profile.BaseURL)
		if err != nil {
			return "", err
		}
		if (parsed.IsAbs() || parsed.Host != "") && (parsed.Scheme != base.Scheme || parsed.Host != base.Host) {
			return "", fmt.Errorf("%s URL must use the configured API origin", resource)
		}
		prefix := "/v2/" + resource + "/"
		if !strings.HasPrefix(parsed.Path, prefix) || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return "", fmt.Errorf("expected a %s resource URL", resource)
		}
		value = strings.TrimPrefix(parsed.Path, prefix)
	}
	if value == "" || strings.Trim(value, "0123456789") != "" {
		return "", fmt.Errorf("%s ID must contain only digits", resource)
	}
	return value, nil
}

func resourceEndpoint(resource, suffix string) func(*cli.Command) (string, error) {
	return func(c *cli.Command) (string, error) {
		id, err := resourceArgument(c, resource)
		if err != nil {
			return "", err
		}
		return "/" + resource + "/" + id + suffix, nil
	}
}

func documentedReadResource(name, resource, usage string) *cli.Command {
	get := endpointCommand("get", "Get "+usage, http.MethodGet, resourceEndpoint(resource, ""))
	get.ArgsUsage = "<id|url>"
	list := endpointCommand("list", "List "+usage, http.MethodGet, func(c *cli.Command) (string, error) {
		return paginatedEndpoint(c, "/"+resource)
	})
	list.Flags = paginationFlags()
	return &cli.Command{Name: name, Usage: "View " + usage, Commands: []*cli.Command{
		list, get,
	}}
}

func paginationFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{Name: "page", Value: 1, Usage: "Page number (one page per request)"},
		&cli.IntFlag{Name: "per-page", Value: 25, Usage: "Records per page (1-100)"},
	}
}

func paginatedEndpoint(c *cli.Command, endpoint string) (string, error) {
	if c.Args().Len() != 0 {
		return "", fmt.Errorf("%s takes no arguments", c.Name)
	}
	if c.Int("page") < 1 {
		return "", fmt.Errorf("page must be at least 1")
	}
	if c.Int("per-page") < 1 || c.Int("per-page") > 100 {
		return "", fmt.Errorf("per-page must be between 1 and 100")
	}
	query := url.Values{}
	if c.IsSet("page") {
		query.Set("page", strconv.Itoa(c.Int("page")))
	}
	if c.IsSet("per-page") {
		query.Set("per_page", strconv.Itoa(c.Int("per-page")))
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	return endpoint, nil
}
