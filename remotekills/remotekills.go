package remotekills

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
)

// RemoteKill represents a feature we're setting (or unsetting) killed for a range of app versions
type RemoteKill struct {
	migrationVersion *string
	split            *string
	reason           *string
	overrideTo       *string
	firstBadVersion  *string
	fixedVersion     *string
}

// New returns a migration object
func New(split, reason, overrideTo, firstBadVersion, fixedVersion *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &RemoteKill{
		migrationVersion: migrationVersion,
		split:            split,
		reason:           reason,
		overrideTo:       overrideTo,
		firstBadVersion:  firstBadVersion,
		fixedVersion:     fixedVersion,
	}, nil
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.RemoteKill) migrations.IMigration {
	return &RemoteKill{
		migrationVersion: migrationVersion,
		split:            &serializable.Split,
		reason:           &serializable.Reason,
		overrideTo:       serializable.OverrideTo,
		firstBadVersion:  serializable.FirstBadVersion,
		fixedVersion:     serializable.FixedVersion,
	}
}

// Validate validates that a feature completion may be persisted
func (r *RemoteKill) Validate() error {
	err := validations.Split(r.split)
	if err != nil {
		return err
	}

	reasonParam := "reason"
	err = validations.SnakeCaseParam(r.split, &reasonParam)
	if err != nil {
		return err
	}

	overrideToParam := "override_to"
	err = validations.OptionalSnakeCaseParam(r.split, &overrideToParam)
	if err != nil {
		return err
	}

	firstBadVersionParam := "first_bad_version"
	err = validations.OptionalAppVersion(r.firstBadVersion, &firstBadVersionParam)
	if err != nil {
		return err
	}

	fixedVersionParam := "fixed_version"
	err = validations.OptionalAppVersion(r.fixedVersion, &fixedVersionParam)
	if err != nil {
		return err
	}

	return nil
}

// Filename generates a filename for this migration
func (r *RemoteKill) Filename() *string {
	var action = "set"
	if r.firstBadVersion == nil {
		action = "unset"
	}

	filename := fmt.Sprintf("%s_%s_split_%s_remote_kill.yml", *r.migrationVersion, action, *r.split)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (r *RemoteKill) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		RemoteKill:        r.serializable(),
	}
}

// SyncPath returns the server path to post the migration to
func (r *RemoteKill) SyncPath() string {
	return "api/v2/migrations/app_remote_kill"
}

// Serializable returns a JSON/YAML serializable representation
func (r *RemoteKill) Serializable() interface{} {
	return r.serializable()
}

func (r *RemoteKill) serializable() *serializers.RemoteKill {
	return &serializers.RemoteKill{
		Split:           *r.split,
		Reason:          *r.reason,
		OverrideTo:      r.overrideTo,
		FirstBadVersion: r.firstBadVersion,
		FixedVersion:    r.fixedVersion,
	}
}

// MigrationVersion returns the migration version
func (r *RemoteKill) MigrationVersion() *string {
	return r.migrationVersion
}
