package splitretirements

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/Betterment/testtrack-cli/validations"
)

// SplitRetirement represents a feature we're marking (un)completed
type SplitRetirement struct {
	migrationVersion *string
	split            *string
	decision         *string
}

// New returns a migration object
func New(split, decision *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &SplitRetirement{
		migrationVersion: migrationVersion,
		split:            split,
		decision:         decision,
	}, nil
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.SplitRetirement) migrations.IMigration {
	return &SplitRetirement{
		migrationVersion: migrationVersion,
		split:            &serializable.Split,
		decision:         &serializable.Decision,
	}
}

// Validate validates that a feature completion may be persisted
func (s *SplitRetirement) Validate() error {
	return validations.Split("split", s.split)
}

// Filename generates a filename for this migration
func (s *SplitRetirement) Filename() *string {
	filename := fmt.Sprintf("%s_create_split_retirement_%s.yml", *s.migrationVersion, *s.split)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (s *SplitRetirement) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		SplitRetirement: &serializers.SplitRetirement{
			Split:    *s.split,
			Decision: *s.decision,
		},
	}
}

// SyncPath returns the server path to post the migration to
func (s *SplitRetirement) SyncPath() string {
	return "api/v2/migrations/split_retirement"
}

// Serializable returns a JSON-serializable representation
func (s *SplitRetirement) Serializable() interface{} {
	return &serializers.SplitRetirement{
		Split:    *s.split,
		Decision: *s.decision,
	}
}

// MigrationVersion returns the migration version
func (s *SplitRetirement) MigrationVersion() *string {
	return s.migrationVersion
}

// ResourceKey returns the natural key of the resource under migration
func (s *SplitRetirement) ResourceKey() splits.SplitKey {
	return splits.SplitKey(*s.split)
}

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (s *SplitRetirement) SameResourceAs(other migrations.IMigration) bool {
	if otherS, ok := other.(splits.ISplitMigration); ok {
		return otherS.ResourceKey() == s.ResourceKey()
	}
	return false
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (s *SplitRetirement) ApplyToSchema(schema *serializers.Schema, _ migrations.Repository, _idempotently bool) error {
	for i, candidate := range schema.Splits {
		if candidate.Name == *s.split {
			weights, err := splits.WeightsFromYAML(candidate.Weights)
			if err != nil {
				return err
			}
			err = weights.ReweightToDecision(*s.decision)
			if err != nil {
				return fmt.Errorf("in split %s in schema: %w", *s.split, err)
			}
			schema.Splits = append(schema.Splits[:i], schema.Splits[i+1:]...)
			return nil
		}
	}
	return nil
}
