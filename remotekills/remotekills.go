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
	err := validations.Split("split_name", r.split)
	if err != nil {
		return err
	}

	err = validations.SnakeCaseParam("reason", r.reason)
	if err != nil {
		return err
	}

	err = validations.OptionalSnakeCaseParam("override_to", r.overrideTo)
	if err != nil {
		return err
	}

	err = validations.OptionalAppVersion("first_bad_version", r.firstBadVersion)
	if err != nil {
		return err
	}

	err = validations.OptionalAppVersion("fixed_version", r.fixedVersion)
	if err != nil {
		return err
	}

	return nil
}

// Filename generates a filename for this migration
func (r *RemoteKill) Filename() *string {
	var action = "create"
	if r.firstBadVersion == nil {
		action = "destroy"
	}

	filename := fmt.Sprintf("%s_%s_remote_kill_%s_%s.yml", *r.migrationVersion, action, *r.split, *r.reason)
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

// Serializable returns a JSON serializable representation
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

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (r *RemoteKill) SameResourceAs(other migrations.IMigration) bool {
	if otherR, ok := other.(*RemoteKill); ok {
		return *otherR.split == *r.split &&
			*otherR.reason == *r.reason
	}
	return false
}

// Inverse returns a logical inverse operation if possible
func (r *RemoteKill) Inverse() (migrations.IMigration, error) {
	if r.firstBadVersion == nil {
		return nil, fmt.Errorf("can't invert remote_kill destroy %s for %s %s", *r.migrationVersion, *r.split, *r.reason)
	}
	return &RemoteKill{
		split:           r.split,
		reason:          r.reason,
		overrideTo:      nil,
		firstBadVersion: nil,
		fixedVersion:    nil,
	}, nil
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (r *RemoteKill) ApplyToSchema(schema *serializers.Schema) error {
	if r.firstBadVersion == nil { // Delete
		for i, candidate := range schema.RemoteKills {
			if candidate.Split == *r.split && candidate.Reason == *r.reason {
				schema.RemoteKills = append(schema.RemoteKills[:i], schema.RemoteKills[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("Couldn't locate remote_kill %s of %s in schema", *r.reason, *r.split)
	}
	for i, candidate := range schema.RemoteKills { // Replace
		if candidate.Split == *r.split && candidate.Reason == *r.reason {
			schema.RemoteKills[i] = *r.serializable()
			return nil
		}
	}
	schema.RemoteKills = append(schema.RemoteKills, *r.serializable()) // Add
	return nil
}
