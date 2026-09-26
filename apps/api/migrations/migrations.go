// Package migrations embeds the ordered database migrations for the Opus API.
package migrations

import "embed"

// FS holds the embedded SQL migration files.
//
//go:embed *.sql
var FS embed.FS
