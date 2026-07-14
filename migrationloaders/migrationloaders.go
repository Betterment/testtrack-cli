package migrationloaders

import (
	"fmt"
	"os"
	"path"
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

// Filenames lists the migration filenames in testtrack/migrate, skipping
// hidden files and directories. It is the single source of truth for which
// directory entries count as migrations — Load and `schema upgrade` must
// agree on that, or the schema's recorded versions drift from what the
// loader sees.
func Filenames() ([]string, error) {
	files, err := os.ReadDir("testtrack/migrate")
	if err != nil {
		return nil, err
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

// Versions returns the unique migration versions recorded in the filenames in
// testtrack/migrate, without reading file contents — usable even on repos
// whose migrations can't be parsed or replayed.
func Versions() ([]string, error) {
	filenames, err := Filenames()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(filenames))
	versions := make([]string, 0, len(filenames))
	for _, filename := range filenames {
		version, err := migrations.ExtractVersionFromFilename(filename)
		if err != nil {
			return nil, fmt.Errorf("%w - delete or rename it if it isn't a testtrack migration", err)
		}
		if seen[version] {
			continue
		}
		seen[version] = true
		versions = append(versions, version)
	}
	return versions, nil
}

// Load loads a set of migrations
func Load() (migrations.Repository, error) {
	filenames, err := Filenames()
	if err != nil {
		return nil, err
	}

	migrationRepo := make(migrations.Repository)
	for _, filename := range filenames {
		migrationVersion, err := migrations.ExtractVersionFromFilename(filename)
		if err != nil {
			return nil, err
		}

		fileBytes, err := os.ReadFile(path.Join("testtrack/migrate", filename))
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
			return nil, fmt.Errorf("testtrack/migrate/%s didn't match a known migration type", filename)
		}
	}
	return migrationRepo, nil
}
