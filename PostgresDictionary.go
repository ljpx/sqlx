package sqlx

// PostgresDictionary is an implementation of Dictionary for use with Postgres.
type PostgresDictionary struct{}

var _ Dictionary = &PostgresDictionary{}

// NewPostgresDictionary returns a new PostgresDictionary.
func NewPostgresDictionary() *PostgresDictionary {
	return &PostgresDictionary{}
}

// Dialect returns Postgres.
func (d *PostgresDictionary) Dialect() Dialect {
	return PostgresDialect
}

// CreateMigrationStateTableIfDoesNotExist returns the Postgres version of this
// query.
func (d *PostgresDictionary) CreateMigrationStateTableIfDoesNotExist() string {
	return `
		CREATE TABLE IF NOT EXISTS migration_state (
			id INTEGER NOT NULL DEFAULT 0 PRIMARY KEY,
			ts INTEGER NOT NULL DEFAULT 0
		);
	`
}

// EnsureCurrentTimestampIsPresentInTable returns the Postgres version of this
// query.
func (d *PostgresDictionary) EnsureCurrentTimestampIsPresentInTable() string {
	return `
		INSERT INTO migration_state
		SELECT id, ts
		FROM (SELECT 0 as id, 0 as ts) t
		WHERE NOT EXISTS (SELECT * FROM migration_state m WHERE m.id = t.id);
	`
}

// GetCurrentTimestamp returns the Postgres version of this query.
func (d *PostgresDictionary) GetCurrentTimestamp() string {
	return `
		SELECT ts FROM migration_state WHERE id = 0;
	`
}

// SetCurrentTimestamp returns the Postgres version of this query.
func (d *PostgresDictionary) SetCurrentTimestamp() string {
	return `
		UPDATE migration_state SET ts = $1 WHERE id = 0;
	`
}
