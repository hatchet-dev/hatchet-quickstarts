// Package quickstarts exposes Hatchet quickstart templates.
package quickstarts

import (
	"embed"
	"io/fs"
)

//go:embed all:templates/*
var content embed.FS

// TemplatesFS returns the embedded quickstart templates.
func TemplatesFS() fs.FS {
	return content
}
