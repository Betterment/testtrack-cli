package migrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// FeatureCompletion represents a feature we're marking (un)completed
type FeatureCompletion struct {
	MigrationVersion *string
	FeatureGate      *string
	Version          *string
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

// NewFeatureCompletion returns a FeatureCompletion migration object
func NewFeatureCompletion(featureGate *string, version *string) FeatureCompletion {
	migrationVersion := generateMigrationVersion()
	return FeatureCompletion{
		MigrationVersion: &migrationVersion,
		FeatureGate:      featureGate,
		Version:          version,
	}
}

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

// PersistMigrationFile writes a migration to disk
func (f *FeatureCompletion) PersistMigrationFile() error {
	stat, err := os.Stat("testtrack/migrate")
	if err != nil {
		return errors.Wrap(err, "migration directory not found - run `testtrack init_project` to resolve")
	}

	if !stat.IsDir() {
		return errors.New("testtrack/migrate is not a directory")
	}

	serializable := f.Serializable()

	out, err := yaml.Marshal(serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		FeatureCompletion: &serializable,
	})

	err = ioutil.WriteFile(f.migrationFilename(), out, 0644)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMigrationFile deletes a migration file from disk
func (f *FeatureCompletion) DeleteMigrationFile() error {
	return os.Remove(f.migrationFilename())
}

func (f *FeatureCompletion) migrationFilename() string {
	var action = "complete"
	if f.Version == nil {
		action = "uncomplete"
	}

	return fmt.Sprintf("testtrack/migrate/%s_%s_feature_%s.yml", *f.MigrationVersion, action, *f.FeatureGate)
}

// Serializable returns a JSON/YAML serializable representation of a feature completion
func (f *FeatureCompletion) Serializable() serializers.FeatureCompletion {
	return serializers.FeatureCompletion{
		FeatureGate: *f.FeatureGate,
		Version:     f.Version,
	}
}

// SerializableMigrationVersion returns a serializable representation of the migration version
func (f *FeatureCompletion) SerializableMigrationVersion() serializers.MigrationVersion {
	return serializers.MigrationVersion{Version: *f.MigrationVersion}
}

func generateMigrationVersion() string {
	t := time.Now().UTC()
	nowEpochSeconds := t.Unix()
	todayEpochSeconds := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	secondsIntoToday := nowEpochSeconds - todayEpochSeconds
	return fmt.Sprintf("%04d%02d%02d%05d", t.Year(), t.Month(), t.Day(), secondsIntoToday)
}

// Sync sends a config change to the TestTrack server
func (f *FeatureCompletion) Sync() (bool, error) {
	resp, err := postToTestTrack("api/v2/migrations/app_feature_completion", f.Serializable())
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case 204:
		return true, nil
	case 422:
		return false, nil
	default:
		return false, fmt.Errorf("got %d status code", resp.StatusCode)
	}
}

func postToTestTrack(path string, body interface{}) (*http.Response, error) {
	ttURLString, ok := os.LookupEnv("TESTTRACK_CLI_URL")
	if !ok {
		return nil, errors.New("TESTTRACK_CLI_URL must be set")
	}

	ttURL, err := url.ParseRequestURI(ttURLString)
	ttURL.Path = strings.TrimRight(ttURL.Path, "/")

	url := strings.Join([]string{
		ttURL.String(),
		path,
	}, "/")

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.Post(url, "application/json", bytes.NewReader(bodyBytes))
}

// SyncMigrationVersion marks a migration version as run on TestTrack server
func (f *FeatureCompletion) SyncMigrationVersion() error {
	resp, err := postToTestTrack("api/v2/migrations", f.SerializableMigrationVersion())
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("got %d status code", resp.StatusCode)
	}

	return nil
}

// Save does the whole operation of validating, persisting, and sending a split config change to the local TT server
func (f *FeatureCompletion) Save() error {
	err := f.Validate()
	if err != nil {
		return err
	}

	err = f.PersistMigrationFile()
	if err != nil {
		return err
	}

	valid, err := f.Sync()
	if err != nil {
		return err
	}

	if !valid {
		f.DeleteMigrationFile()
		return errors.New("Migration unsuccessful on server. Does your feature flag exist?")
	}

	f.SyncMigrationVersion()

	return nil
}
