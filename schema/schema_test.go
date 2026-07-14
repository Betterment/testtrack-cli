package schema

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/stretchr/testify/require"
)

func TestSortSchemaVersionsByHash(t *testing.T) {
	schema := &serializers.Schema{SchemaVersions: []string{
		"2020011712345",
		"2020011712346",
		"2020011712347",
		"2020011712348",
		"2020011712349",
	}}
	sortSchemaVersions(schema)

	expected := []string{
		"2020011712349", // 0dfad45993f8e2122515814d8e62f676fcb33841
		"2020011712347", // 13af4b988d91f57f71abeb0d657c24a18fba27d7
		"2020011712346", // 529bb729c31584a72d98669cf1daa0b9509379ce
		"2020011712348", // a467c4bdcdbdac7030e77951932029e03a8d0920
		"2020011712345", // bef98a4d8faa7aea1b347fb834a25c01be3c44d2
	}
	require.Equal(t, expected, schema.SchemaVersions, "versions should be ordered by sha1 of the version")
}

func TestReadRejectsNewerSerializerVersion(t *testing.T) {
	withSchemaFile(t, `
serializer_version: 99
schema_versions:
- "2020011712345"
`)

	_, err := Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "upgrade your testtrack CLI")
}

func TestReadRejectsOlderSerializerVersion(t *testing.T) {
	withSchemaFile(t, `
serializer_version: 1
schema_version: "2020011712345"
`)

	_, err := Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "testtrack schema upgrade")
}

func TestReadRejectsHybridSchemaFromPre2CLIRewrite(t *testing.T) {
	// A pre-2.0 CLI rewriting a v2 schema round-trips serializer_version: 2 but
	// writes the v1 shape: scalar schema_version, no schema_versions list.
	withSchemaFile(t, `
serializer_version: 2
schema_version: "2020011712345"
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`)

	_, err := Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "pre-2.0 testtrack CLI")
	require.Contains(t, err.Error(), "testtrack schema upgrade")
}

func TestReadRejectsHybridSchemaWithEmptyScalar(t *testing.T) {
	// 1.x write paths that never set the scalar (e.g. `sync`) emit
	// schema_version: "" — the v1 field has no omitempty. Detection must be by
	// key presence, not non-empty value, or this hybrid silently reads as a v2
	// schema with an empty applied-version list.
	withSchemaFile(t, `
serializer_version: 2
schema_version: ""
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`)

	_, err := Read()
	require.Error(t, err)
	require.Contains(t, err.Error(), "pre-2.0 testtrack CLI")
	require.Contains(t, err.Error(), "testtrack schema upgrade")
}

func TestReadRejectsEmptyVersionListWithMigrationsOnDisk(t *testing.T) {
	// A merge resolution that deletes the schema_versions block (or nulls the
	// legacy scalar, which evades the presence guard) leaves a v2 file with no
	// applied-version list. Reading it as-is would ratify the truncated list on
	// the next write.
	for name, contents := range map[string]string{
		"missing list": `
serializer_version: 2
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`,
		"nulled legacy scalar": `
serializer_version: 2
schema_version:
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`,
	} {
		t.Run(name, func(t *testing.T) {
			withSchemaFile(t, contents)
			require.NoError(t, os.MkdirAll("testtrack/migrate", 0755))
			writeSplitMigration(t, "2020011712345", "some_split")

			_, err := Read()
			require.Error(t, err)
			require.Contains(t, err.Error(), "no schema_versions but testtrack/migrate contains migrations")
			require.Contains(t, err.Error(), "testtrack schema upgrade")
		})
	}
}

func TestReadAcceptsEmptyVersionListWithoutMigrations(t *testing.T) {
	// A schema with no recorded versions is valid when there are no migrations
	// on disk (fresh projects, legacy repos with no migrate dir).
	withSchemaFile(t, `
serializer_version: 2
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`)

	_, err := Read()
	require.NoError(t, err)
}

func TestUpgradeRepairsHybridSchema(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.MkdirAll("testtrack/migrate", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "schema.yml"), []byte(`
serializer_version: 2
schema_version: "2020011712345"
splits:
- name: some_split
  weights:
    "false": 100
    "true": 0
`), 0644))
	writeSplitMigration(t, "2020011712345", "some_split")

	schema, err := Upgrade()
	require.NoError(t, err)
	require.Equal(t, []string{"2020011712345"}, schema.SchemaVersions)

	// The legacy scalar must not survive the rewrite, so the repaired file
	// passes Read's hybrid guard.
	contents, err := os.ReadFile(filepath.Join("testtrack", "schema.yml"))
	require.NoError(t, err)
	require.NotContains(t, string(contents), "schema_version:")

	reread, err := Read()
	require.NoError(t, err)
	require.Equal(t, []string{"2020011712345"}, reread.SchemaVersions)
}

