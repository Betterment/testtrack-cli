package identifiertypes

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
)

// IdentifierType represents a feature we're marking (un)completed
type IdentifierType struct {
	migrationVersion *string
	name             *string
}

// New returns a migration object
func New(name *string) (migrations.IMigration, error) {
	migrationVersion, err := migrations.GenerateMigrationVersion()
	if err != nil {
		return nil, err
	}

	return &IdentifierType{
		migrationVersion: migrationVersion,
		name:             name,
	}, nil
}

// FromFile reifies a migration from the yaml serializable representation
func FromFile(migrationVersion *string, serializable *serializers.IdentifierType) migrations.IMigration {
	return &IdentifierType{
		migrationVersion: migrationVersion,
		name:             &serializable.Name,
	}
}

// Validate that the migration may be persisted
func (i *IdentifierType) Validate() error {
	err := validations.SnakeCaseParam("name", i.name)
	if err != nil {
		return err
	}

	return nil
}

// Filename generates a filename for this migration
func (i *IdentifierType) Filename() *string {
	filename := fmt.Sprintf("%s_create_identifier_type_%s.yml", *i.migrationVersion, *i.name)
	return &filename
}

// File returns a serializable MigrationFile for this migration
func (i *IdentifierType) File() *serializers.MigrationFile {
	return &serializers.MigrationFile{
		SerializerVersion: serializers.SerializerVersion,
		IdentifierType:    i.serializable(),
	}
}

// SyncPath returns the server path to post the migration to
func (i *IdentifierType) SyncPath() string {
	return "api/v1/identifier_type"
}

// Serializable returns a JSON serializable representation
func (i *IdentifierType) Serializable() interface{} {
	return i.serializable()
}

func (i *IdentifierType) serializable() *serializers.IdentifierType {
	return &serializers.IdentifierType{
		Name: *i.name,
	}
}

// MigrationVersion returns the migration version
func (i *IdentifierType) MigrationVersion() *string {
	return i.migrationVersion
}

// SameResourceAs returns whether the migrations refer to the same TestTrack resource
func (i *IdentifierType) SameResourceAs(other migrations.IMigration) bool {
	if otherI, ok := other.(*IdentifierType); ok {
		return *otherI.name == *i.name
	}
	return false
}

// Inverse returns a logical inverse operation if possible
func (i *IdentifierType) Inverse() (migrations.IMigration, error) {
	return nil, fmt.Errorf("can't invert identifier_type creation %s %s", *i.migrationVersion, *i.name)
}

// ApplyToSchema applies a migrations changes to in-memory schema representation
func (i *IdentifierType) ApplyToSchema(schema *serializers.Schema, _ migrations.Repository) error {
	for _, candidate := range schema.IdentifierTypes {
		if candidate.Name == *i.name {
			return fmt.Errorf("identifier_type %s already exists", *i.name)
		}
	}
	schema.IdentifierTypes = append(schema.IdentifierTypes, *i.serializable())
	return nil
}
