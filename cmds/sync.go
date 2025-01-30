package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

func readYAML(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening YAML file: %v", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(fileData, &yamlData); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	return yamlData, nil
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

	splits, ok := jsonData["splits"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Error: 'splits' key not found or not a map")
	}

	yamlFilePath := "testtrack/schema.yml"
	yamlData, err := readYAML(yamlFilePath)
	if err != nil {
		return fmt.Errorf("Error reading YAML file: %v", err)
	}

	for key, value := range splits {
		for _, split := range yamlData["splits"].([]interface{}) {
			splitMap, ok := split.(map[interface{}]interface{})
			if !ok {
				continue
			}
			if splitMap["name"] == key {
				valueMap, ok := value.(map[string]interface{})
				if !ok {
					continue
				}
				splitMap["weights"] = valueMap["weights"]
			}
		}
	}

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("error marshalling YAML: %v", err)
	}

	err = os.WriteFile(yamlFilePath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing YAML file: %v", err)
	}

	return nil
}
