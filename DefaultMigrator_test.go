package sqlx

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ljpx/logging"

	"github.com/ljpx/test"
	_ "github.com/mattn/go-sqlite3"
)

type DefaultMigratorFixture struct {
	db               *sql.DB
	databaseFileName string
	logger           *logging.DummyLogger
	migrator         *DefaultMigrator
}

func SetupDefaultMigratorFixture(t *testing.T) *DefaultMigratorFixture {
	rand.Seed(time.Now().UnixNano())

	databaseFileName := fmt.Sprintf("./sqlx-test-%v.db", rand.Int63())
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%v", databaseFileName))
	test.That(t, err).IsNil()

	logger := logging.NewDummyLogger()

	return &DefaultMigratorFixture{
		db:               db,
		databaseFileName: databaseFileName,
		logger:           logger,
		migrator:         NewDefaultMigrator(db, NewSQLite3Dictionary(), logger),
	}
}

func TearDownDefaultMigratorFixture(fixture *DefaultMigratorFixture) {
	fixture.db.Close()
	os.Remove(fixture.databaseFileName)
}

func TestDefaultMigratorInitializesWithNoMigrationsSuccessfully(t *testing.T) {
	// Arrange.
	fixture := SetupDefaultMigratorFixture(t)
	defer TearDownDefaultMigratorFixture(fixture)

	// Act.
	err := fixture.migrator.Migrate(0)

	// Assert.
	test.That(t, err).IsNil()

	validationSQL := `
		SELECT COUNT(*), * FROM migration_state
	`

	var count, id, timestamp int
	row := fixture.db.QueryRow(validationSQL)
	err = row.Scan(&count, &id, &timestamp)
	test.That(t, err).IsNil()

	test.That(t, count).IsEqualTo(1)
	test.That(t, id).IsEqualTo(0)
	test.That(t, timestamp).IsEqualTo(0)
}

func TestDefaultMigratorMigratesForwardsSuccessfully(t *testing.T) {
	// Arrange.
	fixture := SetupDefaultMigratorFixture(t)
	defer TearDownDefaultMigratorFixture(fixture)

	// Act.
	fixture.migrator.Use(&testMigration1{})
	fixture.migrator.Use(&testMigration2{})
	err := fixture.migrator.Migrate(2)

	// Assert.
	test.That(t, err).IsNil()

	validationSQL := `
		SELECT COUNT(*), * FROM user
	`

	var count, id int
	var name string
	row := fixture.db.QueryRow(validationSQL)
	err = row.Scan(&count, &id, &name)
	test.That(t, err).IsNil()

	test.That(t, count).IsEqualTo(1)
	test.That(t, id).IsEqualTo(42)
	test.That(t, name).IsEqualTo("John Smith")

	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 1' successfully.\n")
	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 2' successfully.\n")
}

func TestDefaultMigratorMigratesForwardsStepwiseSuccessfully(t *testing.T) {
	// Arrange.
	fixture := SetupDefaultMigratorFixture(t)
	defer TearDownDefaultMigratorFixture(fixture)

	// Act.
	fixture.migrator.Use(&testMigration1{})
	fixture.migrator.Use(&testMigration2{})

	err := fixture.migrator.Migrate(1)
	test.That(t, err).IsNil()

	err = fixture.migrator.Migrate(2)

	// Assert.
	test.That(t, err).IsNil()

	validationSQL := `
		SELECT COUNT(*), * FROM user
	`

	var count, id int
	var name string
	row := fixture.db.QueryRow(validationSQL)
	err = row.Scan(&count, &id, &name)
	test.That(t, err).IsNil()

	test.That(t, count).IsEqualTo(1)
	test.That(t, id).IsEqualTo(42)
	test.That(t, name).IsEqualTo("John Smith")

	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 1' successfully.\n")
	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 2' successfully.\n")
}

func TestDefaultMigratorMigratesBackwardsSuccessfully(t *testing.T) {
	// Arrange.
	fixture := SetupDefaultMigratorFixture(t)
	defer TearDownDefaultMigratorFixture(fixture)

	fixture.migrator.Use(&testMigration1{})
	fixture.migrator.Use(&testMigration2{})
	err := fixture.migrator.Migrate(2)
	test.That(t, err).IsNil()

	// Act.
	err = fixture.migrator.Migrate(1)

	// Assert.
	test.That(t, err).IsNil()

	validationSQL := `
		SELECT COUNT(*), * FROM user
	`

	var count, id *int
	var name *string
	row := fixture.db.QueryRow(validationSQL)
	err = row.Scan(&count, &id, &name)
	test.That(t, *count).IsEqualTo(0)
	test.That(t, id).IsNil()
	test.That(t, name).IsNil()

	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 1' successfully.\n")
	fixture.logger.AssertLogged(t, "Migrated up to 'Test Migration 2' successfully.\n")
	fixture.logger.AssertLogged(t, "Migrated down from 'Test Migration 2' successfully.\n")
}

func TestDefaultMigratorMigratesForwardsConcurrentlySuccessfully(t *testing.T) {
	// Arrange.
	fixture := SetupDefaultMigratorFixture(t)
	defer TearDownDefaultMigratorFixture(fixture)

	errc := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(3)

	closure := func() {
		migrator := NewDefaultMigrator(fixture.db, NewSQLite3Dictionary(), logging.NewDummyLogger())
		migrator.Use(&testMigration1{})
		migrator.Use(&testMigration2{})
		errc <- migrator.Migrate(2)
		wg.Done()
	}

	// Act.
	go closure()
	go closure()
	go closure()
	go func() {
		wg.Wait()
		close(errc)
	}()

	// Assert.
	for err := range errc {
		test.That(t, err).IsNil()
	}
}

// -----------------------------------------------------------------------------

type testMigration1 struct{}

var _ Migration = &testMigration1{}

func (m *testMigration1) Name() string {
	return "Test Migration 1"
}

func (m *testMigration1) Timestamp() uint64 {
	return 1
}

func (m *testMigration1) Up(dialect Dialect, tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE user (
			id INTEGER NOT NULL PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)

	return err
}

func (m *testMigration1) Down(dialect Dialect, tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE user;
	`)

	return err
}

type testMigration2 struct{}

var _ Migration = &testMigration2{}

func (m *testMigration2) Name() string {
	return "Test Migration 2"
}

func (m *testMigration2) Timestamp() uint64 {
	return 2
}

func (m *testMigration2) Up(dialect Dialect, tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO user (id, name) VALUES(42, "John Smith");
	`)

	return err
}

func (m *testMigration2) Down(dialect Dialect, tx *sql.Tx) error {
	_, err := tx.Exec(`
		DELETE FROM user WHERE id = 42;
	`)

	return err
}
