package serializers

import (
	"gopkg.in/yaml.v2"
)

// SerializerVersion is the current version of the schema file format so we
// can evolve it over time.
//
// Version 2 replaced the single scalar schema_version high-water-mark with a
// schema_versions list of every applied migration version, eliminating the
// merge-conflict hotspot that the scalar created (every new migration rewrote
// the same line). The list is ordered by a hash of each version so concurrently
// added migrations scatter through the file instead of clustering.
const SerializerVersion = 2

// MigrationSerializerVersion is the current version of the migration file
// format. It is versioned independently of the schema file: the schema format
// changed in v2 but migration files didn't, and stamping them with the schema's
// version would burn version numbers for a format that hasn't evolved.
const MigrationSerializerVersion = 1

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
	Name    string         `yaml:"name"`
	Weights map[string]int `yaml:"weights"`
	Owner   string         `yaml:"owner,omitempty"`
}

// SplitJSON is the JSON-marshalabe representation of a Split
type SplitJSON struct {
	Name              string         `json:"name"`
	WeightingRegistry map[string]int `json:"weighting_registry"`
}

// RemoteRegistrySplit is the JSON-marshalable representation of a server-provided split configuration
type RemoteRegistrySplit struct {
	Weights map[string]int `json:"weights"`
}

// RemoteRegistry is the JSON-marshalable representation of a server-provided split registry
type RemoteRegistry struct {
	Splits map[string]RemoteRegistrySplit `json:"splits"`
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
	Name string `yaml:"name" json:"name"`
}

// SchemaSplit is the schema-file YAML-marshalable representation of a split's state
type SchemaSplit struct {
	Name    string         `yaml:"name" json:"name"`
	Weights map[string]int `yaml:"weights" json:"weights"`
	Decided bool           `yaml:"decided,omitempty" json:"decided,omitempty"`
	Owner   string         `yaml:"owner,omitempty" json:"owner,omitempty"`
}

// Schema is the YAML-marshalable representation of the TestTrack schema for
// migration validation and bootstrapping of new ecosystems
type Schema struct {
	SerializerVersion int      `yaml:"serializer_version" json:"serializer_version"`
	SchemaVersions    []string `yaml:"schema_versions,omitempty" json:"schema_versions,omitempty"`
	// LegacySchemaVersion is the v1 scalar high-water mark. It only exists so
	// reads can detect a file that carries it — either a plain v1 schema, or a
	// hybrid produced by a pre-2.0 CLI rewriting a v2 schema (those round-trip
	// serializer_version: 2 while writing the v1 shape). It's a pointer because
	// detection is by key presence, not value: 1.x always writes the key (no
	// omitempty), and rewrite paths that never set the scalar (e.g. 1.x `sync`)
	// emit schema_version: "", which must still trip the guard. It is never
	// written: `schema upgrade` clears it and omitempty drops nil from output.
	LegacySchemaVersion *string             `yaml:"schema_version,omitempty" json:"schema_version,omitempty"`
	Splits              []SchemaSplit       `yaml:"splits,omitempty" json:"splits,omitempty"`
	IdentifierTypes     []IdentifierType    `yaml:"identifier_types,omitempty" json:"identifier_types,omitempty"`
	RemoteKills         []RemoteKill        `yaml:"remote_kills,omitempty" json:"remote_kills,omitempty"`
	FeatureCompletions  []FeatureCompletion `yaml:"feature_completions,omitempty" json:"feature_completions,omitempty"`
}

// AddVersion records a migration version as applied in the schema, ignoring
// duplicates. Ordering is handled at write time, so callers needn't sort.
func (s *Schema) AddVersion(version string) {
	for _, v := range s.SchemaVersions {
		if v == version {
			return
		}
	}
	s.SchemaVersions = append(s.SchemaVersions, version)
}

// LegacySchema represents the Rails migration-piggybacked testtrack schema files of old
type LegacySchema struct {
	IdentifierTypes []string      `yaml:"identifier_types"`
	Splits          yaml.MapSlice `yaml:"splits"`
}
