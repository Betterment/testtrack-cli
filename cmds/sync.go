package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

	var jsonData serializers.SplitRegistryJSON
	server.Get("api/v2/split_registry.json", &jsonData)

	remoteSplits := jsonData.Splits

	localSchema, err := schema.Read()
	if err != nil {
		return err
	}

	for remoteSplitName, remoteSplit := range remoteSplits {
		for ind, localSplit := range localSchema.Splits {
			if localSplit.Name == remoteSplitName {
				localSchema.Splits[ind].Weights = convertToMapSlice(remoteSplit.Weights)
			}
		}
	}

	schema.Write(localSchema)

	return nil
}

// convertToMapSlice converts a map[string]int to yaml.MapSlice
func convertToMapSlice(weights map[string]int) yaml.MapSlice {
	var mapSlice yaml.MapSlice
	for k, v := range weights {
		mapSlice = append(mapSlice, yaml.MapItem{Key: k, Value: v})
	}
	return mapSlice
}
