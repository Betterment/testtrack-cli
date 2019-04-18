package featurecompletions

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
)

// FeatureCompletion represents a feature we're marking (un)completed
type FeatureCompletion struct {
	migrationVersion *string
	featureGate      *string
	version          *string
}

// New returns a migration object
func New(featureGate *string, version *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &FeatureCompletion{
		migrationVersion: migrationVersion,
		featureGate:      featureGate,
		version:          version,
	}, nil
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.FeatureCompletion) migrations.IMigration {
	return &FeatureCompletion{
		migrationVersion: migrationVersion,
		featureGate:      &serializable.FeatureGate,
		version:          serializable.Version,
	}
}

// Validate validates that a feature completion may be persisted
func (f *FeatureCompletion) Validate() error {
	err := validations.FeatureGate("feature_gate_name", f.featureGate)
	if err != nil {
		return err
	}

	err = validations.OptionalAppVersion("version", f.version)
	if err != nil {
		return err
	}

	return nil
}

// Filename generates a filename for this migration
func (f *FeatureCompletion) Filename() *string {
	var action = "complete"
	if f.version == nil {
		action = "uncomplete"
	}

	filename := fmt.Sprintf("%s_%s_feature_%s.yml", *f.migrationVersion, action, *f.featureGate)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (f *FeatureCompletion) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		FeatureCompletion: f.serializable(),
	}
}

// SyncPath returns the server path to post the migration to
func (f *FeatureCompletion) SyncPath() string {
	return "api/v2/migrations/app_feature_completion"
}

// Serializable returns a JSON/YAML serializable representation
func (f *FeatureCompletion) Serializable() interface{} {
	return f.serializable()
}

func (f *FeatureCompletion) serializable() *serializers.FeatureCompletion {
	return &serializers.FeatureCompletion{
		FeatureGate: *f.featureGate,
		Version:     f.version,
	}
}

// MigrationVersion returns the migration version
func (f *FeatureCompletion) MigrationVersion() *string {
	return f.migrationVersion
}
