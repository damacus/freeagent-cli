package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v3"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func withStatementFile(cmd *cli.Command) *cli.Command {
	for _, flag := range cmd.Flags {
		if f, ok := flag.(*cli.StringFlag); ok && f.Name == "body" {
			f.Required = false
		}
	}
	cmd.Usage = "Upload a JSON, OFX, QIF or supported CSV statement; verify with bank list"
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "file", Usage: "Statement file (OFX, QIF, QBO or supported CSV)"})
	original := cmd.Action
	cmd.Action = func(ctx context.Context, c *cli.Command) error {
		if (c.String("file") == "") == (c.String("body") == "") {
			return fmt.Errorf("provide exactly one of --file or --body")
		}
		if c.String("file") == "" {
			return original(ctx, c)
		}
		if c.Args().Len() != 0 {
			return fmt.Errorf("import-statement takes no arguments")
		}
		account, err := documentedResourceURL(c, "bank_accounts", c.String("bank-account"))
		if err != nil {
			return err
		}
		path := "/bank_transactions/statement?" + url.Values{"bank_account": {account}}.Encode()
		file, err := os.Open(c.String("file"))
		if err != nil {
			return err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("statement must be a non-empty regular file")
		}
		filename := filepath.Base(file.Name())
		contentType := "application/octet-stream"
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".ofx", ".qbo":
			contentType = "application/x-ofx"
		case ".qif":
			contentType = "application/x-qif"
		case ".csv":
			contentType = "text/csv"
		default:
			return fmt.Errorf("statement file must be OFX, QBO, QIF or CSV")
		}
		rt, err := runtimeFrom(c)
		if err != nil {
			return err
		}
		if c.Bool("dry-run") {
			data, _ := json.Marshal(map[string]any{"method": "POST", "path": path, "file_name": filename, "content_type": contentType, "size": info.Size()})
			return renderEndpointResponse(data, rt.JSONOutput)
		}
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "statement", "filename": filename}))
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(part, file); err != nil {
			return err
		}
		if err = writer.Close(); err != nil {
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
		if _, _, _, err = client.Do(ctx, http.MethodPost, path, &body, writer.FormDataContentType()); err != nil {
			return err
		}
		return renderEndpointResponse([]byte(`{"uploaded":true,"import_verified":false}`), rt.JSONOutput)
	}
	return cmd
}
