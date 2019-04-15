package cmds

import (
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/spf13/cobra"
)

var uncompleteFeatureDoc = `
Marks a potentially previously-completed feature as incomplete for this app.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature complete.
`

func init() {
	rootCmd.AddCommand(uncompleteFeatureCmd)
}

var uncompleteFeatureCmd = &cobra.Command{
	Use:   "uncomplete_feature [feature_gate] [version]",
	Short: "mark a feature gate as incomplete",
	Long:  uncompleteFeatureDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uncompleteFeature(args[0])
	},
}

func uncompleteFeature(featureGate string) error {
	featureCompletion, err := migrations.NewFeatureCompletion(&featureGate, nil)
	if err != nil {
		return err
	}

	err = featureCompletion.Create()
	if err != nil {
		return err
	}

	return nil
}
