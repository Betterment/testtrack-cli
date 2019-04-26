package migrationrepositories

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/identifiertypes"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/splitdecisions"
	"github.com/Betterment/testtrack-cli/splitretirements"
	"github.com/Betterment/testtrack-cli/splits"
	"gopkg.in/yaml.v2"
)

// MigrationRepository is a map of migrations indexed by migration version
type MigrationRepository map[string]migrations.IMigration

// Load loads a set of migrations
func Load() (MigrationRepository, error) {
	files, err := ioutil.ReadDir("testtrack/migrate")
	if err != nil {
		return nil, err
	}

	migrationRepo := make(MigrationRepository)
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
			migrationRepo[migrationVersion] = featurecompletions.FromFile(&migrationVersion, migrationFile.FeatureCompletion)
		} else if migrationFile.RemoteKill != nil {
			migrationRepo[migrationVersion] = remotekills.FromFile(&migrationVersion, migrationFile.RemoteKill)
		} else if migrationFile.Split != nil {
			migrationRepo[migrationVersion], err = splits.FromFile(&migrationVersion, migrationFile.Split)
			if err != nil {
				return nil, err
			}
		} else if migrationFile.SplitRetirement != nil {
			migrationRepo[migrationVersion] = splitretirements.FromFile(&migrationVersion, migrationFile.SplitRetirement)
		} else if migrationFile.SplitDecision != nil {
			migrationRepo[migrationVersion] = splitdecisions.FromFile(&migrationVersion, migrationFile.SplitDecision)
		} else if migrationFile.IdentifierType != nil {
			migrationRepo[migrationVersion] = identifiertypes.FromFile(&migrationVersion, migrationFile.IdentifierType)
		} else {
			return nil, fmt.Errorf("testtrack/migrate/%s didn't match a known migration type", file.Name())
		}
	}
	return migrationRepo, nil
}

// SortedVersions sorts and returns the migration versions in a repo because
// maps don't preserve order in go
func (m *MigrationRepository) SortedVersions() []string {
	versions := make([]string, 0, len(*m))

	for version := range *m {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions
}
