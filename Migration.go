package sqlx

import "database/sql"

// Migration defines the methods that any database migration must implement.
// Name is a human-friendly name for the migration.  Timestamp is used to
// determine what order migrations should be run in.  Enabled determines if the
// migration is enabled or not.  Up and Down apply and tear down the migration
// respectively.
type Migration interface {
	Name() string
	Timestamp() uint64

	Up(dialect Dialect, tx *sql.Tx) error
	Down(dialect Dialect, tx *sql.Tx) error
}
