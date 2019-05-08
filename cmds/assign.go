package cmds

import (
	"github.com/Betterment/testtrack-cli/fakeassignments"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var assignDoc = `
Overrides all assignments for a split in the fake TestTrack server.

Example:

testtrack assign my_fancy_experiment --variant treatment
`

var assignVariant string

func init() {
	assignCmd.Flags().StringVar(&assignVariant, "variant", "", "Variant to assign")
	assignCmd.MarkFlagRequired("variant")
	assignCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix split with app_name (supports legacy splits)")
	rootCmd.AddCommand(assignCmd)
}

var assignCmd = &cobra.Command{
	Use:   "assign split_name",
	Short: "Assign a variant of a split",
	Long:  assignDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return assign(args[0], assignVariant)
	},
}

func assign(name, variant string) error {
	currentAppName, err := getAppName()
	if err != nil {
		return err
	}
	mergedSchema, err := schema.ReadMerged()
	if err != nil {
		return err
	}
	err = validations.AutoPrefixAndValidateSplit("split_name", &name, currentAppName, mergedSchema, noPrefix, false)
	if err != nil {
		return err
	}
	err = validations.VariantExistsInSchema("variant", &variant, name, mergedSchema)
	if err != nil {
		return err
	}

	fakeAssigns, err := fakeassignments.Read()
	if err != nil {
		return err
	}

	(*fakeAssigns)[name] = variant
	err = fakeassignments.Write(fakeAssigns)
	if err != nil {
		return err
	}

	return nil
}
