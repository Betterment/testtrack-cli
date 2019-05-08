package cmds

import (
	"errors"
	"strings"

	"github.com/Betterment/testtrack-cli/fakeassignments"
	"github.com/Betterment/testtrack-cli/serializers"
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

var unassignAll bool

func init() {
	unassignCmd.Flags().BoolVar(&unassignAll, "all", false, "Unassign all splits")
	unassignCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix split with app_name (supports legacy splits)")
	rootCmd.AddCommand(unassignCmd)
}

var unassignCmd = &cobra.Command{
	Use:   "unassign [split_name]",
	Short: "Remove a split assignment",
	Long:  unassignDoc,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) == 1 {
			name = args[0]
		}
		return unassign(name, unassignAll)
	},
}

func unassign(name string, all bool) error {
	nameProvided := len(name) > 0
	if all == nameProvided {
		return errors.New("split_name and --all are mutually exclusive but one is required")
	}

	if all {
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

	fakeAssigns, err := fakeassignments.Read()
	if err != nil {
		return err
	}

	fakeSchema := &serializers.Schema{}
	for split := range *fakeAssigns {
		fakeSchema.Splits = append(fakeSchema.Splits, serializers.SchemaSplit{
			Name: split,
		})
	}

	err = validations.AutoPrefixAndValidateSplit("split_name", &name, currentAppName, fakeSchema, noPrefix, false)
	if err != nil {
		return errors.New(strings.Replace(err.Error(), " in schema", " in assignments", 1))
	}

	delete(*fakeAssigns, name)
	err = fakeassignments.Write(fakeAssigns)
	if err != nil {
		return err
	}

	return nil
}
