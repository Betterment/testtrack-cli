package serializers

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
