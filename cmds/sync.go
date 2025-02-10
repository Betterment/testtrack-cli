package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/spf13/cobra"
)

var syncDoc = `
Sync the local schema TestTrack assignments with the remote production TestTrack assignments.
`

func init() {
	rootCmd.AddCommand(syncCommand)
}

var syncCommand = &cobra.Command{
	Use:   "sync",
	Short: "Sync TestTrack assignments with production",
	Long:  syncDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return Sync()
	},
}

// Sync synchronizes the local schema TestTrack assignments with the remote production TestTrack assignments.
func Sync() error {
	server, err := servers.New()
	if err != nil {
		return err
	}

	var splitRegistry serializers.RemoteRegistry
	err = server.Get("api/v2/split_registry.json", &splitRegistry)
	if err != nil {
		return err
	}

	localSchema, err := schema.Read()
	if err != nil {
		return err
	}

	for ind, localSplit := range localSchema.Splits {
		remoteSplit, exists := splitRegistry.Splits[localSplit.Name]
		if exists {
			remoteWeights := splits.Weights(remoteSplit.Weights)
			localSchema.Splits[ind].Weights = remoteWeights.ToYAML()
		}
	}

	if err := schema.Write(localSchema); err != nil {
		return err
	}

	return nil
}
