package sqlx

// SQLite3Dictionary is an implementation of Dictionary for use with SQLite3.
type SQLite3Dictionary struct{}

var _ Dictionary = &SQLite3Dictionary{}

// NewSQLite3Dictionary returns a new SQLite3Dictionary.
func NewSQLite3Dictionary() *SQLite3Dictionary {
	return &SQLite3Dictionary{}
}

// Dialect returns SQLite3.
func (d *SQLite3Dictionary) Dialect() Dialect {
	return SQLite3Dialect
}

// CreateMigrationStateTableIfDoesNotExist returns the SQLite3 version of this
// query.
func (d *SQLite3Dictionary) CreateMigrationStateTableIfDoesNotExist() string {
	return `
		CREATE TABLE IF NOT EXISTS migration_state (
			id INTEGER NOT NULL DEFAULT 0 PRIMARY KEY,
			ts INTEGER NOT NULL DEFAULT 0
		);
	`
}

// EnsureCurrentTimestampIsPresentInTable returns the SQLite3 version of this
// query.
func (d *SQLite3Dictionary) EnsureCurrentTimestampIsPresentInTable() string {
	return `
		INSERT INTO migration_state
		SELECT id, ts
		FROM (SELECT 0 as id, 0 as ts) t
		WHERE NOT EXISTS (SELECT * FROM migration_state m WHERE m.id = t.id);
	`
}

// GetCurrentTimestamp returns the SQLite3 version of this query.
func (d *SQLite3Dictionary) GetCurrentTimestamp() string {
	return `
		SELECT ts FROM migration_state WHERE id = 0;
	`
}

// SetCurrentTimestamp returns the SQLite3 version of this query.
func (d *SQLite3Dictionary) SetCurrentTimestamp() string {
	return `
		UPDATE migration_state SET ts = ? WHERE id = 0;
	`
}
