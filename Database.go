package sqlx

import "database/sql"

// Database redefines the methods implemented on *sql.DB that are used by this
// package.
type Database interface {
	Begin() (*sql.Tx, error)
}

var _ Database = &sql.DB{}
