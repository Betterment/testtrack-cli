package schema

import (
	"testing"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/stretchr/testify/require"
)

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
