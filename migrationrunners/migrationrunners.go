package migrationrunners

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/pkg/errors"
)

// Runner runs sets of migrations
type Runner struct {
	server servers.IServer
	schema *serializers.Schema
}

// New returns a Runner ready to use
func New() (*Runner, error) {
	server, err := servers.New()
	if err != nil {
		return nil, err
	}

	schema, err := schema.Load()
	if err != nil {
		return nil, err
	}

	return &Runner{server: server, schema: schema}, nil
}

// RunOutstanding runs all outstanding migrations
func (r *Runner) RunOutstanding() error {
	migrationsByVersion, err := migrationloaders.Load()
	if err != nil {
		return err
	}

	appliedMigrationVersions, err := r.getAppliedMigrationVersions()
	if err != nil {
		return err
	}

	for _, version := range appliedMigrationVersions {
		delete(migrationsByVersion, version.Version)
	}

	versions := migrationloaders.GetSortedVersions(migrationsByVersion)

	for _, version := range versions {
		mgr := migrationmanagers.NewWithDependencies(migrationsByVersion[version], r.server, r.schema)
		err := mgr.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Undo unapplies the latest migration if possible, removes it from local
// TestTrack server, and deletes the migration file
func (r *Runner) Undo() error {
	migration, err := r.unapplyLatest()
	if err != nil {
		return err
	}

	migrationVersion := *migration.MigrationVersion()
	filepaths, err := filepath.Glob(fmt.Sprintf("testtrack/migrate/%s_*.yml", migrationVersion))
	if err != nil {
		return err
	}
	if len(filepaths) != 1 {
		return fmt.Errorf("Couldn't find exactly one migration %s to delete", migrationVersion)
	}

	err = r.server.Delete(fmt.Sprintf("api/v2/migrations/%s", migrationVersion))
	if err != nil {
		return err
	}

	err = schema.Dump(r.schema)
	if err != nil {
		return err
	}

	return os.Remove(filepaths[0])
}

func (r *Runner) unapplyLatest() (migrations.IMigration, error) {
	migrationsByVersion, err := migrationloaders.Load()
	if err != nil {
		return nil, err
	}

	versions := migrationloaders.GetSortedVersions(migrationsByVersion)

	if len(versions) == 0 {
		return nil, errors.New("no migration to undo")
	}

	latestMigration := migrationsByVersion[versions[len(versions)-1]]

	var previousMigration migrations.IMigration
	for i := len(versions) - 2; i >= 0; i-- {
		m := migrationsByVersion[versions[i]]
		if m.SameResourceAs(latestMigration) {
			previousMigration = m
			break
		}
	}

	if previousMigration == nil {
		previousMigration, err = latestMigration.Inverse()
		if err != nil {
			return nil, errors.Wrap(err, "can't undo - you may want to `testtrack create` a new migration for this resource and then delete this migration file")
		}
	}
	r.schema.SchemaVersion = versions[len(versions)-2]

	mgr := migrationmanagers.NewWithDependencies(previousMigration, r.server, r.schema)
	err = mgr.Apply()
	if err != nil {
		return nil, err
	}
	return latestMigration, nil
}

func (r *Runner) getAppliedMigrationVersions() ([]serializers.MigrationVersion, error) {
	appliedMigrationVersions := make([]serializers.MigrationVersion, 0)

	err := r.server.Get("api/v2/migrations", &appliedMigrationVersions)
	if err != nil {
		return nil, err
	}

	return appliedMigrationVersions, nil
}
