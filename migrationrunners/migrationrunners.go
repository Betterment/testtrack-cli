package migrationrunners

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/identifiertypes"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/Betterment/testtrack-cli/splitdecisions"
	"github.com/Betterment/testtrack-cli/splitretirements"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Runner runs sets of migrations
type Runner struct {
	server servers.IServer
}

// New returns a Runner ready to use
func New() (*Runner, error) {
	server, err := servers.New()
	if err != nil {
		return nil, err
	}

	return &Runner{server: server}, nil
}

// RunOutstanding runs all outstanding migrations
func (r *Runner) RunOutstanding() error {
	migrationsByVersion, err := loadMigrations()
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

	versions := getSortedVersions(migrationsByVersion)

	for _, version := range versions {
		mgr := migrationmanagers.NewWithServer(migrationsByVersion[version], r.server)
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

	return os.Remove(filepaths[0])
}

func (r *Runner) unapplyLatest() (migrations.IMigration, error) {
	migrationsByVersion, err := loadMigrations()
	if err != nil {
		return nil, err
	}

	versions := getSortedVersions(migrationsByVersion)

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

	mgr := migrationmanagers.NewWithServer(previousMigration, r.server)
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

func loadMigrations() (map[string]migrations.IMigration, error) {
	files, err := ioutil.ReadDir("testtrack/migrate")
	if err != nil {
		return nil, err
	}

	migrationsByVersion := make(map[string]migrations.IMigration)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue // Skip hidden files
		}

		migrationVersion, err := migrations.ExtractVersionFromFilename(file.Name())
		if err != nil {
			return nil, err
		}

		fileBytes, err := ioutil.ReadFile(path.Join("testtrack/migrate", file.Name()))
		if err != nil {
			return nil, err
		}

		var migrationFile serializers.MigrationFile
		err = yaml.Unmarshal(fileBytes, &migrationFile)
		if err != nil {
			return nil, err
		}

		if migrationFile.FeatureCompletion != nil {
			migrationsByVersion[migrationVersion] = featurecompletions.FromFile(&migrationVersion, migrationFile.FeatureCompletion)
		} else if migrationFile.RemoteKill != nil {
			migrationsByVersion[migrationVersion] = remotekills.FromFile(&migrationVersion, migrationFile.RemoteKill)
		} else if migrationFile.Split != nil {
			migrationsByVersion[migrationVersion], err = splits.FromFile(&migrationVersion, migrationFile.Split)
			if err != nil {
				return nil, err
			}
		} else if migrationFile.SplitRetirement != nil {
			migrationsByVersion[migrationVersion] = splitretirements.FromFile(&migrationVersion, migrationFile.SplitRetirement)
		} else if migrationFile.SplitDecision != nil {
			migrationsByVersion[migrationVersion] = splitdecisions.FromFile(&migrationVersion, migrationFile.SplitDecision)
		} else if migrationFile.IdentifierType != nil {
			migrationsByVersion[migrationVersion] = identifiertypes.FromFile(&migrationVersion, migrationFile.IdentifierType)
		} else {
			return nil, fmt.Errorf("testtrack/migrate/%s didn't match a known migration type", file.Name())
		}
	}
	return migrationsByVersion, nil
}

func getSortedVersions(migrationsByVersion map[string]migrations.IMigration) []string {
	versions := make([]string, 0, len(migrationsByVersion))

	for version := range migrationsByVersion {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions
}
