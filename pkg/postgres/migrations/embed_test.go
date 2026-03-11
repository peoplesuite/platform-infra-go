package migrations

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFS_ContainsMigrationFiles(t *testing.T) {
	entries, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "embedded FS should contain at least one .sql file")

	hasUp := false
	hasDown := false
	for _, name := range entries {
		if len(name) >= 7 && name[len(name)-7:] == ".up.sql" {
			hasUp = true
		}
		if len(name) >= 9 && name[len(name)-9:] == ".down.sql" {
			hasDown = true
		}
	}
	require.True(t, hasUp, "embedded FS should contain at least one .up.sql file")
	require.True(t, hasDown, "embedded FS should contain at least one .down.sql file")
}
