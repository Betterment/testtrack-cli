package migrationmanagers

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// MigrationManager manages the lifecycle of a migration
type MigrationManager struct {
	migration migrations.IMigration
	server    servers.IServer
}

// New returns a fully-loaded MigrationManager
func New(migration migrations.IMigration) (*MigrationManager, error) {
	server, err := servers.New()
	if err != nil {
		return nil, err
	}

	return &MigrationManager{
		migration: migration,
		server:    server,
	}, nil
}

// NewWithServer returns a MigrationManager using a provided Server
func NewWithServer(migration migrations.IMigration, server servers.IServer) *MigrationManager {
	return &MigrationManager{
		migration: migration,
		server:    server,
	}
}

// Save does the whole operation of validating, persisting, and sending a migration to the local TT server
func (m *MigrationManager) Save() error {
	err := m.migration.Validate()
	if err != nil {
		return err
	}

	err = m.persistFile()
	if err != nil {
		return err
	}

	valid, err := m.sync()
	if err != nil {
		return err
	}

	if !valid {
		m.deleteFile()
		return errors.New("Migration unsuccessful on server. Does your feature flag exist?")
	}

	return m.syncVersion()
}

// Run applies a migration to the TestTrack server
func (m *MigrationManager) Run() error {
	err := m.Apply()
	if err != nil {
		return err
	}
	return m.syncVersion()
}

// Apply applies a migration to the TestTrack server without recording the version
func (m *MigrationManager) Apply() error {
	err := m.migration.Validate()
	if err != nil {
		return err
	}

	valid, err := m.sync()
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("Migration unsuccessful on server. Does your feature flag exist?")
	}
	return nil
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

func (m *MigrationManager) sync() (bool, error) {
	resp, err := m.server.Post(m.migration.SyncPath(), m.migration.Serializable())
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case 204:
		return true, nil
	case 422:
		return false, nil
	default:
		return false, fmt.Errorf("got %d status code", resp.StatusCode)
	}
}

func (m *MigrationManager) syncVersion() error {
	resp, err := m.server.Post("api/v2/migrations", &serializers.MigrationVersion{Version: *m.migration.MigrationVersion()})
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("got %d status code", resp.StatusCode)
	}

	return nil
}
