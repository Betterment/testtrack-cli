package cmds

import (
	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/spf13/cobra"
)

var uncompleteFeatureDoc = `
Marks all versions of this app as feature-incomplete for a feature gate.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature-complete.

Server-side apps will typically ignore this setting and show features
regardless of feature-completeness because there is no legacy code in the
field for customers to use.
`

func init() {
	destroyCmd.AddCommand(uncompleteFeatureCmd)
}

var uncompleteFeatureCmd = &cobra.Command{
	Use:   "feature_completion feature_gate_name",
	Short: "Mark all versions of this app feature-incomplete",
	Long:  uncompleteFeatureDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uncompleteFeature(args[0])
	},
}

func uncompleteFeature(featureGate string) error {
	featureCompletion, err := featurecompletions.New(&featureGate, nil)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(featureCompletion)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
