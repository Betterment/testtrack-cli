package schema

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
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

// Finds the path to the schema file (preferring JSON), or returns testtrack/schema.json
func findSchemaPath() (string, bool) {
	if _, err := os.Stat("testtrack/schema.json"); err == nil {
		return "testtrack/schema.json", true
	}
	if _, err := os.Stat("testtrack/schema.yml"); err == nil {
		return "testtrack/schema.yml", true
	}
	return "testtrack/schema.json", false
}

// readSchemaFile locates and unmarshals the schema file, rejecting files
// written by a newer CLI. exists is false (with a nil schema and nil error)
// when no schema file is present; callers decide how to handle that.
func readSchemaFile() (schema *serializers.Schema, schemaPath string, exists bool, err error) {
	schemaPath, exists = findSchemaPath()
	if !exists {
		return nil, schemaPath, false, nil
	}
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, schemaPath, true, err
	}
	var s serializers.Schema
	err = yaml.Unmarshal(schemaBytes, &s)
	if err != nil {
		return nil, schemaPath, true, err
	}
	if s.SerializerVersion > serializers.SerializerVersion {
		return nil, schemaPath, true, fmt.Errorf(
			"%s was written by a newer testtrack CLI (serializer_version %d, this CLI supports %d). Please upgrade your testtrack CLI",
			schemaPath, s.SerializerVersion, serializers.SerializerVersion,
		)
	}
	return &s, schemaPath, true, nil
}

// Read a schema from disk or generate one
func Read() (*serializers.Schema, error) {
	schema, schemaPath, exists, err := readSchemaFile()
	if err != nil {
		return nil, err
	}
	if !exists {
		return Generate()
	}
	if schema.SerializerVersion < serializers.SerializerVersion {
		// An older file predates a format change and may be missing data that
		// can't be reconstructed by re-reading it (e.g. the v1 scalar
		// schema_version carried no per-migration list). Refuse to read it
		// rather than silently round-trip a lossy upgrade; `schema upgrade`
		// converts it in place, preserving the materialized state.
		return nil, fmt.Errorf(
			"%s uses an older schema format (serializer_version %d, this CLI writes %d). Run `testtrack schema upgrade` to upgrade it",
			schemaPath, schema.SerializerVersion, serializers.SerializerVersion,
		)
	}
	if schema.LegacySchemaVersion != nil {
		// A pre-2.0 CLI rewriting a v2 schema produces a hybrid: it round-trips
		// serializer_version: 2 but writes the v1 shape (scalar schema_version,
		// no schema_versions list), so the version check above can't catch it.
		// Reading it as-is would silently continue with an empty applied-version
		// list. Key presence, not value, is the tell: 1.x write paths that never
		// set the scalar (e.g. `sync`) emit schema_version: "".
		return nil, fmt.Errorf(
			"%s has serializer_version %d but contains the legacy schema_version field - it was likely rewritten by a pre-2.0 testtrack CLI. Run `testtrack schema upgrade` to repair it, and make sure no older CLI touches it again",
			schemaPath, schema.SerializerVersion,
		)
	}
	if len(schema.SchemaVersions) == 0 {
		// No machine write produces an empty version list alongside migrations
		// on disk — that state means the schema_versions block was lost to a
		// hand-edit or merge resolution (or the legacy scalar was nulled out,
		// which also evades the presence check above). Reading it as-is would
		// ratify the truncated list on the next write. A missing or unreadable
		// migrate dir counts as no migrations.
		if filenames, err := migrationloaders.Filenames(); err == nil && len(filenames) > 0 {
			return nil, fmt.Errorf(
				"%s has no schema_versions but testtrack/migrate contains migrations - the applied-version list was likely lost in a merge. Run `testtrack schema upgrade` to rebuild it",
				schemaPath,
			)
		}
	}
	return schema, nil
}

