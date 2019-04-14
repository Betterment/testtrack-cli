package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
)

var markFeatureCompleteDoc = `
Marks a feature as complete for this app, allowing app versions greater than or
equal to the specified version to see the feature according to their weights.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature complete.
`

func init() {
	rootCmd.AddCommand(markFeatureCompleteCmd)
}

var markFeatureCompleteCmd = &cobra.Command{
	Use:   "mark_feature_complete [feature_gate_name] [app_version_number]",
	Short: "mark a feature flag as complete",
	Long:  markFeatureCompleteDoc,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return markFeatureComplete()
	},
}

func markFeatureComplete() error {
	fmt.Println("Booya!")
	return nil
}
