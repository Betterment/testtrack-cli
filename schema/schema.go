package schema

import (
	"io/ioutil"
	"os"

	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/serializers"
	"gopkg.in/yaml.v2"
)

// Load loads from disk or instantiates an empty Schema struct
func Load() (*serializers.Schema, error) {
	if _, err := os.Stat("testtrack/schema.yml"); os.IsNotExist(err) {
		schema := &serializers.Schema{SerializerVersion: serializers.SerializerVersion}
		err := applyAllMigrationsToSchema(schema)
		if err != nil {
			return nil, err
		}
		err = Dump(schema)
		if err != nil {
			return nil, err
		}
		return schema, nil
	}
	schemaBytes, err := ioutil.ReadFile("testtrack/schema.yml")
	if err != nil {
		return nil, err
	}
	var schema serializers.Schema
	err = yaml.Unmarshal(schemaBytes, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Dump dumps a schema to disk
func Dump(schema *serializers.Schema) error {
	out, err := yaml.Marshal(schema)

	err = ioutil.WriteFile("testtrack/schema.yml", out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func applyAllMigrationsToSchema(schema *serializers.Schema) error {
	migrationsByVersion, err := migrationloaders.Load()
	if err != nil {
		return err
	}

	versions := migrationloaders.GetSortedVersions(migrationsByVersion)

	for _, version := range versions {
		err = migrationsByVersion[version].ApplyToSchema(schema)
		if err != nil {
			return err
		}
	}
	return nil
}
