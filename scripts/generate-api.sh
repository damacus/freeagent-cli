#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 --config oapi-codegen.yaml spec.yaml
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
printf '//go:build ignore\n\n' > "$tmp"
cat internal/freeagentapi/types.gen.go >> "$tmp"
cat "$tmp" > internal/freeagentapi/types.gen.go
