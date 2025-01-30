package cmds

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Betterment/testtrack-cli/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var syncDoc = `
Sync the local schema TestTrack assignments with the remote production TestTrack assignments.

Example:

testtrack sync http:://example.com/split_registry.json
`

func init() {
	rootCmd.AddCommand(syncCommand)
}

var syncCommand = &cobra.Command{
	Use:   "sync",
	Short: "Sync TestTrack assignments with production",
	Long:  syncDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return Sync(args[0])
	},
}

func toMapSlice(m map[string]interface{}) yaml.MapSlice {
	mapSlice := yaml.MapSlice{}
	for k, v := range m {
		mapSlice = append(mapSlice, yaml.MapItem{Key: k, Value: v})
	}
	return mapSlice
}

// Sync synchronizes the local schema TestTrack assignments with the remote production TestTrack assignments.
func Sync(remoteURL string) error {
	res, err := http.Get(remoteURL)

	if err != nil {
		return fmt.Errorf("Error fetching JSON: %v", err)
	}

	defer res.Body.Close()
	var jsonData map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&jsonData); err != nil {
		return fmt.Errorf("Error decoding JSON: %v", err)
	}

	remoteSplits, ok := jsonData["splits"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Error: 'splits' key not found or not a map")
	}

	localSchema, err := schema.Read()
	if err != nil {
		return err
	}

	for remoteSplitName, remoteWeight := range remoteSplits {
		for ind, localSplit := range localSchema.Splits {
			if localSplit.Name == remoteSplitName {
				remoteWeightMap, ok := remoteWeight.(map[string]interface{})
				if !ok {
					return fmt.Errorf("failed to cast remoteWeight to map[string]interface{}")
				}
				if weightsMap, ok := remoteWeightMap["weights"].(map[string]interface{}); ok {
					localSchema.Splits[ind].Weights = toMapSlice(weightsMap)
				} else {
					return fmt.Errorf("failed to cast weights to yaml.MapSlice")
				}
			}
		}
	}

	schema.Write(localSchema)

	return nil
}
