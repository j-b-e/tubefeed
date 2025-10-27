package templates

import "embed"

// FS contains all embedded HTML templates
//
//go:embed *.html
var FS embed.FS
