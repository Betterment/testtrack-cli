package schemaloaders

import (
	"fmt"
	"reflect"

	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/identifiertypes"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrationrepositories"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/Betterment/testtrack-cli/splitdecisions"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// SchemaLoader loads schemas into TestTrack
type SchemaLoader struct {
	server        servers.IServer
	schema        *serializers.Schema
	migrationRepo *migrationrepositories.MigrationRepository
}

// New returns a SchemaLoader ready to use
func New() (*SchemaLoader, error) {
	server, err := servers.New()
	if err != nil {
		return nil, err
	}

	schema, err := schema.Read()
	if err != nil {
		return nil, err
	}

	migrationRepo, err := migrationrepositories.Load()
	if err != nil {
		return nil, err
	}

	return &SchemaLoader{server: server, schema: schema, migrationRepo: &migrationRepo}, nil
}

// Load the schema into TestTrack server, marking all migrations as applied
func (s *SchemaLoader) Load() error {
	migrations := []migrations.IMigration{}

	for i := range s.schema.IdentifierTypes {
		migrations = append(migrations, identifiertypes.FromFile(nil, &s.schema.IdentifierTypes[i]))
	}
	for _, split := range s.schema.Splits {
		splitMigrations, err := schemaSplitMigrations(split)
		if err != nil {
			return err
		}
		migrations = append(migrations, splitMigrations...)
	}
	for i := range s.schema.RemoteKills {
		migrations = append(migrations, remotekills.FromFile(nil, &s.schema.RemoteKills[i]))
	}
	for i := range s.schema.FeatureCompletions {
		migrations = append(migrations, featurecompletions.FromFile(nil, &s.schema.FeatureCompletions[i]))
	}

	newSchema := &serializers.Schema{
		SerializerVersion: serializers.SerializerVersion,
		SchemaVersion:     s.schema.SchemaVersion,
	}
	for _, migration := range migrations {
		err := migrationmanagers.NewWithDependencies(migration, s.server, newSchema).Apply()
		if err != nil {
			return err
		}
	}

	schema.SortAlphabetically(newSchema)
	if !reflect.DeepEqual(*s.schema, *newSchema) {
		before, err := yaml.Marshal(s.schema)
		if err != nil {
			return err
		}
		after, err := yaml.Marshal(newSchema)
		if err != nil {
			return err
		}
		return fmt.Errorf("testtrack bug! load resulted in different schema.\n\nBefore:\n\n%s\n\nAfter:\n\n%s", before, after)
	}

	for _, version := range s.migrationRepo.SortedVersions() {
		if version > s.schema.SchemaVersion {
			fmt.Println("Schema load complete, but there are migrations newer than the schema file - run testtrack migrate to apply them.")
			break
		}
		err := migrationmanagers.NewWithDependencies((*s.migrationRepo)[version], s.server, newSchema).SyncVersion()
		if err != nil {
			return err
		}
	}

	return nil
}

func schemaSplitMigrations(schemaSplit serializers.SchemaSplit) ([]migrations.IMigration, error) {
	split, err := splits.FromFile(nil, &serializers.SplitYAML{
		Name:    schemaSplit.Name,
		Weights: schemaSplit.Weights,
	})
	if err != nil {
		return nil, err
	}

	migrations := []migrations.IMigration{split}

	if schemaSplit.Decided {
		var decision *string
		weights, err := splits.WeightsYAMLToMap(schemaSplit.Weights)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("schema split %s invalid", schemaSplit.Name))
		}
		for variant, weight := range *weights {
			if weight == 100 {
				decision = &variant
			}
		}
		if decision == nil {
			return nil, fmt.Errorf("decided schema split %s has no 100%% weighted variant", schemaSplit.Name)
		}
		migrations = append(migrations, splitdecisions.FromFile(nil, &serializers.SplitDecision{
			Split:   schemaSplit.Name,
			Variant: *decision,
		}))
	}
	return migrations, nil
}
