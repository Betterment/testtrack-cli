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

// fakeServer records the paths it's asked to POST and always succeeds.
type fakeServer struct {
	postedPaths []string
}

func (f *fakeServer) Get(string, interface{}) error { return nil }

func (f *fakeServer) Delete(string) error { return nil }

func (f *fakeServer) Post(path string, _ interface{}) (*http.Response, error) {
	f.postedPaths = append(f.postedPaths, path)
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
	// Only the two recorded versions get marked applied on the server; the
	// schema here has no splits/etc., so these are the only posts.
	require.Equal(t, []string{"api/v2/migrations", "api/v2/migrations"}, server.postedPaths)
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
