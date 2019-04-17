package featurecompletions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
)

// FeatureCompletion represents a feature we're marking (un)completed
type FeatureCompletion struct {
	migrationVersion *string
	featureGate      *string
	version          *string
}

var featureGateRegex = regexp.MustCompile(`^[a-z_\d]+_enabled$`)
var featureGateMaxLength = 128 // This is arbitrary but way bigger than you need and smaller than the column will fit
var decimalIntegerRegexPart = `(?:0|[1-9]\d*)`
var appVersionRegex = regexp.MustCompile(strings.Join([]string{
	`^(?:`,
	decimalIntegerRegexPart,
	`\.){0,2}`,
	decimalIntegerRegexPart,
	`$`,
}, ""))

var appVersionMaxLength = 18 // This conforms to iOS version numering rules

// New returns a FeatureCompletion migration object
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
	if !featureGateRegex.MatchString(*f.featureGate) {
		return fmt.Errorf("feature_gate '%s' must be snake_case alphanumeric and end in _enabled", *f.featureGate)
	}

	if len(*f.featureGate) > featureGateMaxLength {
		return fmt.Errorf("feature_gate '%s' must be %d characters or less", *f.featureGate, featureGateMaxLength)
	}

	if f.version != nil {
		if !appVersionRegex.MatchString(*f.version) {
			return fmt.Errorf("version '%s' must be made up of no more than three integers with dots in between", *f.version)
		}

		if len(*f.version) > appVersionMaxLength {
			return fmt.Errorf("version '%s' must be %d characters or less", *f.version, appVersionMaxLength)
		}
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

// MigrationVersion returns the migration version
func (f *FeatureCompletion) MigrationVersion() *string {
	return f.migrationVersion
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

// ServerPath returns the path to post the migration to
func (f *FeatureCompletion) ServerPath() *string {
	serverPath := "api/v2/migrations/app_feature_completion"
	return &serverPath
}
