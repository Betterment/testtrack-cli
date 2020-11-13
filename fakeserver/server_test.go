package fakeserver

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Betterment/testtrack-cli/fakeassignments"

	"encoding/json"

	"github.com/stretchr/testify/require"
)

var testSchema = `
serializer_version: 1
schema_version: "2020011774023"
splits:
- name: test.test_experiment
  weights:
    control: 60
    treatment: 40
- name: test.test2_experiment
  weights:
    control: 60
    treatment: 40
`

func TestMain(m *testing.M) {
	current, exists := os.LookupEnv("TESTTRACK_FAKE_SERVER_CONFIG_DIR")

	dir, err := ioutil.TempDir("", "testtrack-cli")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	schemasDir := filepath.Join(dir, "schemas")
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		log.Fatal(err)
	}

	schemaContent := []byte(testSchema)
	if err := ioutil.WriteFile(filepath.Join(schemasDir, "test.yml"), schemaContent, 0644); err != nil {
		log.Fatal(err)
	}

	os.Setenv("TESTTRACK_FAKE_SERVER_CONFIG_DIR", dir)
	exitCode := m.Run()
	if exists {
		os.Setenv("TESTTRACK_FAKE_SERVER_CONFIG_DIR", current)
	}
	os.Exit(exitCode)
}

func TestSplitRegistry(t *testing.T) {
	t.Run("it loads split registry v2", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v2/split_registry", nil))

		require.Equal(t, http.StatusOK, w.Code)

		registry := v2SplitRegistry{}
		err := json.Unmarshal(w.Body.Bytes(), &registry)
		require.Nil(t, err)

		require.Equal(t, 1, registry.ExperienceSamplingWeight)
		require.Equal(t, 60, registry.Splits["test.test_experiment"].Weights["control"])
		require.Equal(t, 40, registry.Splits["test.test_experiment"].Weights["treatment"])
		require.Equal(t, false, registry.Splits["test.test_experiment"].FeatureGate)
	})

	t.Run("it loads split registry v3", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v3/builds/2020-01-02T03:04:05/split_registry", nil))

		require.Equal(t, http.StatusOK, w.Code)

		registry := v2SplitRegistry{}
		err := json.Unmarshal(w.Body.Bytes(), &registry)
		require.Nil(t, err)

		require.Equal(t, 1, registry.ExperienceSamplingWeight)
		require.Equal(t, 60, registry.Splits["test.test_experiment"].Weights["control"])
		require.Equal(t, 40, registry.Splits["test.test_experiment"].Weights["treatment"])
		require.Equal(t, false, registry.Splits["test.test_experiment"].FeatureGate)
	})
}

func TestCors(t *testing.T) {
	os.Setenv("TESTTRACK_ALLOWED_ORIGINS", "allowed.com")

	t.Run("it fails cors with an unallowed origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		request := httptest.NewRequest("GET", "/api/v2/split_registry", nil)
		request.Header.Add("Origin", "http://www.denied.com")

		h.ServeHTTP(w, request)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "", w.HeaderMap.Get("Access-Control-Allow-Origin"))
	})

	t.Run("it passes cors with an allowed origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		request := httptest.NewRequest("GET", "/api/v2/split_registry", nil)
		request.Header.Add("Origin", "http://www.allowed.com")

		h.ServeHTTP(w, request)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "http://www.allowed.com", w.HeaderMap.Get("Access-Control-Allow-Origin"))
	})

	os.Unsetenv("TESTTRACK_ALLOWED_ORIGINS")
}

func TestPersistAssignment(t *testing.T) {
	os.Remove("testdata/assignments.yml")

	t.Run("it persists assignments to yaml", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		data := url.Values{}
		data.Set("split_name", "test.test_experiment")
		data.Set("variant", "control")

		request := httptest.NewRequest("POST", "/api/v1/assignment_override", strings.NewReader(data.Encode()))
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		h.ServeHTTP(w, request)

		require.Equal(t, http.StatusNoContent, w.Code)

		assignments, err := fakeassignments.Read()
		require.Nil(t, err)
		require.Equal(t, "control", (*assignments)["test.test_experiment"])
	})
}

func TestPersistAssignmentV2(t *testing.T) {
	os.Remove("testdata/assignments.yml")

	t.Run("it persists assignments to yaml", func(t *testing.T) {
		w := httptest.NewRecorder()
		h := createHandler()

		overrides := v2AssignmentOverrideRequestBody{
			Assignments: []v1Assignment{
				v1Assignment{
					SplitName: "test.test_experiment",
					Variant:   "control",
				},
				v1Assignment{
					SplitName: "test.test2_experiment",
					Variant:   "treatment",
				},
			},
		}
		data, err := json.Marshal(overrides)
		require.Nil(t, err)

		request := httptest.NewRequest("POST", "/api/v2/visitors/1/assignment_overrides", bytes.NewReader(data))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Content-Length", strconv.Itoa(len(data)))

		h.ServeHTTP(w, request)

		require.Equal(t, http.StatusNoContent, w.Code)

		assignments, err := fakeassignments.Read()
		require.Nil(t, err)
		require.Equal(t, "control", (*assignments)["test.test_experiment"])
		require.Equal(t, "treatment", (*assignments)["test.test2_experiment"])
	})
}
