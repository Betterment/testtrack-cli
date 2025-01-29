// fetch json from https://tt.betterment.com/api/v2/split_registry.json
// iterate through splits that start with "retail."
// replace assignments of values at ~/src/retail/testtrack/schema.yml

package cmds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		Sync()
	},
}

func readYAML(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
			return nil, fmt.Errorf("error opening YAML file: %v", err)
	}
	defer file.Close()

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(fileData, &yamlData); err != nil {
			return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	return yamlData, nil
}

func Sync() {
	url := "https://tt.betterment.com/api/v2/split_registry.json"
	res, err := http.Get(url)

	if err != nil {
		fmt.Printf("Error fetching JSON: %v\n", err)
		return
	}

	defer res.Body.Close()
	var jsonData map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&jsonData); err != nil {
			fmt.Printf("Error decoding JSON: %v\n", err)
			return
	}

	splits, ok := jsonData["splits"].(map[string]interface{})
	if !ok {
			fmt.Println("Error: 'splits' key not found or not a map")
			return
	}

	// yamlFilePath := "testtrack/schema.yml"
	yamlFilePath := "../../retail/retail/testtrack/schema.yml"
	yamlData, err := readYAML(yamlFilePath)
	if err != nil {
			fmt.Printf("Error reading YAML file: %v\n", err)
			return
	}

	fmt.Printf("YAML data: %+v\n", yamlData)

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

	fmt.Printf("Yaml data:", yamlData)

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		fmt.Errorf("error marshalling YAML: %v", err)
	}

	err = ioutil.WriteFile(yamlFilePath, yamlBytes, 0644)
	if err != nil {
		fmt.Errorf("error writing YAML file: %v", err)
	}
}
