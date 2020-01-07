package sqlx

// Dictionary defines the methods that a SQL Dialect Dictionary must implement.
type Dictionary interface {
	Dialect() Dialect

	CreateMigrationStateTableIfDoesNotExist() string
	EnsureCurrentTimestampIsPresentInTable() string

	GetCurrentTimestamp() string
	SetCurrentTimestamp() string
}
