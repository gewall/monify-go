// Package web embeds templates and static assets into the binary.
package web

import "embed"

//go:embed template static
var FS embed.FS
