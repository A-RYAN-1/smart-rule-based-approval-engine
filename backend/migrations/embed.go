// Package migrationfiles embeds the SQL migrations into the binary, so they
// run the same way locally, in tests and on any host.
package migrationfiles

import "embed"

// FS holds every *.sql file in this directory. Files are named
// <unix-timestamp>_<name>.up.sql / .down.sql and applied in order.
//
//go:embed *.sql
var FS embed.FS
