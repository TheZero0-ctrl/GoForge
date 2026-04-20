package dbmigrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var ErrNoChange = errors.New("no migration change")

type ErrDirty struct {
	Version int
}

func (e ErrDirty) Error() string {
	return fmt.Sprintf("dirty database version %d", e.Version)
}

type Runner interface {
	Up(sourceURL, databaseURL string) error
	DownSteps(sourceURL, databaseURL string, steps int) error
	Force(sourceURL, databaseURL string, version int) error
}

type OSRunner struct{}

func NewRunner() *OSRunner {
	return &OSRunner{}
}

func (r *OSRunner) Up(sourceURL, databaseURL string) error {
	return Up(sourceURL, databaseURL)
}

func (r *OSRunner) DownSteps(sourceURL, databaseURL string, steps int) error {
	return DownSteps(sourceURL, databaseURL, steps)
}

func (r *OSRunner) Force(sourceURL, databaseURL string, version int) error {
	return Force(sourceURL, databaseURL, version)
}

func Up(sourceURL, databaseURL string) error {
	m, err := newMigrator(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	upErr := m.Up()
	return closeAndMap(m, upErr, "apply migrations")
}

func DownSteps(sourceURL, databaseURL string, steps int) error {
	m, err := newMigrator(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	downErr := m.Steps(-steps)
	return closeAndMap(m, downErr, "rollback migrations")
}

func Force(sourceURL, databaseURL string, version int) error {
	m, err := newMigrator(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	forceErr := m.Force(version)
	return closeAndMap(m, forceErr, "force migration version")
}

func newMigrator(sourceURL, databaseURL string) (*migrate.Migrate, error) {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("initialize migrator: %w", err)
	}
	return m, nil
}

func closeAndMap(m *migrate.Migrate, opErr error, opName string) error {
	sourceErr, databaseErr := m.Close()

	if opErr != nil && !errors.Is(opErr, migrate.ErrNoChange) {
		var dirty migrate.ErrDirty
		if errors.As(opErr, &dirty) {
			return ErrDirty{Version: dirty.Version}
		}
		return fmt.Errorf("%s: %w", opName, opErr)
	}

	if sourceErr != nil {
		return fmt.Errorf("close migration source: %w", sourceErr)
	}

	if databaseErr != nil {
		return fmt.Errorf("close migration database: %w", databaseErr)
	}

	if errors.Is(opErr, migrate.ErrNoChange) {
		return ErrNoChange
	}

	return nil
}
