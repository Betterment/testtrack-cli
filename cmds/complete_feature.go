package cmds

import (
	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var completeFeatureDoc = `
Marks a version of this app as feature-complete for a feature gate, allowing
end-users with app versions greater than or equal to the specified version to
see the feature according to their weights.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature-complete.

Server-side apps will typically ignore this setting and show features
regardless of feature-completeness because there is no legacy code in the
field for customers to use.

You can reverse complete_feature with the uncomplete_feature command.
`

func init() {
	rootCmd.AddCommand(completeFeatureCmd)
}

var completeFeatureCmd = &cobra.Command{
	Use:   "complete_feature feature_gate_name app_version",
	Short: "Mark an app version feature-complete for a feature gate",
	Long:  completeFeatureDoc,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return completeFeature(args[0], args[1])
	},
}

func completeFeature(featureGate, version string) error {
	param := "app_version"
	// This validation is the difference between complete_feature and uncomplete_feature which is why it's inline
	err := validations.Presence(&version, &param)
	if err != nil {
		return err
	}

	featureCompletion, err := featurecompletions.New(&featureGate, &version)
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