func TestUpgradeConvertsInPlacePreservingBodyWithoutReplaying(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.MkdirAll("testtrack/migrate", 0755))

	// A v1 schema whose body references a split that has NO create migration on
	// disk — i.e. it could not be rebuilt by replaying (a common real-world case).
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "schema.yml"), []byte(`
serializer_version: 1
schema_version: "2020011712346"
splits:
- name: legacy_split_created_out_of_band
  weights:
    "false": 100
    "true": 0
`), 0644))

	// Migration files on disk decide that orphan split — replaying these from
	// scratch would fail, but upgrade doesn't replay.
	writeSplitMigration(t, "2020011712345", "some_other_split")
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "migrate", "2020011712346_create_split_decision_legacy_split_created_out_of_band.yml"),
		[]byte("serializer_version: 1\nsplit_decision:\n  split: legacy_split_created_out_of_band\n  variant: \"false\"\n"), 0644))

	schema, err := Upgrade()
	require.NoError(t, err)

	require.Equal(t, serializers.SerializerVersion, schema.SerializerVersion)
	// Version list rebuilt from the migration filenames.
	require.ElementsMatch(t, []string{"2020011712345", "2020011712346"}, schema.SchemaVersions)
	// Materialized body preserved as-is (not replayed/rebuilt).
	require.Len(t, schema.Splits, 1)
	require.Equal(t, "legacy_split_created_out_of_band", schema.Splits[0].Name)

	// And the file is now readable by the normal guarded Read().
	reread, err := Read()
	require.NoError(t, err)
	require.Equal(t, serializers.SerializerVersion, reread.SerializerVersion)
}

func TestUpgradeUsesFilenamesOnlyAndDedupes(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.MkdirAll("testtrack/migrate", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "schema.yml"), []byte(`
serializer_version: 1
schema_version: "2020011712347"
`), 0644))

	// Unparsable contents must not break upgrade — it only needs filenames.
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "migrate", "2020011712345_create_split_broken.yml"),
		[]byte("{{{ not yaml"), 0644))
	// Two files sharing a version (it happens in real repos) record it once.
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "migrate", "2020011712346_create_split_twin_a.yml"),
		[]byte("also garbage"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "migrate", "2020011712346_create_split_twin_b.yml"),
		[]byte("also garbage"), 0644))
	// Hidden files are skipped, matching the migration loader.
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "migrate", ".DS_Store"),
		[]byte{0}, 0644))

	schema, err := Upgrade()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"2020011712345", "2020011712346"}, schema.SchemaVersions)
}

func TestSortAlphabeticallyRemoteKillsIsDeterministic(t *testing.T) {
	// These three kills are mutually incomparable under the old buggy
	// comparator (Split < Split && Reason < Reason), which made the written
	// order depend on input order.
	kills := []serializers.RemoteKill{
		{Split: "a_split", Reason: "z_reason"},
		{Split: "b_split", Reason: "m_reason"},
		{Split: "c_split", Reason: "a_reason"},
	}
	forward := &serializers.Schema{RemoteKills: []serializers.RemoteKill{kills[0], kills[1], kills[2]}}
	reversed := &serializers.Schema{RemoteKills: []serializers.RemoteKill{kills[2], kills[1], kills[0]}}

	SortAlphabetically(forward)
	SortAlphabetically(reversed)

	require.Equal(t, forward.RemoteKills, reversed.RemoteKills,
		"remote kills must sort to the same order regardless of input order")
	require.Equal(t, []serializers.RemoteKill{kills[0], kills[1], kills[2]}, forward.RemoteKills)
}

func TestReadAcceptsCurrentSerializerVersion(t *testing.T) {
	withSchemaFile(t, `
serializer_version: 2
schema_versions:
- "2020011712345"
- "2020011712346"
`)

	schema, err := Read()
	require.NoError(t, err)
	require.Equal(t, []string{"2020011712345", "2020011712346"}, schema.SchemaVersions)
}

func TestGenerateRecordsAllMigrationVersions(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.MkdirAll("testtrack/migrate", 0755))

	versions := []string{"2020011712345", "2020011712346", "2020011712347"}
	for i, version := range versions {
		writeSplitMigration(t, version, []string{"alpha_experiment", "bravo_experiment", "charlie_experiment"}[i])
	}

	schema, err := Generate()
	require.NoError(t, err)

	require.ElementsMatch(t, versions, schema.SchemaVersions,
		"every migration version should be recorded in schema_versions")
}

// withSchemaFile writes a schema file into a temp testtrack dir and chdirs there.
func withSchemaFile(t *testing.T, contents string) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.MkdirAll("testtrack", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("testtrack", "schema.yml"), []byte(contents), 0644))
}

func writeSplitMigration(t *testing.T, version, name string) {
	t.Helper()
	contents := "serializer_version: " + strconv.Itoa(serializers.MigrationSerializerVersion) + "\n" +
		"split:\n" +
		"  name: " + name + "\n" +
		"  weights:\n" +
		"    control: 50\n" +
		"    treatment: 50\n"
	path := filepath.Join("testtrack", "migrate", version+"_"+name+".yml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0644))
}
