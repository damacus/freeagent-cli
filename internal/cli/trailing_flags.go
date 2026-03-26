package cli

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

type positionalUpdateInput struct {
	id    string
	flags map[string]string
}

func parsePositionalUpdateInput(c *cli.Context, allowedFlags ...string) (positionalUpdateInput, error) {
	args := c.Args().Slice()
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return positionalUpdateInput{}, fmt.Errorf("resource id or url required")
	}

	input := positionalUpdateInput{
		id:    args[0],
		flags: make(map[string]string, len(allowedFlags)),
	}

	allowed := make(map[string]struct{}, len(allowedFlags))
	for _, name := range allowedFlags {
		allowed[name] = struct{}{}
	}

	remaining := args[1:]
	for i := 0; i < len(remaining); i++ {
		token := remaining[i]
		if !strings.HasPrefix(token, "--") {
			return positionalUpdateInput{}, fmt.Errorf("unexpected extra argument %q", token)
		}

		name, value, hasValue := strings.Cut(strings.TrimPrefix(token, "--"), "=")
		if name == "" {
			return positionalUpdateInput{}, fmt.Errorf("invalid trailing flag %q", token)
		}
		if _, ok := allowed[name]; !ok {
			return positionalUpdateInput{}, fmt.Errorf("unknown trailing flag %q", token)
		}

		if !hasValue {
			if i+1 >= len(remaining) {
				return positionalUpdateInput{}, fmt.Errorf("flag %q requires a value", "--"+name)
			}
			next := remaining[i+1]
			if strings.HasPrefix(next, "--") {
				return positionalUpdateInput{}, fmt.Errorf("flag %q requires a value", "--"+name)
			}
			value = next
			i++
		}

		input.flags[name] = value
	}

	return input, nil
}

func (i positionalUpdateInput) ID() string {
	return i.id
}

func (i positionalUpdateInput) String(c *cli.Context, name string) string {
	if value, ok := i.flags[name]; ok {
		return value
	}
	return c.String(name)
}
