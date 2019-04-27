package schema

import (
	"io/ioutil"
	"os"
	"sort"

	"github.com/Betterment/testtrack-cli/migrationrepositories"
	"github.com/Betterment/testtrack-cli/serializers"
	"gopkg.in/yaml.v2"
)

// Read a schema from disk or generate one
func Read() (*serializers.Schema, error) {
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

// Generate a schema from migrations on the filesystem and write it to disk
func Generate() (*serializers.Schema, error) {
	schema := &serializers.Schema{SerializerVersion: serializers.SerializerVersion}
	err := applyAllMigrationsToSchema(schema)
	if err != nil {
		return nil, err
	}
	err = Write(schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// Write a schema to disk after alpha-sorting its resources
func Write(schema *serializers.Schema) error {
	SortAlphabetically(schema)
	out, err := yaml.Marshal(schema)

	err = ioutil.WriteFile("testtrack/schema.yml", out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func applyAllMigrationsToSchema(schema *serializers.Schema) error {
	migrationRepo, err := migrationrepositories.Load()
	if err != nil {
		return err
	}

	versions := migrationRepo.SortedVersions()

	for _, version := range versions {
		err = migrationRepo[version].ApplyToSchema(schema)
		if err != nil {
			return err
		}
	}
	if len(versions) != 0 {
		schema.SchemaVersion = versions[len(versions)-1]
	}
	return nil
}

// SortAlphabetically sorts the schema's resource slices by their natural keys
func SortAlphabetically(schema *serializers.Schema) {
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
