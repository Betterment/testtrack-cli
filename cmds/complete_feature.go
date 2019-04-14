package cmds

import (
	"errors"
	"fmt"

	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/spf13/cobra"
)

var completeFeatureDoc = `
Marks a feature as complete for this app, allowing app versions greater than or
equal to the specified version to see the feature according to their weights.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature complete.
`

func init() {
	rootCmd.AddCommand(completeFeatureCmd)
}

var completeFeatureCmd = &cobra.Command{
	Use:   "complete_feature [feature_gate] [version]",
	Short: "mark a feature gate as complete",
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

	featureCompletion := migrations.FeatureCompletion{
		FeatureGate: &featureGate,
		Version:     &version,
	}

	err := featureCompletion.Save()
	if err != nil {
		return err
	}

	fmt.Println("Great success!")
	return nil
}
