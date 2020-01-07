package sqlx

import "database/sql"

// Tx redefines the methods implemented on *sql.Tx that are used by this
// package.
type Tx interface {
	Commit() error
	Rollback() error

	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

var _ Tx = &sql.Tx{}
