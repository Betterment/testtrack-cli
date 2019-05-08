package cmds

import (
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/splitretirements"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var destroySplitDoc = `
Destroy soft-deletes a split (i.e. a feature gate or experiment) or updates an
already-destroyed split's decision. The split will continue to be returned to
clients in the field built before it was destroyed.

Destroying a split is also known as retirement. Retired splits have a decision
so old clients know which variant to choose.

Feature-completion and remote-kill both take precedence over the decision of a
retired split, though, so clients with incomplete or broken versions of a
feature will not be enabled for the split even if a retired split's decision
was to enable the feature.

Example:

testtrack destroy split my_fancy_experiment --decision treatment

You must provide a decision to retire a split so clients in the field will know
which variant to choose.

You may retire the same split multiple times to amend the decision, or revive
it by recreating it via 'create experiment' or 'create feature_flag'
`

var destroySplitDecision string

func init() {
	destroySplitCmd.Flags().StringVar(&destroySplitDecision, "decision", "", "Variant that clients in the field should see after retirement")
	destroySplitCmd.MarkFlagRequired("decision")
	destroySplitCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix split with app_name (supports legacy splits)")
	destroySplitCmd.Flags().BoolVar(&force, "force", false, "Force destroy if split isn't found in schema, e.g. if split is retired")
	destroyCmd.AddCommand(destroySplitCmd)
}

var destroySplitCmd = &cobra.Command{
	Use:   "split name",
	Short: "Destroy (retire) a split or modify a retired split's decision",
	Long:  destroySplitDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return destroySplit(args[0], destroySplitDecision)
	},
}

func destroySplit(name, decision string) error {
	schema, err := schema.Read()
	if err != nil {
		return err
	}

	err = validations.NonPrefixedSplit("name", &name)
	if err != nil {
		return err
	}

	appName, err := getAppName()
	if err != nil {
		return err
	}

	err = validations.AutoPrefixAndValidateSplit("name", &name, appName, schema, noPrefix, force)
	if err != nil {
		return err
	}

	splitRetirement, err := splitretirements.New(&name, &decision)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(splitRetirement)
	if err != nil {
		return err
	}

	err = mgr.CreateMigration()
	if err != nil {
		return err
	}

	return nil
}
