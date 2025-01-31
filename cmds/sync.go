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
	RunE: func(cmd *cobra.Command, args []string) error {
		return Sync()
	},
}

// Sync synchronizes the local schema TestTrack assignments with the remote production TestTrack assignments.
func Sync() error {
	server, err := servers.New()
	if err != nil {
		return err
	}

	var splitRegistry serializers.SplitRegistry
	server.Get("api/v2/split_registry.json", &splitRegistry)

	localSchema, err := schema.Read()
	if err != nil {
		return err
	}

	for ind, localSplit := range localSchema.Splits {
		remoteSplit, exists := splitRegistry.Splits[localSplit.Name]
		if exists {
			weights := splits.Weights(remoteSplit.Weights)
			localSchema.Splits[ind].Weights = weights.ToYAML()
		}
	}

	schema.Write(localSchema)

	return nil
}
