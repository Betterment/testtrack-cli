package splits

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
	"gopkg.in/yaml.v2"
)

// SplitKey is the resource key for migrations impacting split state
type SplitKey string

// ISplitMigration defines the interface that allows split-impacting migration
// types to determine whether they operate on the same split
type ISplitMigration interface {
	ResourceKey() SplitKey
}

// Weights represents the weightings of a split
type Weights map[string]int

// Split represents a feature we're marking (un)completed
type Split struct {
	migrationVersion *string
	name             *string
	weights          *Weights
}

// New returns a migration object
func New(name *string, weights *Weights) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &Split{
		migrationVersion: migrationVersion,
		name:             name,
		weights:          weights,
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

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.SplitYAML) (migrations.IMigration, error) {
	weights, err := WeightsYAMLToMap(serializable.Weights)
	if err != nil {
		return nil, err
	}
	return &Split{
		migrationVersion: migrationVersion,
		name:             &serializable.Name,
		weights:          weights,
	}, nil
}

// WeightsYAMLToMap converts YAML-serializable weights to a weights map
func WeightsYAMLToMap(yamlWeights yaml.MapSlice) (*Weights, error) {
	weights := make(Weights)
	cumulativeWeight := 0
	for _, item := range yamlWeights {
		variant, ok := item.Key.(string)
		if !ok {
			return nil, fmt.Errorf("variant %v is not a string", item.Key)
		}
		weight, ok := item.Value.(int)
		if !ok {
			return nil, fmt.Errorf("weighting %v is not an int", item.Value)
		}
		if weight < 0 {
			return nil, fmt.Errorf("weight %d is less than zero", weight)
		}
		cumulativeWeight += weight
		weights[variant] = weight
	}
	if cumulativeWeight != 100 {
		return nil, fmt.Errorf("weights must sum to 100, got %d", cumulativeWeight)
	}
	return &weights, nil
}

// Validate validates that a feature completion may be persisted
func (s *Split) Validate() error {
	return validations.PrefixedSplit("name", s.name)
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
			Weights: WeightsMapToYAML(s.weights),
		},
	}
}

// WeightsMapToYAML converts weights to a YAML-serializable representation
func WeightsMapToYAML(weights *Weights) yaml.MapSlice {
	var variants = make([]string, 0, len(*weights))
	for variant := range *weights {
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	weightsYaml := make(yaml.MapSlice, 0, len(variants))
	for _, variant := range variants {
		weightsYaml = append(weightsYaml, yaml.MapItem{Key: variant, Value: (*weights)[variant]})
	}
	return weightsYaml
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

// Inverse returns a logical inverse operation if possible
func (s *Split) Inverse() (migrations.IMigration, error) {
	return nil, fmt.Errorf("can't invert split creation %s", *s.name)
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (s *Split) ApplyToSchema(schema *serializers.Schema) error {
	for i, candidate := range schema.Splits {
		if candidate.Name == *s.name {
			schema.Splits[i].Decided = false
			schemaWeights, err := WeightsYAMLToMap(candidate.Weights)
			if err != nil {
				return err
			}
			for variant := range *schemaWeights {
				(*schemaWeights)[variant] = 0
			}
			for variant, weight := range *s.weights {
				(*schemaWeights)[variant] = weight
			}
			schema.Splits[i].Weights = WeightsMapToYAML(schemaWeights)
			return nil
		}
	}
	schemaSplit := serializers.SchemaSplit{
		Name:    *s.name,
		Weights: WeightsMapToYAML(s.weights),
		Decided: false,
	}
	schema.Splits = append(schema.Splits, schemaSplit)
	return nil
}
