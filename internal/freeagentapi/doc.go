// Package freeagentapi contains typed models for the FreeAgent REST API.
// Response shapes are derived from live API responses; request shapes mirror
// the API's accepted JSON payloads.
//
// To regenerate reference types from the OpenAPI spec (query-param structs
// only — the spec has no response schemas):
//
//go:generate oapi-codegen --config ../../oapi-codegen.yaml ../../spec.yaml
package freeagentapi
