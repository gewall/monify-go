// Package migrations embeds the goose SQL migration files so the binary can run
// them itself via `monify migrate`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