// Upgrade converts an existing schema file to the current serializer format in
// place. Unlike Generate it does not replay migrations, so it preserves the
// already-materialized state (splits, decisions, retirements, etc.) and works
// even on schemas that can't be rebuilt from scratch — e.g. apps whose
// testtrack/migrate predates some splits or references ones created out of
// band. The only thing it rebuilds is the applied-version list, which it reads
// from the migration filenames on disk.
func Upgrade() (*serializers.Schema, error) {
	schema, _, exists, err := readSchemaFile()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("no testtrack schema file to upgrade. Run testtrack schema generate to create one")
	}

	versions, err := migrationloaders.Versions()
	if err != nil {
		if os.IsNotExist(err) {
			// Legacy repos can have a schema with no testtrack/migrate dir at
			// all (git doesn't track empty dirs); there are no versions to
			// record, which is a valid v2 state.
			versions = nil
		} else {
			return nil, err
		}
	}
	schema.SchemaVersions = versions
	schema.SerializerVersion = serializers.SerializerVersion
	schema.LegacySchemaVersion = nil

	err = Write(schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
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

// Write a schema to disk after sorting its resources into a stable order
func Write(schema *serializers.Schema) error {
	SortAlphabetically(schema)
	sortSchemaVersions(schema)

	schemaPath, _ := findSchemaPath()

	var out []byte
	var err error
	if filepath.Ext(schemaPath) == ".yml" {
		out, err = yaml.Marshal(schema)
	} else {
		out, err = json.MarshalIndent(schema, "", "  ")
	}
	if err != nil {
		return err
	}

	err = os.WriteFile(schemaPath, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Link a schema to the user's home dir
func Link(force bool) error {
	schemaPath, exists := findSchemaPath()
	if !exists {
		return errors.New("testtrack/schema.{json,yml} does not exist. Are you in your app root dir? If so, call testtrack init_project first")
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	dirname := path.Base(dir)
	configDir, err := paths.FakeServerConfigDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(*configDir+"/schemas", 0755)
	if err != nil {
		return err
	}
	ext := filepath.Ext(schemaPath)
	path := fmt.Sprintf("%s/schemas/%s%s", *configDir, dirname, ext)
	if force {
		os.Remove(path) // If this fails it might just not exist, we'll error on the next line if something else is up
	}
	return os.Symlink(dir+"/"+schemaPath, path)
}

// ReadMerged merges schemas linked at ~/testtrack/schemas into a single virtual schema.
//
// It deliberately bypasses Read's serializer-version and hybrid guards: the
// linked schemas belong to *other* apps that upgrade on their own schedule,
// and hard-failing on a neighbor's stale schema would break assign/fakeserver
// for every app on the machine. This tolerance is safe because the body
// fields consumed here (splits, identifier_types, remote_kills,
// feature_completions) have never changed shape across serializer versions —
// a future version that reshapes them must add per-file version handling
// here.
func ReadMerged() (*serializers.Schema, error) {
	configDir, err := paths.FakeServerConfigDir()
	if err != nil {
		return nil, err
	}
	paths, err := filepath.Glob(*configDir + "/schemas/*.*")
	if err != nil {
		return nil, err
	}
	var mergedSchema serializers.Schema
	for _, path := range paths {
		schemaBytes, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // It's OK if this file doesn't exist (e.g. broken symlink, app was uninstalled), we'll just skip it.
			}
			return nil, err
		}

		var schema serializers.Schema
		err = yaml.Unmarshal(schemaBytes, &schema)
		if err != nil {
			return nil, err
		}
		// Merge into master schema
		mergedSchema.Splits = append(mergedSchema.Splits, schema.Splits...)
		mergedSchema.FeatureCompletions = append(mergedSchema.FeatureCompletions, schema.FeatureCompletions...)
		mergedSchema.RemoteKills = append(mergedSchema.RemoteKills, schema.RemoteKills...)
		mergedSchema.IdentifierTypes = append(mergedSchema.IdentifierTypes, schema.IdentifierTypes...)
	}
	return &mergedSchema, nil
}

func mergeLegacySchema(schema *serializers.Schema) error {
	if _, err := os.Stat("db/test_track_schema.yml"); os.IsNotExist(err) {
		return nil
	}
	legacySchemaBytes, err := os.ReadFile("db/test_track_schema.yml")
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
		weightsYAML, ok := mapSlice.Value.(map[string]int)
		if !ok {
			return fmt.Errorf("expected weights, got %v", mapSlice.Value)
		}
		weights, err := splits.NewWeights(weightsYAML)
		if err != nil {
			return err
		}

		schema.Splits = append(schema.Splits, serializers.SchemaSplit{
			Name:    name,
			Weights: *weights,
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
	schema.SchemaVersions = versions
	return nil
}

// sortSchemaVersions orders the applied-version list by the SHA-1 of each
// version. The order is otherwise meaningless; hashing scatters newly added
// versions through the list instead of clustering them by timestamp, so two
// branches that each append a migration rarely touch the same lines. Hashes
// are precomputed so each version is hashed once rather than on every
// comparison.
func sortSchemaVersions(schema *serializers.Schema) {
	hashes := make(map[string][sha1.Size]byte, len(schema.SchemaVersions))
	for _, version := range schema.SchemaVersions {
		hashes[version] = sha1.Sum([]byte(version))
	}
	sort.Slice(schema.SchemaVersions, func(i, j int) bool {
		a := hashes[schema.SchemaVersions[i]]
		b := hashes[schema.SchemaVersions[j]]
		return bytes.Compare(a[:], b[:]) < 0
	})
}

// SortAlphabetically sorts the schema's resource slices by their natural keys
func SortAlphabetically(schema *serializers.Schema) {
	sort.Slice(schema.RemoteKills, func(i, j int) bool {
		if schema.RemoteKills[i].Split != schema.RemoteKills[j].Split {
			return schema.RemoteKills[i].Split < schema.RemoteKills[j].Split
		}
		return schema.RemoteKills[i].Reason < schema.RemoteKills[j].Reason
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
