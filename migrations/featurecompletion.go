package migrations

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// FeatureCompletion represents a feature we're marking completed
type FeatureCompletion struct {
	FeatureGate *string
	Version     *string
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

// Validate validates that a feature completion may be persisted
func (f *FeatureCompletion) Validate() error {
	if !featureGateRegex.MatchString(*f.FeatureGate) {
		return fmt.Errorf("feature_gate '%s' must be snake_case alphanumeric and end in _enabled", *f.FeatureGate)
	}

	if len(*f.FeatureGate) > featureGateMaxLength {
		return fmt.Errorf("feature_gate '%s' must be %d characters or less", *f.FeatureGate, featureGateMaxLength)
	}

	if f.Version != nil {
		if !appVersionRegex.MatchString(*f.Version) {
			return fmt.Errorf("version '%s' must be made up of no more than three integers with dots in between", *f.Version)
		}

		if len(*f.Version) > appVersionMaxLength {
			return fmt.Errorf("version '%s' must be %d characters or less", *f.Version, appVersionMaxLength)
		}
	}

	return nil
}

// PersistMigration writes a migration to disk
func (f *FeatureCompletion) PersistMigration() error {
	stat, err := os.Stat("testtrack/migrate")
	if err != nil {
		return errors.Wrap(err, "migration directory not found - run `testtrack init_project` to resolve")
	}

	if !stat.IsDir() {
		return errors.New("testtrack/migrate is not a directory")
	}

	serialized := serializers.FeatureCompletion{
		FeatureGate: *f.FeatureGate,
		Version:     f.Version,
	}

	out, err := yaml.Marshal(serializers.Migration{
		SerializerVersion: serializers.SerializerVersion,
		FeatureCompletion: &serialized,
	})

	var action = "complete"
	if f.Version == nil {
		action = "uncomplete"
	}

	err = ioutil.WriteFile(fmt.Sprintf("testtrack/migrate/%s_%s_feature_%s.yml", MigrationTimestamp(), action, *f.FeatureGate), out, 0644)
	if err != nil {
		return err
	}

	return nil
}

// MigrationTimestamp returns a rails-style string-collatable timestamp identifier for prefixing migration filenames
func MigrationTimestamp() string {
	t := time.Now().UTC()
	epochSeconds := t.Unix()
	todayEpochSeconds := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	secondsIntoToday := epochSeconds - todayEpochSeconds
	return fmt.Sprintf("%04d%02d%02d%05d", t.Year(), t.Month(), t.Day(), secondsIntoToday)
}

// Save does the whole operation of validating, persisting, and sending a split config change to the local TT server
func (f *FeatureCompletion) Save() error {
	err := f.Validate()
	if err != nil {
		return err
	}

	err = f.PersistMigration()
	if err != nil {
		return err
	}

	return nil
}
