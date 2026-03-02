//go:build tools

// Package tools tracks dependencies used only by code-generation tools
// (e.g. atlas-provider-gorm for the Atlas schema loader).
// The "tools" build tag is never set during normal compilation, but
// go mod tidy respects it and keeps these imports in go.mod / go.sum.
package tools

import _ "ariga.io/atlas-provider-gorm/gormschema"
