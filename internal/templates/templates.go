package templates

import "embed"

//go:embed all:base all:api all:cli all:crud all:services all:worker all:snippets all:views all:command all:crud_db
var FS embed.FS
