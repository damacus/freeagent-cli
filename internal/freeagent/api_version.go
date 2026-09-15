package freeagent

import "context"

type apiVersionKey struct{}

const AttachmentsAPIVersion = "2026-09-01"

// WithAPIVersion applies a version to this request and its pagination/retries.
func WithAPIVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersionKey{}, version)
}

func APIVersion(ctx context.Context) string {
	version, _ := ctx.Value(apiVersionKey{}).(string)
	return version
}
