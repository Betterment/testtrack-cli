package cmds

import (
	"errors"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/spf13/cobra"
)

var completeFeatureDoc = `
Marks a verson of this app as feature-complete, allowing end-users with app
versions greater than or equal to the specified version to see the feature
according to their weights.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature-complete.

Server-side apps will typically ignore this setting and show features
regardless of feature-completeness because there is no legacy code in the
field for customers to use.
`

func init() {
	rootCmd.AddCommand(completeFeatureCmd)
}

var completeFeatureCmd = &cobra.Command{
	Use:   "complete_feature [feature_gate] [app_version]",
	Short: "Mark an app version feature-complete",
	Long:  completeFeatureDoc,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return completeFeature(args[0], args[1])
	},
}

func completeFeature(featureGate, version string) error {
	if len(version) == 0 {
		return errors.New("Version must be present")
	}

	featureCompletion, err := migrations.NewFeatureCompletion(&featureGate, &version)
	if err != nil {
		return err
	}

	err = featureCompletion.Create()
	if err != nil {
		return err
	}

	return nil
}
