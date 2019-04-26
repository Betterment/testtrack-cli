package migrationloaders

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

// Load loads a set of migrations
func Load() (map[string]migrations.IMigration, error) {
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

// GetSortedVersions sorts and returns the migration versions from a map of migrations by version
func GetSortedVersions(migrationsByVersion map[string]migrations.IMigration) []string {
	versions := make([]string, 0, len(migrationsByVersion))

	for version := range migrationsByVersion {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions
}
