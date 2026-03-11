package migrations

import "embed"

// FS holds migration SQL files in this directory. Use with postgres.RunMigrations:
//
//	postgres.RunMigrations(db, migrations.FS)
//
// Add or replace *.sql files here (e.g. 000001_initial.up.sql, 000001_initial.down.sql).
//
//go:embed *.sql
var FS embed.FS
