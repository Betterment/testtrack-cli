package cmds

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/splitdecisions"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var decideDoc = `
Decides a split (a feature gate or experiment) in TestTrack or edits a previous
decision, automatically reviving it if it was previously destroyed. Decided
splits will continue to be returned to all clients, so it's appropriate to
decide a split before you remove all references to the split from code.

Example:

testtrack decide my_fancy_experiment --variant treatment

You may decide the same split multiple times to amend the decision, or later
retire it via 'destroy split' or undecide and reweight it via 'create
experiment' or 'create feature_flag'
`

var decideVariant string

func init() {
	decideCmd.Flags().StringVar(&decideVariant, "variant", "", "Variant that all clients should see going forward")
	decideCmd.MarkFlagRequired("variant")
	rootCmd.AddCommand(decideCmd)
}

var decideCmd = &cobra.Command{
	Use:   "decide split_name",
	Short: "Decide a split,  or modify a retired split's decision",
	Long:  decideDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return decide(args[0], decideVariant)
	},
}

func decide(name, variant string) error {
	appName, err := getAppName()
	if err != nil {
		return err
	}

	err = validations.NonPrefixedSplit("name", &name)
	if err != nil {
		return err
	}

	name = fmt.Sprintf("%s.%s", appName, name)

	splitDecision, err := splitdecisions.New(&name, &variant)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(splitDecision)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
