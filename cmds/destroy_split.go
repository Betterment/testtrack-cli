package cmds

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/splitretirements"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var destroySplitDoc = `
Retires a split (a feature gate or experiment) in TestTrack or updates a
retired split's decision. The split will continue to be returned to clients in
the field built before it was retired.

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
	destroyCmd.AddCommand(destroySplitCmd)
}

var destroySplitCmd = &cobra.Command{
	Use:   "split name",
	Short: "Retire a split or modify a retired split's decision",
	Long:  destroySplitDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return destroySplit(args[0], destroySplitDecision)
	},
}

func destroySplit(name, decision string) error {
	appName, err := getAppName()
	if err != nil {
		return err
	}

	err = validations.NonPrefixedSplit("name", &name)
	if err != nil {
		return err
	}

	name = fmt.Sprintf("%s.%s", appName, name)

	splitRetirement, err := splitretirements.New(&name, &decision)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(splitRetirement)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
