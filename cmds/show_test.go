package cmds

import (
	"net/http"
	"testing"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	// variant names are left-padded to the widest name, weights right-aligned,
	// and variants sorted: "false" before "true"
	expected := name + "\n  false  100%\n  true     0%"
	require.Equal(t, expected, output)
}

func TestSplitWeightsAlignsToWidestVariantAndWeight(t *testing.T) {
	name := "my_app.checkout_experiment"
	server := &fakeServer{registry: registryWith(name, map[string]int{
		"control":   5,
		"treatment": 95,
	})}

	output, err := splitWeights(server, name, false)
	require.NoError(t, err)

	// names left-aligned to "treatment" (9), weights right-aligned to width 2
	expected := name + "\n  control     5%\n  treatment  95%"
	require.Equal(t, expected, output)
}

func TestSplitWeightsJSON(t *testing.T) {
	name := "retail.cash_in_portfolios_q2_2026_enabled"
	weights := map[string]int{"false": 100, "true": 0}
	server := &fakeServer{registry: registryWith(name, weights)}

	output, err := splitWeights(server, name, true)
	require.NoError(t, err)
	require.JSONEq(t, `{"false":100,"true":0}`, output)
}

func TestShowRequiresServerURL(t *testing.T) {
	t.Setenv("TESTTRACK_CLI_URL", "")

	err := Show("any.split", false)
	require.Error(t, err)
}

func TestSplitWeightsNotFound(t *testing.T) {
	server := &fakeServer{registry: registryWith("some.other.split", map[string]int{"true": 100})}

	_, err := splitWeights(server, "no.such.split", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
	require.Contains(t, err.Error(), "testtrack sync")
}
