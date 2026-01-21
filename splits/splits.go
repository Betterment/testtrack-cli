package splits

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
)

// SplitKey is the resource key for migrations impacting split state
type SplitKey string

// ISplitMigration defines the interface that allows split-impacting migration
// types to determine whether they operate on the same split
type ISplitMigration interface {
	ResourceKey() SplitKey
}

// Split represents a feature we're marking (un)completed
type Split struct {
	migrationVersion *string
	name             *string
	weights          *Weights
	owner            *string
}

// New returns a migration object
func New(name *string, weights *Weights, owner *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &Split{
		migrationVersion: migrationVersion,
		name:             name,
		weights:          weights,
		owner:            owner,
	}, nil
}

var weightRecordSeparatorRegex = regexp.MustCompile(`, *`)
var weightKeyValueSeparatorRegex = regexp.MustCompile(`: *`)

// WeightsFromString parses a `variant: 0, another_variant: 100`-style string into a weights map
func WeightsFromString(weights string) (*Weights, error) {
	weights = strings.Trim(weights, " ")
	weightRecords := weightRecordSeparatorRegex.Split(weights, -1)
	result := make(Weights)
	cumulativeWeight := 0
	for _, weightRecord := range weightRecords {
		weightKV := weightKeyValueSeparatorRegex.Split(weightRecord, 3)
		if len(weightKV) != 2 {
			return nil, fmt.Errorf("can't parse weight key/value pair %s", weightRecord)
		}
		variant := weightKV[0]
		err := validations.SnakeCaseParam("weighting variant", &variant)
		if err != nil {
			return nil, err
		}
		weightUint, err := strconv.ParseUint(weightKV[1], 10, 8)
		if err != nil {
			return nil, err
		}
		weight := int(weightUint)
		cumulativeWeight += weight
		result[variant] = weight
	}
	if cumulativeWeight != 100 {
		return nil, fmt.Errorf("weights must sum to 100, got %d", cumulativeWeight)
	}
	return &result, nil
}

// IsFeatureGateFromName returns true if name ends with '_enabled'
func IsFeatureGateFromName(name string) bool {
	return strings.HasSuffix(name, "_enabled")
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.SplitYAML) (migrations.IMigration, error) {
	weights, err := NewWeights(serializable.Weights)
	if err != nil {
		return nil, err
	}
	return &Split{
		migrationVersion: migrationVersion,
		name:             &serializable.Name,
		owner:            &serializable.Owner,
		weights:          weights,
	}, nil
}

// Validate validates that a feature completion may be persisted
func (s *Split) Validate() error {
	return validations.Split("name", s.name)
}

// Filename generates a filename for this migration
func (s *Split) Filename() *string {
	filename := fmt.Sprintf("%s_create_split_%s.yml", *s.migrationVersion, *s.name)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (s *Split) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		Split: &serializers.SplitYAML{
			Name:    *s.name,
			Weights: *s.weights,
			Owner:   *s.owner,
		},
	}
}

// SyncPath returns the server path to post the migration to
func (s *Split) SyncPath() string {
	return "api/v2/migrations/split"
}

// Serializable returns a JSON-serializable representation
func (s *Split) Serializable() interface{} {
	return &serializers.SplitJSON{
		Name:              *s.name,
		WeightingRegistry: *s.weights,
	}
}

// MigrationVersion returns the migration version
func (s *Split) MigrationVersion() *string {
	return s.migrationVersion
}

// ResourceKey returns the natural key of the resource under migration
func (s *Split) ResourceKey() SplitKey {
	return SplitKey(*s.name)
}

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (s *Split) SameResourceAs(other migrations.IMigration) bool {
	if otherS, ok := other.(ISplitMigration); ok {
		return otherS.ResourceKey() == s.ResourceKey()
	}
	return false
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (s *Split) ApplyToSchema(schema *serializers.Schema, migrationRepo migrations.Repository, _idempotently bool) error {
	for i, candidate := range schema.Splits { // Replace
		if candidate.Name == *s.name {
			schemaWeights, err := NewWeights(candidate.Weights)
			if err != nil {
				return err
			}
			schemaWeights.Merge(*s.weights)
			schema.Splits[i].Decided = false
			schema.Splits[i].Weights = *schemaWeights
			return nil
		}
	}
	if s.migrationVersion != nil { // Revive weights from old migration
		split := MostRecentNamed(*s.name, *s.migrationVersion, migrationRepo)
		if split != nil {
			weights := split.Weights()
			weights.Merge(*s.weights)
			schema.Splits = append(schema.Splits, serializers.SchemaSplit{
				Name:    *s.name,
				Weights: *weights,
				Decided: false,
			})
			return nil
		}
	}
	schemaSplit := serializers.SchemaSplit{ // Create
		Name:    *s.name,
		Weights: *s.weights,
		Decided: false,
		Owner:   *s.owner,
	}
	schema.Splits = append(schema.Splits, schemaSplit)
	return nil
}

// Weights of the split
func (s *Split) Weights() *Weights {
	return s.weights
}

// MostRecentNamed returns the most recent matching migration in a repo
func MostRecentNamed(name, migrationVersion string, migrationRepo migrations.Repository) *Split {
	versions := migrationRepo.SortedVersions()
	migrationIndex := -1
	for i, version := range versions {
		if version == migrationVersion {
			migrationIndex = i
			break
		}
	}
	if migrationIndex < 1 {
		return nil
	}
	for i := migrationIndex - 1; i >= 0; i-- {
		split, ok := migrationRepo[versions[i]].(*Split)
		if ok && *split.name == name {
			return split
		}
	}
	return nil
}
