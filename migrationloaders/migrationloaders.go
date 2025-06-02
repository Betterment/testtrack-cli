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

// Load loads a set of migrations
func Load() (migrations.Repository, error) {
	files, err := os.ReadDir("testtrack/migrate")
	if err != nil {
		return nil, err
	}

	migrationRepo := make(migrations.Repository)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue // Skip hidden files
		}

		migrationVersion, err := migrations.ExtractVersionFromFilename(file.Name())
		if err != nil {
			return nil, err
		}

		fileBytes, err := os.ReadFile(path.Join("testtrack/migrate", file.Name()))
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
