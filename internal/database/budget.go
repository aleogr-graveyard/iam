package database

// The connection budget on the shared lab Postgres (ADR 0001, item 6). The
// instance gives the iam a ceiling of six connections (CONNECTION LIMIT 6);
// TestConnectionBudget fails when the sum below exceeds it, and
// TestConnectionBudget_MatchesPoolConfigs ties the constants to the pools the
// binary really opens. Scheduled work calls the service and uses its pool, so
// it costs nothing here. No code may open a connection outside these pools:
// no LISTEN/NOTIFY, no dedicated connection; an exhausted pool waits.
const (
	// ServicePool is the pool of one service instance (database.Open).
	ServicePool int32 = 2
	// MaxInstances is the most service instances running at once.
	MaxInstances int32 = 1
	// RolloutRevisions is the revisions alive together during a rollout.
	RolloutRevisions int32 = 2
	// MigrationJob is the pool of the migration job (database.OpenMigrator).
	MigrationJob int32 = 1
	// Person is one operator in psql.
	Person int32 = 1
	// Ceiling is the connection limit of the iam database.
	Ceiling int32 = 6
)
