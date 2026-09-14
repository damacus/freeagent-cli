package cli

import (
	"fmt"
	"github.com/urfave/cli/v3"
	"net/http"
	"net/url"
	"time"
)

func bankExplanationListCommand() *cli.Command {
	cmd := endpointCommand("list", "List one page of bank explanations", http.MethodGet, func(c *cli.Command) (string, error) {
		endpoint, err := paginatedEndpoint(c, "/bank_transaction_explanations")
		if err != nil {
			return "", err
		}
		u, _ := url.Parse(endpoint)
		q := u.Query()
		account, err := documentedResourceURL(c, "bank_accounts", c.String("bank-account"))
		if err != nil {
			return "", err
		}
		q.Set("bank_account", account)
		if err := addDateQueries(c, q); err != nil {
			return "", err
		}
		u.RawQuery = q.Encode()
		return u.String(), nil
	})
	cmd.Flags = append(paginationFlags(), &cli.StringFlag{Name: "bank-account", Required: true, Usage: "Bank account ID or URL"}, &cli.StringFlag{Name: "from", Usage: "First date (YYYY-MM-DD)"}, &cli.StringFlag{Name: "to", Usage: "Last date (YYYY-MM-DD)"}, &cli.StringFlag{Name: "updated-since", Usage: "Updated since date or RFC3339 timestamp"})
	return cmd
}
func addDateQueries(c *cli.Command, q url.Values) error {
	for flag, key := range map[string]string{"from": "from_date", "to": "to_date", "updated-since": "updated_since"} {
		value := c.String(flag)
		if value == "" {
			continue
		}
		err := validateEndpointDate(flag, value)
		if err != nil && flag == "updated-since" {
			_, err = time.Parse(time.RFC3339Nano, value)
		}
		if err != nil {
			return fmt.Errorf("%s must be a valid date%s", flag, map[bool]string{true: " or RFC3339 timestamp"}[flag == "updated-since"])
		}
		q.Set(key, value)
	}
	if c.String("from") != "" && c.String("to") != "" && c.String("from") > c.String("to") {
		return fmt.Errorf("from must not be after to")
	}
	return nil
}
