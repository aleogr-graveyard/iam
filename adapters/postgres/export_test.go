package postgres

import (
	"io/fs"

	"github.com/jackc/pgx/v5"
)

// RawTx returns the database transaction under a Tx or SystemTx given to fn,
// so that RLS tests can run statements the repositories never issue.
func RawTx(tx any) pgx.Tx {
	switch v := tx.(type) {
	case *tenantTx:
		return v.tx
	case *systemTx:
		return v.tx
	}
	panic("postgres: RawTx: not a transaction of this adapter")
}

// MigrationVersions loads the migrations of files and returns their versions.
func MigrationVersions(files fs.FS) ([]string, error) {
	ms, err := loadMigrations(files)
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.version
	}
	return out, err
}

// EmbeddedMigrations are the migrations Migrate applies.
var EmbeddedMigrations fs.FS = migrationFiles

// Sanitize is the driver-error filter of every returned error.
var Sanitize = sanitize

// WriteError maps a failed write.
var WriteError = writeError
