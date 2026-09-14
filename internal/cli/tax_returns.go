package cli

import (
	"net/http"
	"strings"

	"github.com/urfave/cli/v3"
)

func taxReturnsCommand(name, resource string, perUser, payments, datedPayments bool) *cli.Command {
	root := &cli.Command{
		Name:        name,
		Usage:       "View returns and manage FreeAgent status markers",
		Description: "Filing markers record status in FreeAgent; they do not submit a return to HMRC or Companies House. Payment markers do not transfer money.",
	}
	operations := []string{"list", "get", "mark-filed", "mark-unfiled"}
	if payments {
		operations = append(operations, "mark-paid", "mark-unpaid")
	}
	for _, operation := range operations {
		method := http.MethodGet
		usage := strings.ReplaceAll(operation, "-", " ") + " return"
		if strings.HasPrefix(operation, "mark-") {
			method = http.MethodPut
			usage += " status in FreeAgent only"
		}
		cmd := endpointCommand(operation, usage, method, func(c *cli.Command) (string, error) {
			base := "/" + resource
			if perUser {
				id, err := documentedResourceID(c, "users", c.String("user"))
				if err != nil {
					return "", err
				}
				base = "/users/" + id + base
			}
			if operation == "list" {
				return paginatedEndpoint(c, base)
			}
			date, err := datedArgument(c)
			if err != nil {
				return "", err
			}
			base += "/" + date
			if operation == "get" {
				return base, nil
			}
			if datedPayments && (operation == "mark-paid" || operation == "mark-unpaid") {
				paymentDate := c.String("payment-date")
				if err := validateEndpointDate("payment date", paymentDate); err != nil {
					return "", err
				}
				base += "/payments/" + paymentDate
			}
			return base + "/mark_as_" + strings.TrimPrefix(operation, "mark-"), nil
		})
		if operation != "list" {
			cmd.ArgsUsage = "<period-end:YYYY-MM-DD>"
		} else {
			cmd.Flags = append(cmd.Flags, paginationFlags()...)
		}
		if perUser {
			cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "user", Required: true, Usage: "User ID or API URL"})
		}
		if datedPayments && (operation == "mark-paid" || operation == "mark-unpaid") {
			cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "payment-date", Required: true, Usage: "Payment due date (YYYY-MM-DD)"})
		}
		root.Commands = append(root.Commands, cmd)
	}
	return root
}
