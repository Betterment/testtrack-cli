package migrationmanagers

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// MigrationManager manages the lifecycle of a migration
type MigrationManager struct {
	migration migrations.IMigration
	server    servers.IServer
	schema    *serializers.Schema
}

// New returns a MigrationManager without server connectivity for filesystem-only side effects
func New(migration migrations.IMigration) (*MigrationManager, error) {
	schema, err := schema.Read()
	if err != nil {
		return nil, err
	}

	return &MigrationManager{
		migration: migration,
		server:    nil,
		schema:    schema,
	}, nil
}

// NewWithDependencies returns a MigrationManager using a provided Server
func NewWithDependencies(migration migrations.IMigration, server servers.IServer, schema *serializers.Schema) *MigrationManager {
	return &MigrationManager{
		migration: migration,
		server:    server,
		schema:    schema,
	}
}

// Save does the whole operation of validating and persisting a
// migration to disk, and updating the schema
func (m *MigrationManager) Save() error {
	err := m.migration.Validate()
	if err != nil {
		return err
	}

	err = m.persistFile()
	if err != nil {
		return err
	}

	migrationRepo, err := migrationloaders.Load()
	if err != nil {
		return err
	}
	err = m.ApplyToSchema(migrationRepo)
	if err != nil {
		return err
	}

	return schema.Write(m.schema)
}

// Run applies a migration to the TestTrack server, writing out the schema
func (m *MigrationManager) Run(migrationRepo migrations.Repository) error {
	err := m.Apply(migrationRepo)
	if err != nil {
		return err
	}
	err = m.SyncVersion()
	if err != nil {
		return err
	}
	return schema.Write(m.schema)
}

// ApplyToSchema validates and applies a migration to the in-memory schema representation
func (m *MigrationManager) ApplyToSchema(migrationRepo migrations.Repository) error {
	err := m.migration.Validate()
	if err != nil {
		return err
	}

	err = m.migration.ApplyToSchema(m.schema, migrationRepo)
	if err != nil {
		return err
	}

	appliedVersion := m.migration.MigrationVersion()
	if appliedVersion != nil && m.schema.SchemaVersion < *appliedVersion {
		m.schema.SchemaVersion = *appliedVersion
	}
	return nil
}

// Apply applies a migration to the TestTrack server and in-memory schema
// without recording the version to TestTrack server
func (m *MigrationManager) Apply(migrationRepo migrations.Repository) error {
	err := m.migration.Validate()
	if err != nil {
		return err
	}

	err = m.migration.ApplyToSchema(m.schema, migrationRepo)
	if err != nil {
		return err
	}
	resp, err := m.server.Post(m.migration.SyncPath(), m.migration.Serializable())
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 204:
		return nil
	case 422:
		return errors.New("Migration unsuccessful on server. Does your split exist?")
	default:
		return fmt.Errorf("got %d status code", resp.StatusCode)
	}
}

func (m *MigrationManager) persistFile() error {
	stat, err := os.Stat("testtrack/migrate")
	if err != nil {
		return errors.Wrap(err, "migration directory not found - run `testtrack init_project` to resolve")
	}

	if !stat.IsDir() {
		return errors.New("testtrack/migrate is not a directory")
	}

	out, err := yaml.Marshal(m.migration.File())

	err = ioutil.WriteFile(fmt.Sprintf("testtrack/migrate/%s", *m.migration.Filename()), out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (m *MigrationManager) deleteFile() error {
	return os.Remove(fmt.Sprintf("testtrack/migrate/%s", *m.migration.Filename()))
}

// SyncVersion marks schema versions as applied on TestTrack server
func (m *MigrationManager) SyncVersion() error {
	resp, err := m.server.Post("api/v2/migrations", &serializers.MigrationVersion{Version: *m.migration.MigrationVersion()})
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("got %d status code", resp.StatusCode)
	}

	appliedVersion := m.migration.MigrationVersion()
	if m.schema.SchemaVersion < *appliedVersion {
		m.schema.SchemaVersion = *appliedVersion
	}

	return nil
}
