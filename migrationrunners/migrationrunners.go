package migrationrunners

import (
	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
)

// Runner runs sets of migrations
type Runner struct {
	server servers.IServer
	schema *serializers.Schema
}

// New returns a Runner configured with the provided server
func New(server servers.IServer) (*Runner, error) {
	schema, err := schema.Read()
	if err != nil {
		return nil, err
	}

	return &Runner{server: server, schema: schema}, nil
}

// RunOutstanding runs all outstanding migrations
func (r *Runner) RunOutstanding() error {
	migrationRepo, err := r.getOutstandingMigrations()
	if err != nil {
		return err
	}

	versions := migrationRepo.SortedVersions()

	for _, version := range versions {
		mgr := migrationmanagers.NewWithDependencies(migrationRepo[version], r.server, r.schema)
		err := mgr.Migrate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) getOutstandingMigrations() (migrations.Repository, error) {
	migrationRepo, err := migrationloaders.Load()
	if err != nil {
		return nil, err
	}

	appliedMigrationVersions, err := r.getAppliedMigrationVersions()
	if err != nil {
		return nil, err
	}

	for _, version := range appliedMigrationVersions {
		delete(migrationRepo, version.Version)
	}
	return migrationRepo, nil
}

func (r *Runner) getAppliedMigrationVersions() ([]serializers.MigrationVersion, error) {
	appliedMigrationVersions := make([]serializers.MigrationVersion, 0)

	err := r.server.Get("api/v2/migrations", &appliedMigrationVersions)
	if err != nil {
		return nil, err
	}

	return appliedMigrationVersions, nil
}
