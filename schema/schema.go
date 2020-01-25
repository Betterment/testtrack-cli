package schema

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/Betterment/testtrack-cli/migrationloaders"
	"github.com/Betterment/testtrack-cli/paths"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/splits"
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
	err := mergeLegacySchema(schema)
	if err != nil {
		return nil, err
	}
	err = applyAllMigrationsToSchema(schema)
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

// Link a schema to the user's home dir
func Link(force bool) error {
	if _, err := os.Stat("testtrack/schema.yml"); os.IsNotExist(err) {
		return errors.New("testtrack/schema.yml does not exist. Are you in your app root dir? If so, call testtrack init_project first")
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	dirname := path.Base(dir)
	configDir, err := paths.ConfigDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(*configDir+"/schemas", 0755)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/schemas/%s.yml", *configDir, dirname)
	if force {
		os.Remove(path) // If this fails it might just not exist, we'll error on the next line if something else is up
	}
	return os.Symlink(dir+"/testtrack/schema.yml", path)
}

// ReadMerged merges schemas linked at ~/testtrack/schemas into a single virtual schema
func ReadMerged() (*serializers.Schema, error) {
	configDir, err := paths.ConfigDir()
	if err != nil {
		return nil, err
	}
	paths, err := filepath.Glob(*configDir + "/schemas/*.yml")
	if err != nil {
		return nil, err
	}
	var mergedSchema serializers.Schema
	for _, path := range paths {
		// Deref symlink
		fi, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			path, err = os.Readlink(path)
			if err != nil {
				continue // It's OK if this symlink isn't traversable (e.g. app was uninstalled), we'll just skip it.
			}
		}
		// Read file
		schemaBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var schema serializers.Schema
		err = yaml.Unmarshal(schemaBytes, &schema)
		if err != nil {
			return nil, err
		}
		// Merge into master schema
		for _, split := range schema.Splits {
			mergedSchema.Splits = append(mergedSchema.Splits, split)
		}
		for _, featureCompletion := range schema.FeatureCompletions {
			mergedSchema.FeatureCompletions = append(mergedSchema.FeatureCompletions, featureCompletion)
		}
		for _, remoteKill := range schema.RemoteKills {
			mergedSchema.RemoteKills = append(mergedSchema.RemoteKills, remoteKill)
		}
		for _, identifierType := range schema.IdentifierTypes {
			mergedSchema.IdentifierTypes = append(mergedSchema.IdentifierTypes, identifierType)
		}
	}
	return &mergedSchema, nil
}

func mergeLegacySchema(schema *serializers.Schema) error {
	if _, err := os.Stat("db/test_track_schema.yml"); os.IsNotExist(err) {
		return nil
	}
	legacySchemaBytes, err := ioutil.ReadFile("db/test_track_schema.yml")
	if err != nil {
		return err
	}
	var legacySchema serializers.LegacySchema
	err = yaml.Unmarshal(legacySchemaBytes, &legacySchema)
	if err != nil {
		return err
	}
	for _, name := range legacySchema.IdentifierTypes {
		schema.IdentifierTypes = append(schema.IdentifierTypes, serializers.IdentifierType{
			Name: name,
		})
	}
	for _, mapSlice := range legacySchema.Splits {
		name, ok := mapSlice.Key.(string)
		if !ok {
			return fmt.Errorf("expected split name, got %v", mapSlice.Key)
		}
		weightsYAML, ok := mapSlice.Value.(yaml.MapSlice)
		if !ok {
			return fmt.Errorf("expected weights, got %v", mapSlice.Value)
		}
		weights, err := splits.WeightsFromYAML(weightsYAML)
		if err != nil {
			return err
		}

		schema.Splits = append(schema.Splits, serializers.SchemaSplit{
			Name:    name,
			Weights: weights.ToYAML(),
			Decided: false,
		})
	}
	return nil
}

func applyAllMigrationsToSchema(schema *serializers.Schema) error {
	migrationRepo, err := migrationloaders.Load()
	if err != nil {
		return err
	}

	versions := migrationRepo.SortedVersions()

	for _, version := range versions {
		err = migrationRepo[version].ApplyToSchema(schema, migrationRepo, false)
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
