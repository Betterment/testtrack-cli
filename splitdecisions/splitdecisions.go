package splitdecisions

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/pkg/errors"
)

// SplitDecision represents a feature we're marking (un)completed
type SplitDecision struct {
	migrationVersion *string
	split            *string
	variant          *string
}

// New returns a migration object
func New(split, variant *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &SplitDecision{
		migrationVersion: migrationVersion,
		split:            split,
		variant:          variant,
	}, nil
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.SplitDecision) migrations.IMigration {
	return &SplitDecision{
		migrationVersion: migrationVersion,
		split:            &serializable.Split,
		variant:          &serializable.Variant,
	}
}

// Validate validates that a migration may be persisted
func (s *SplitDecision) Validate() error {
	return validations.Split("split", s.split)
}

// Filename generates a filename for this migration
func (s *SplitDecision) Filename() *string {
	filename := fmt.Sprintf("%s_create_split_decision_%s.yml", *s.migrationVersion, *s.split)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (s *SplitDecision) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		SplitDecision: &serializers.SplitDecision{
			Split:   *s.split,
			Variant: *s.variant,
		},
	}
}

// SyncPath returns the server path to post the migration to
func (s *SplitDecision) SyncPath() string {
	return "api/v2/migrations/split_decision"
}

// Serializable returns a JSON-serializable representation
func (s *SplitDecision) Serializable() interface{} {
	return &serializers.SplitDecision{
		Split:   *s.split,
		Variant: *s.variant,
	}
}

// MigrationVersion returns the migration version
func (s *SplitDecision) MigrationVersion() *string {
	return s.migrationVersion
}

// ResourceKey returns the natural key of the resource under migration
func (s *SplitDecision) ResourceKey() splits.SplitKey {
	return splits.SplitKey(*s.split)
}

// Weights represents the weightings of a split
func (s *SplitDecision) Weights() *splits.Weights {
	return nil // SplitDecisions don't have weights
}

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (s *SplitDecision) SameResourceAs(other migrations.IMigration) bool {
	if otherS, ok := other.(splits.ISplitMigration); ok {
		return otherS.ResourceKey() == s.ResourceKey()
	}
	return false
}

// Inverse returns a logical inverse operation if possible
func (s *SplitDecision) Inverse() (migrations.IMigration, error) {
	return nil, fmt.Errorf("can't invert split decision %s", *s.split)
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (s *SplitDecision) ApplyToSchema(schema *serializers.Schema, migrationRepo migrations.Repository) error {
	for i, candidate := range schema.Splits {
		if candidate.Name == *s.split {
			schema.Splits[i].Decided = true
			weights, err := splits.WeightsFromYAML(candidate.Weights)
			if err != nil {
				return err
			}
			err = weights.ReweightToDecision(*s.variant)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("in split %s in schema", *s.split))
			}
			schema.Splits[i].Weights = weights.ToYAML()
			return nil
		}
	}
	if s.migrationVersion != nil {
		split := splits.MostRecentNamed(*s.split, *s.migrationVersion, migrationRepo)
		if split != nil {
			weights := split.Weights()
			err := weights.ReweightToDecision(*s.variant)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("in most recent split migration %s", *s.split))
			}
			schema.Splits = append(schema.Splits, serializers.SchemaSplit{
				Name:    *s.split,
				Weights: weights.ToYAML(),
				Decided: true,
			})
			return nil
		}
	}
	return fmt.Errorf("Couldn't locate split %s in schema to decide", *s.split)
}
