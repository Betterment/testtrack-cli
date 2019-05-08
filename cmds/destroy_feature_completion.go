package cmds

import (
	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var destroyFeatureCompletionDoc = `
Marks all versions of this app as feature-incomplete for a feature gate.

Apps with clients in the field (e.g. mobile) will only see false for feature
gates until they are marked feature-complete.

Server-side apps will typically ignore this setting and show features
regardless of feature-completeness because there is no legacy code in the
field for customers to use.
`

func init() {
	destroyFeatureCompletionCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix feature gate with app_name to refer to legacy splits")
	destroyFeatureCompletionCmd.Flags().BoolVar(&force, "force", false, "Force creation if feature gate isn't found in schema, e.g. if split is retired")
	destroyCmd.AddCommand(destroyFeatureCompletionCmd)
}

var destroyFeatureCompletionCmd = &cobra.Command{
	Use:   "feature_completion feature_gate_name",
	Short: "Mark all versions of this app feature-incomplete",
	Long:  destroyFeatureCompletionDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return destroyFeatureCompletion(&args[0])
	},
}

func destroyFeatureCompletion(featureGate *string) error {
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

	featureCompletion, err := featurecompletions.New(featureGate, nil)
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
