package serializers

import "gopkg.in/yaml.v2"

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
