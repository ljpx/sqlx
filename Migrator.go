package sqlx

import "github.com/ljpx/logging"

// Migrator defines the methods that any type capable of migrating a database
// must implement.  Use adds a migration to the list of migrations managed by
// the migrator.  Migrate migrates the database to the provided timestamp.
// Migrate is safe for use across multiple concurrent instances of an
// application.
type Migrator interface {
	Use(migration Migration)
	Migrate(timestamp uint64) error
}

// NewMigrator returns a new migrator for the provided database using the
// provided dictionary.  It is an alias of NewDefaultMigrator.
func NewMigrator(db Database, dictionary Dictionary, logger logging.Logger) Migrator {
	return NewDefaultMigrator(db, dictionary, logger)
}
