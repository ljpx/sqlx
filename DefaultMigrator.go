package sqlx

import (
	"database/sql"
	"sort"
	"time"
)

// DefaultMigrator is the default migrator in the package.  It is not thread
// safe.
type DefaultMigrator struct {
	db         Database
	dictionary Dictionary

	migrations map[uint64]Migration
}

var _ Migrator = &DefaultMigrator{}

// The delay between migration attempts, and the maximum number of attempts to
// make.
const (
	MigrationAttemptDelay    = time.Second
	MigrationMaximumAttempts = 5
)

// NewDefaultMigrator creates a new DefaultMigrator for the provided database
// using the provided Dictionary.
func NewDefaultMigrator(db Database, dictionary Dictionary) *DefaultMigrator {
	return &DefaultMigrator{
		db:         db,
		dictionary: dictionary,

		migrations: make(map[uint64]Migration),
	}
}

// Use adds a migration to the migrator.
func (m *DefaultMigrator) Use(migration Migration) {
	m.migrations[migration.Timestamp()] = migration
}

// Migrate migrates the database to the migration with the provided timestamp.
func (m *DefaultMigrator) Migrate(timestamp uint64) error {
	createTableSQL := m.dictionary.CreateMigrationStateTableIfDoesNotExist()
	ensureTimestampSQL := m.dictionary.EnsureCurrentTimestampIsPresentInTable()
	getCurrentTimestampSQL := m.dictionary.GetCurrentTimestamp()
	setCurrentTimestampSQL := m.dictionary.SetCurrentTimestamp()

	return retryUnderTransaction(m.db, MigrationAttemptDelay, MigrationMaximumAttempts, func(tx *sql.Tx) error {
		_, err := tx.Exec(createTableSQL)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ensureTimestampSQL)
		if err != nil {
			return err
		}

		var currentTimestamp uint64
		row := tx.QueryRow(getCurrentTimestampSQL)
		err = row.Scan(&currentTimestamp)
		if err != nil {
			return err
		}

		migrations, direction := findMigrationsInRange(m.migrations, currentTimestamp, timestamp)
		if len(migrations) == 0 {
			return nil
		}

		for _, migration := range migrations {
			if direction == 1 {
				err = migrationUpDownUp(tx, m.dictionary.Dialect(), migration)
			} else {
				err = migrationDownUpDown(tx, m.dictionary.Dialect(), migration)
			}

			if err != nil {
				return err
			}

			_, err = tx.Exec(setCurrentTimestampSQL, migration.Timestamp())
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func findMigrationsInRange(migrations map[uint64]Migration, current uint64, target uint64) ([]Migration, int) {
	if current == target {
		return nil, 0
	}

	direction := 1
	if current > target {
		direction = -1
	}

	result := []Migration{}

	for k, v := range migrations {
		if direction == 1 && (k <= current || k > target) {
			continue
		}

		if direction == -1 && (k > current || k <= target) {
			continue
		}

		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		less := result[i].Timestamp() < result[j].Timestamp()

		if direction == -1 {
			return !less
		}

		return less
	})

	return result, direction
}

func migrationUpDownUp(tx *sql.Tx, dialect Dialect, migration Migration) error {
	err := migration.Up(dialect, tx)
	if err != nil {
		return err
	}

	err = migration.Down(dialect, tx)
	if err != nil {
		return err
	}

	return migration.Up(dialect, tx)
}

func migrationDownUpDown(tx *sql.Tx, dialect Dialect, migration Migration) error {
	err := migration.Down(dialect, tx)
	if err != nil {
		return err
	}

	err = migration.Up(dialect, tx)
	if err != nil {
		return err
	}

	return migration.Down(dialect, tx)
}

func retryUnderTransaction(db Database, delay time.Duration, totalAllowedAttempts uint, closure func(tx *sql.Tx) error) error {
	c := uint(0)

	for c < totalAllowedAttempts {
		err := attemptUnderTransaction(db, closure)
		if err == nil {
			break
		}

		c++
		if c == totalAllowedAttempts {
			return err
		}

		time.Sleep(delay)
	}

	return nil
}

func attemptUnderTransaction(db Database, closure func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	didCommit := false
	defer func() {
		if !didCommit {
			tx.Rollback()
		}
	}()

	err = closure(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	didCommit = true
	return nil
}
