package schema

import (
	"io/ioutil"
	"os"
	"sort"

	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/serializers"
	"gopkg.in/yaml.v2"
)

// Load a schema from disk or generate one
func Load() (*serializers.Schema, error) {
	if _, err := os.Stat("testtrack/schema.yml"); os.IsNotExist(err) {
		return Generate()
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

// Generate a schema from migrations on the filesystem and dump it to disk
func Generate() (*serializers.Schema, error) {
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

// Dump a schema to disk after alpha-sorting its resources
func Dump(schema *serializers.Schema) error {
	sortAlphabetically(schema)
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

func sortAlphabetically(schema *serializers.Schema) {
	sort.Slice(schema.RemoteKills, func(i, j int) bool {
		return schema.RemoteKills[i].Split < schema.RemoteKills[j].Split &&
			schema.RemoteKills[i].Reason < schema.RemoteKills[j].Reason
	})
	sort.Slice(schema.FeatureCompletions, func(i, j int) bool {
		return schema.FeatureCompletions[i].FeatureGate < schema.FeatureCompletions[j].FeatureGate
	})
	sort.Slice(schema.Splits, func(i, j int) bool {
		return schema.Splits[i].Name < schema.Splits[j].Name
	})
	sort.Slice(schema.IdentifierTypes, func(i, j int) bool {
		return schema.IdentifierTypes[i].Name < schema.IdentifierTypes[j].Name
	})
}
