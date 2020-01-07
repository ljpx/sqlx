package sqlx

// Dialect is a simple string alias type that represents different SQL dialects
// e.g. "Postgres" or "SQLite3".
type Dialect string

// The sqlx package provides support for two dialects by default: "Postgres"
// and "SQLite3".
const (
	PostgresDialect = "Postgres"
	SQLite3Dialect  = "SQLite3"
)
