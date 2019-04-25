package serializers

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// SerializerVersion is the current version of the migration file format so we can evolve over time
const SerializerVersion = 1

// MigrationVersion is a JSON-marshalable representation of migration version (timestamp)
type MigrationVersion struct {
	Version string `json:"version"`
}

// MigrationFile is the YAML-marshalable root of a migration file
type MigrationFile struct {
	SerializerVersion int                `yaml:"serializer_version"`
	FeatureCompletion *FeatureCompletion `yaml:"feature_completion,omitempty"`
	RemoteKill        *RemoteKill        `yaml:"remote_kill,omitempty"`
	Split             *SplitYAML         `yaml:"split,omitempty"`
	SplitRetirement   *SplitRetirement   `yaml:"split_retirement,omitempty"`
	SplitDecision     *SplitDecision     `yaml:"split_decision,omitempty"`
	IdentifierType    *IdentifierType    `yaml:"identifier_type,omitempty"`
}

// FeatureCompletion is the marshalable representation of a FeatureCompletion
type FeatureCompletion struct {
	FeatureGate string  `yaml:"feature_gate" json:"feature_gate"`
	Version     *string `yaml:"version" json:"version"`
}

// RemoteKill is the marshalable representation of a RemoteKill
type RemoteKill struct {
	Split           string  `yaml:"split" json:"split"`
	Reason          string  `yaml:"reason" json:"reason"`
	OverrideTo      *string `yaml:"override_to" json:"override_to"`
	FirstBadVersion *string `yaml:"first_bad_version" json:"first_bad_version"`
	FixedVersion    *string `yaml:"fixed_version" json:"fixed_version"`
}

// SplitYAML is the YAML-marshalable representation of a Split
type SplitYAML struct {
	Name    string        `yaml:"name"`
	Weights yaml.MapSlice `yaml:"weights"`
}

// SplitJSON is is the JSON-marshalabe representation of a Split
type SplitJSON struct {
	Name              string         `json:"name"`
	WeightingRegistry map[string]int `json:"weighting_registry"`
}

// SplitRetirement is the JSON and YAML-marshalable representation of a SplitRetirement
type SplitRetirement struct {
	Split    string `json:"split"`
	Decision string `json:"decision"`
}

// SplitDecision is the JSON and YAML-marshalable representation of a SplitDecision
type SplitDecision struct {
	Split   string `json:"split"`
	Variant string `json:"variant"`
}

// IdentifierType is the JSON and YAML-marshalable representation of an IdentifierType
type IdentifierType struct {
	Name string `json:"name"`
}

// SchemaSplit is the schema-file YAML-marshalable representation of a split's state
type SchemaSplit struct {
	Name    string        `yaml:"name"`
	Weights yaml.MapSlice `yaml:"weights"`
	Decided bool          `yaml:"decided"`
}

// Schema is the YAML-marshalable representation of the TestTrack schema for
// migration validation and bootstrapping of new ecosystems
type Schema struct {
	SerializerVersion  int                 `yaml:"serializer_version"`
	SchemaVersion      string              `yaml:"schema_version"`
	RemoteKills        []RemoteKill        `yaml:"remote_kills,omitempty"`
	FeatureCompletions []FeatureCompletion `yaml:"feature_completions,omitempty"`
	Splits             []SchemaSplit       `yaml:"splits,omitempty"`
	IdentifierTypes    []IdentifierType    `yaml:"identifier_types,omitempty"`
}

// LoadSchema loads from disk or instantiates an empty Schema struct
func LoadSchema() (*Schema, error) {
	if _, err := os.Stat("testtrack/schema.yml"); os.IsNotExist(err) {
		return &Schema{SerializerVersion: SerializerVersion}, nil
	}
	schemaBytes, err := ioutil.ReadFile("testtrack/schema.yml")
	if err != nil {
		return nil, err
	}
	var schema Schema
	err = yaml.Unmarshal(schemaBytes, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// DumpSchema dumps to disk or deletes a schema file if unneeded
func DumpSchema(schema *Schema) error {
	if len(schema.SchemaVersion) == 0 {
		if _, err := os.Stat("testtrack/schema.yml"); os.IsNotExist(err) {
			return nil
		}
		err := os.Remove("testtrack/schema.yml")
		if err != nil {
			return err
		}
	}

	out, err := yaml.Marshal(schema)

	err = ioutil.WriteFile("testtrack/schema.yml", out, 0644)
	if err != nil {
		return err
	}

	return nil
}
