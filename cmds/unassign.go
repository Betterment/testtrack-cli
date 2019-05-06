package cmds

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/fakeassignments"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var unassignDoc = `
Removes split assignment from all visitors in the testtrack fake server.

This command can be used to remove the assignment for a specific split
or to remove all split assignments.

Example:

testtrack unassign my_fancy_experiment
testtrack unassign --all
`

var unassignVariant string
var unassignAll bool

func init() {
	unassignCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix split with app_name (supports legacy splits)")
	unassignCmd.Flags().BoolVar(&unassignAll, "all", false, "Unassign all splits")
	rootCmd.AddCommand(unassignCmd)
}

var unassignCmd = &cobra.Command{
	Use:   "unassign [split_name]",
	Short: "Remove a split assignment",
	Long:  unassignDoc,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return unassign(name, unassignAll)
	},
}

func unassign(name string, all bool) error {
	if all == true {
		return runUnassignAll()
	}
	return runUnassign(name)
}

func runUnassignAll() error {
	fakeAssigns := make(map[string]string)
	err := fakeassignments.Write(&fakeAssigns)
	if err != nil {
		return err
	}
	return nil
}

func runUnassign(name string) error {
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
		return fmt.Errorf("split_name '%s' not found in schema", name)
	}

	fakeAssigns, err := fakeassignments.Read()
	if err != nil {
		return err
	}

	delete(*fakeAssigns, name)
	err = fakeassignments.Write(fakeAssigns)
	if err != nil {
		return err
	}

	return nil
}
