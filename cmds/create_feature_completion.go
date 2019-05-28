package cmds

import (
	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createFeatureCompletionDoc = `
Marks a version of this app as feature-complete for a feature gate, allowing
end-users with app versions greater than or equal to the specified version to
see the feature according to their weights.

Example:

testtrack create feature_completion foo_enabled --app_version 1.0

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature-complete.

Server-side apps will typically ignore this setting and show features
regardless of feature-completeness because there is no legacy code in the
field for customers to use.

You can reverse complete_feature with the uncomplete_feature command.
`

var createFeatureCompletionAppVersion string

func init() {
	createFeatureCompletionCmd.Flags().StringVar(&createFeatureCompletionAppVersion, "app_version", "", "App version (required)")
	createFeatureCompletionCmd.MarkFlagRequired("app_version")
	createFeatureCompletionCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix feature gate with app_name to refer to legacy splits")
	createFeatureCompletionCmd.Flags().BoolVar(&force, "force", false, "Force creation if feature gate isn't found in schema, e.g. if split is retired")
	createCmd.AddCommand(createFeatureCompletionCmd)
}

var createFeatureCompletionCmd = &cobra.Command{
	Use:   "feature_completion feature_gate_name",
	Short: "Mark an app version feature-complete for a feature gate",
	Long:  createFeatureCompletionDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createFeatureCompletion(&args[0], &createFeatureCompletionAppVersion)
	},
}

func createFeatureCompletion(featureGate, version *string) error {
	currentAppName, err := getAppName()
	if err != nil {
		return err
	}
	mergedSchema, err := schema.ReadMerged()
	if err != nil {
		return err
	}
	err = validations.AutoPrefixAndValidateSplit("feature_gate_name", featureGate, currentAppName, mergedSchema, noPrefix, force)
	if err != nil {
		return err
	}
	err = validations.FeatureGateSuffix("feature_gate_name", featureGate)
	if err != nil {
		return err
	}

	// This validation is the difference between complete_feature and uncomplete_feature which is why it's inline
	err = validations.Presence("app_version", version)
	if err != nil {
		return err
	}

	featureCompletion, err := featurecompletions.New(featureGate, version)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(featureCompletion)
	if err != nil {
		return err
	}

	err = mgr.CreateMigration()
	if err != nil {
		return err
	}

	return nil
}
