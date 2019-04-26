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
	var action = "create"
	if f.version == nil {
		action = "destroy"
	}

	filename := fmt.Sprintf("%s_%s_feature_completion_%s.yml", *f.migrationVersion, action, *f.featureGate)
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

// Serializable returns a JSON serializable representation
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

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (f *FeatureCompletion) SameResourceAs(other migrations.IMigration) bool {
	if otherF, ok := other.(*FeatureCompletion); ok {
		return *otherF.featureGate == *f.featureGate
	}
	return false
}

// Inverse returns a logical inverse operation if possible
func (f *FeatureCompletion) Inverse() (migrations.IMigration, error) {
	if f.version == nil {
		return nil, fmt.Errorf("can't invert feature_completion destroy %s for %s", *f.migrationVersion, *f.featureGate)
	}
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}
	return &FeatureCompletion{
		migrationVersion: migrationVersion,
		featureGate:      f.featureGate,
		version:          nil,
	}, nil
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (f *FeatureCompletion) ApplyToSchema(schema *serializers.Schema) error {
	if f.version == nil { // Delete
		for i, candidate := range schema.FeatureCompletions {
			if candidate.FeatureGate == *f.featureGate {
				schema.FeatureCompletions = append(schema.FeatureCompletions[:i], schema.FeatureCompletions[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("Couldn't locate feature_completion of %s in schema", *f.featureGate)
	}
	for i, candidate := range schema.FeatureCompletions { // Replace
		if candidate.FeatureGate == *f.featureGate {
			schema.FeatureCompletions[i] = *f.serializable()
		}
		return nil
	}
	schema.FeatureCompletions = append(schema.FeatureCompletions, *f.serializable()) // Add
	return nil
}
