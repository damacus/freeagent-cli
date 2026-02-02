package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"freegant-cli/internal/config"

	"github.com/urfave/cli/v3"
)

func rawCommand() *cli.Command {
	return &cli.Command{
		Name:  "raw",
		Usage: "Break-glass raw API call",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "method", Value: "GET"},
			&cli.StringFlag{Name: "path", Usage: "Path like /v2/invoices"},
			&cli.StringFlag{Name: "body", Usage: "JSON file to send"},
		},
		Action: rawAction,
	}
}

func rawAction(ctx context.Context, cmd *cli.Command) error {
	rt, err := runtimeFrom(cmd)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(ctx, rt, profile)
	if err != nil {
		return err
	}

	path := cmd.String("path")
	if path == "" {
		return fmt.Errorf("path is required")
	}

	method := cmd.String("method")
	if method == "" {
		method = http.MethodGet
	}

	var body []byte
	if bodyPath := cmd.String("body"); bodyPath != "" {
		data, err := os.ReadFile(bodyPath)
		if err != nil {
			return err
		}
		body = data
	}

	var resp []byte
	if body != nil {
		resp, _, _, err = client.Do(ctx, method, path, bytes.NewReader(body), "application/json")
	} else {
		resp, _, _, err = client.Do(ctx, method, path, nil, "")
	}
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	if json.Valid(resp) {
		var indented bytes.Buffer
		if err := json.Indent(&indented, resp, "", "  "); err == nil {
			_, _ = os.Stdout.Write(indented.Bytes())
			_, _ = os.Stdout.Write([]byte("\n"))
			return nil
		}
	}

	_, _ = os.Stdout.Write(resp)
	_, _ = os.Stdout.Write([]byte("\n"))
	return nil
}
