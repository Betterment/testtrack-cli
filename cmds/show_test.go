package cmds

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Betterment/testtrack-cli/serializers"
)

// fakeServer is a test double for servers.IServer that returns a canned registry.
type fakeServer struct {
	registry serializers.RemoteRegistry
}

func (f *fakeServer) Get(_ string, v interface{}) error {
	*v.(*serializers.RemoteRegistry) = f.registry
	return nil
}

func (f *fakeServer) Post(_ string, _ interface{}) (*http.Response, error) {
	return nil, nil
}

func (f *fakeServer) Delete(_ string) error {
	return nil
}

func registryWith(name string, weights map[string]int) serializers.RemoteRegistry {
	return serializers.RemoteRegistry{
		Splits: map[string]serializers.RemoteRegistrySplit{
			name: {Weights: weights},
		},
	}
}

func TestSplitWeightsHumanReadable(t *testing.T) {
	name := "retail.cash_in_portfolios_q2_2026_enabled"
	server := &fakeServer{registry: registryWith(name, map[string]int{"false": 100, "true": 0})}

	output, err := splitWeights(server, name, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// variant names are left-padded to the widest name, weights right-aligned
	expected := name + "\n  false  100%\n  true     0%"
	if output != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, output)
	}

	// variants must be sorted: "false" before "true"
	if strings.Index(output, "false") > strings.Index(output, "true") {
		t.Errorf("expected variants sorted by name, got:\n%s", output)
	}
}

func TestSplitWeightsJSON(t *testing.T) {
	name := "retail.cash_in_portfolios_q2_2026_enabled"
	weights := map[string]int{"false": 100, "true": 0}
	server := &fakeServer{registry: registryWith(name, weights)}

	output, err := splitWeights(server, name, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var roundTripped map[string]int
	if err := json.Unmarshal([]byte(output), &roundTripped); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if roundTripped["false"] != 100 || roundTripped["true"] != 0 {
		t.Errorf("JSON did not round-trip to the weights map, got: %s", output)
	}
}

func TestShowRequiresServerURL(t *testing.T) {
	t.Setenv("TESTTRACK_CLI_URL", "")

	if err := Show("any.split", false); err == nil {
		t.Fatal("expected an error when TESTTRACK_CLI_URL is unset, got nil")
	}
}

func TestSplitWeightsNotFound(t *testing.T) {
	server := &fakeServer{registry: registryWith("some.other.split", map[string]int{"true": 100})}

	_, err := splitWeights(server, "no.such.split", false)
	if err == nil {
		t.Fatal("expected an error for a missing split, got nil")
	}
	if !strings.Contains(err.Error(), "not found") || !strings.Contains(err.Error(), "testtrack sync") {
		t.Errorf("expected not-found error with hint, got: %v", err)
	}
}
