package schemaloaders

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/stretchr/testify/require"
)

const unrecordedWarning = "migrations on disk not recorded in the schema file"

// fakeServer records the paths it's asked to POST and succeeds unless
// failAfter is set, in which case Posts beyond that count return a 500.
type fakeServer struct {
	postedPaths []string
	failAfter   int
}

func (f *fakeServer) Get(string, interface{}) error { return nil }

func (f *fakeServer) Delete(string) error { return nil }

func (f *fakeServer) Post(path string, _ interface{}) (*http.Response, error) {
	f.postedPaths = append(f.postedPaths, path)
	if f.failAfter > 0 && len(f.postedPaths) > f.failAfter {
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
}

func TestLoadSyncsOnlyRecordedVersionsAndWarnsOnce(t *testing.T) {
	server := &fakeServer{}
	// Two versions are recorded in the schema; two migration files on disk are not.
	schema := &serializers.Schema{SchemaVersions: []string{"2020011712345", "2020011712346"}}
	repo := migrations.Repository{
		"2020011712345": newSplitMigration(t, "2020011712345"),
		"2020011712346": newSplitMigration(t, "2020011712346"),
		"2020011712347": newSplitMigration(t, "2020011712347"),
		"2020011712348": newSplitMigration(t, "2020011712348"),
	}
	loader := &SchemaLoader{server: server, schema: schema, migrationRepo: &repo}

	out := captureStdout(t, func() {
		require.NoError(t, loader.Load())
	})

	require.Equal(t, 1, strings.Count(out, unrecordedWarning),
		"warning should print exactly once no matter how many unrecorded migrations exist")
	// The warning enumerates exactly the unrecorded versions so the user can
	// inspect them, and explains both repair paths.
	require.Contains(t, out, "2020011712347")
	require.Contains(t, out, "2020011712348")
	require.NotContains(t, out, "  2020011712345\n")
	require.Contains(t, out, "testtrack schema upgrade")
	require.Contains(t, out, "WITHOUT applying them")
	// Only the two recorded versions get marked applied on the server; the
	// schema here has no splits/etc., so these are the only posts.
	require.Equal(t, []string{"api/v2/migrations", "api/v2/migrations"}, server.postedPaths)
}

func TestLoadWarningSurvivesSyncVersionFailure(t *testing.T) {
	// The first SyncVersion succeeds, the second gets a 500 — the warning must
	// already be on stdout, not swallowed by the error return.
	server := &fakeServer{failAfter: 1}
	schema := &serializers.Schema{SchemaVersions: []string{"2020011712345", "2020011712346"}}
	repo := migrations.Repository{
		"2020011712345": newSplitMigration(t, "2020011712345"),
		"2020011712346": newSplitMigration(t, "2020011712346"),
		"2020011712347": newSplitMigration(t, "2020011712347"),
	}
	loader := &SchemaLoader{server: server, schema: schema, migrationRepo: &repo}

	out := captureStdout(t, func() {
		require.Error(t, loader.Load())
	})

	require.Equal(t, 1, strings.Count(out, unrecordedWarning),
		"warning must print even when a later SyncVersion fails")
	require.Contains(t, out, "2020011712347")
}

func TestLoadDoesNotWarnWhenEveryMigrationIsRecorded(t *testing.T) {
	server := &fakeServer{}
	schema := &serializers.Schema{SchemaVersions: []string{"2020011712345", "2020011712346"}}
	repo := migrations.Repository{
		"2020011712345": newSplitMigration(t, "2020011712345"),
		"2020011712346": newSplitMigration(t, "2020011712346"),
	}
	loader := &SchemaLoader{server: server, schema: schema, migrationRepo: &repo}

	out := captureStdout(t, func() {
		require.NoError(t, loader.Load())
	})

	require.NotContains(t, out, unrecordedWarning)
	require.Equal(t, []string{"api/v2/migrations", "api/v2/migrations"}, server.postedPaths)
}

func newSplitMigration(t *testing.T, version string) migrations.IMigration {
	t.Helper()
	migration, err := splits.FromFile(&version, &serializers.SplitYAML{
		Name:    "split_" + version,
		Weights: map[string]int{"control": 50, "treatment": 50},
	})
	require.NoError(t, err)
	return migration
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		os.Stdout = original
		r.Close()
	}()

	fn()
	require.NoError(t, w.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}
