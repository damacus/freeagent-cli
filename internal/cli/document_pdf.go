package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func documentPDFCommand(resource string) *cli.Command {
	return &cli.Command{
		Name: "pdf", Usage: "Export a PDF, or return its API envelope with --json", ArgsUsage: "<id|url>",
		Flags: []cli.Flag{&cli.StringFlag{Name: "output", Usage: "New PDF file to write (required unless --json)"}},
		Action: action(func(c *cli.Command) error {
			endpoint, err := resourceEndpoint(resource, "/pdf")(c)
			if err != nil {
				return err
			}
			rt, err := runtimeFrom(c)
			if err != nil {
				return err
			}
			output := c.String("output")
			if output == "" && !rt.JSONOutput {
				return fmt.Errorf("--output is required, or use --json for the API response")
			}
			if output != "" && rt.JSONOutput {
				return fmt.Errorf("choose --output or --json, not both")
			}
			response, err := requestDocumentedEndpoint(c, http.MethodGet, endpoint, nil)
			if err != nil {
				return err
			}
			if rt.JSONOutput {
				return writeJSONOutput(response)
			}
			var result struct {
				PDF struct {
					Content string `json:"content"`
				} `json:"pdf"`
			}
			if err := json.Unmarshal(response, &result); err != nil {
				return fmt.Errorf("decode PDF response: %w", err)
			}
			data, err := base64.StdEncoding.DecodeString(result.PDF.Content)
			if err != nil {
				return fmt.Errorf("decode PDF content: %w", err)
			}
			if !strings.HasPrefix(string(data), "%PDF-") {
				return fmt.Errorf("API response does not contain PDF data")
			}
			file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return fmt.Errorf("create PDF file: %w", err)
			}
			_, writeErr := file.Write(data)
			closeErr := file.Close()
			if writeErr != nil {
				_ = os.Remove(output)
				return writeErr
			}
			if closeErr != nil {
				_ = os.Remove(output)
				return closeErr
			}
			fmt.Fprintf(os.Stdout, "Saved PDF to %s\n", cleanTerminalValue(output))
			return nil
		}),
	}
}
