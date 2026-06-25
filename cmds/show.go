package cmds

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"github.com/spf13/cobra"
)

var showJSON bool

var showDoc = `
Show the current variant weights for a split from the remote TestTrack server.

Reads the split registry from the server configured by TESTTRACK_CLI_URL and
prints the weights for the named split. It does not modify the local schema.

The split name is matched verbatim against the remote registry, so pass the
fully-qualified name (e.g. my_app.my_feature_enabled).
`

func init() {
	showCommand.Flags().BoolVar(&showJSON, "json", false, "output weights as JSON")
	rootCmd.AddCommand(showCommand)
}

var showCommand = &cobra.Command{
	Use:   "show <split>",
	Short: "Show remote variant weights for a split",
	Long:  showDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return Show(args[0], showJSON)
	},
}

// Show prints the remote variant weights for the named split.
func Show(name string, asJSON bool) error {
	server, err := servers.New()
	if err != nil {
		return err
	}

	output, err := splitWeights(server, name, asJSON)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// splitWeights fetches the named split from the server and renders its weights.
func splitWeights(server servers.IServer, name string, asJSON bool) (string, error) {
	var registry serializers.RemoteRegistry
	if err := server.Get("api/v2/split_registry", &registry); err != nil {
		return "", err
	}

	split, ok := registry.Splits[name]
	if !ok {
		return "", fmt.Errorf("split %q not found in remote registry; check the name or run `testtrack sync`", name)
	}

	return formatWeights(name, split.Weights, asJSON)
}

// formatWeights renders a split's variant weights for display.
func formatWeights(name string, weights map[string]int, asJSON bool) (string, error) {
	if asJSON {
		bytes, err := json.Marshal(weights)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	variants := make([]string, 0, len(weights))
	nameWidth, weightWidth := 0, 0
	for variant, weight := range weights {
		variants = append(variants, variant)
		if len(variant) > nameWidth {
			nameWidth = len(variant)
		}
		if w := len(strconv.Itoa(weight)); w > weightWidth {
			weightWidth = w
		}
	}
	sort.Strings(variants)

	var b strings.Builder
	b.WriteString(name)
	for _, variant := range variants {
		fmt.Fprintf(&b, "\n  %-*s  %*d%%", nameWidth, variant, weightWidth, weights[variant])
	}
	return b.String(), nil
}
