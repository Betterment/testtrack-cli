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

func writeSplitMigration(t *testing.T, version, name string) {
	t.Helper()
	contents := "serializer_version: " + strconv.Itoa(serializers.SerializerVersion) + "\n" +
		"split:\n" +
		"  name: " + name + "\n" +
		"  weights:\n" +
		"    control: 50\n" +
		"    treatment: 50\n"
	path := filepath.Join("testtrack", "migrate", version+"_"+name+".yml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0644))
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
