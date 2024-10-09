//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc -f ../../sqlc.yaml generate
package sqlc

import _ "embed"

//go:embed schema.sql
var Schema string
