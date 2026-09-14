#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 --config oapi-codegen.yaml spec.yaml
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
printf '//go:build ignore\n\n' > "$tmp"
cat internal/freeagentapi/types.gen.go >> "$tmp"
cat "$tmp" > internal/freeagentapi/types.gen.go
# The type-only generator is not an operation inventory; publish one separately.
awk '
BEGIN { print "# API specification operations\n"; print "Generated from spec.yaml by scripts/generate-api.sh. This lists declared operations, not live verification.\n"; print "| Method | Path |"; print "| --- | --- |" }
/^  \// { path=$0; sub(/^  /,"",path); sub(/:$/,"",path) }
/^    (get|post|put|delete|patch):/ { method=$1; sub(/:$/,"",method); print "| " toupper(method) " | `" path "` |" }
' spec.yaml > docs/api-spec-operations.md
