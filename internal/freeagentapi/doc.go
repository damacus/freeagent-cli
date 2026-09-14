// Package freeagentapi contains hand-maintained runtime models and generated
// reference types for the FreeAgent API. Raw CLI JSON is never filtered by these
// models. Reference types carry the ignore build tag to avoid conflicting with
// compatibility models used by existing commands.
//
//go:generate sh ../../scripts/generate-api.sh
package freeagentapi
