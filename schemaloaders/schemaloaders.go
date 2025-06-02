package schemaloaders

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/identifiertypes"
	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/Betterment/testtrack-cli/splitdecisions"
	"github.com/Betterment/testtrack-cli/splits"
)

// SchemaLoader loads schemas into TestTrack
type SchemaLoader struct {
	server        servers.IServer
	schema        *serializers.Schema
	migrationRepo *migrations.Repository
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

	migrationRepo, err := migrationloaders.Load()
	if err != nil {
		return nil, err
	}

	return &SchemaLoader{server: server, schema: schema, migrationRepo: &migrationRepo}, nil
}

// Load the schema into TestTrack server, marking all migrations as applied
func (s *SchemaLoader) Load() error {
	ms := []migrations.IMigration{}

	for i := range s.schema.IdentifierTypes {
		ms = append(ms, identifiertypes.FromFile(nil, &s.schema.IdentifierTypes[i]))
	}
	for _, split := range s.schema.Splits {
		splitMigrations, err := schemaSplitMigrations(split)
		if err != nil {
			return err
		}
		ms = append(ms, splitMigrations...)
	}
	for i := range s.schema.RemoteKills {
		ms = append(ms, remotekills.FromFile(nil, &s.schema.RemoteKills[i]))
	}
	for i := range s.schema.FeatureCompletions {
		ms = append(ms, featurecompletions.FromFile(nil, &s.schema.FeatureCompletions[i]))
	}

	for _, migration := range ms {
		err := migrationmanagers.NewWithServer(migration, s.server).Sync()
		if err != nil {
			return err
		}
	}

	for _, version := range s.migrationRepo.SortedVersions() {
		if version > s.schema.SchemaVersion {
			fmt.Println("Schema load complete, but there are migrations newer than the schema file - run testtrack migrate to apply them.")
			break
		}
		err := migrationmanagers.NewWithServer((*s.migrationRepo)[version], s.server).SyncVersion()
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
		weights, err := splits.WeightsFromYAML(schemaSplit.Weights)
		if err != nil {
			return nil, fmt.Errorf("schema split %s invalid: %w", schemaSplit.Name, err)
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
