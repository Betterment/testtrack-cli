package cmds

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createExperimentDoc = `
Creates or updates an experiment split.

Example:

testtrack create experiment my_fancy_experiment

Experiments will default to having two variants, control and treatment, with
weightings of 50% each.

Weights are specified as a string and must sum to 100:

--weights "variant_1: 25, variant_2: 25, variant_3: 50"

Do not use --no-prefix to create a new split. It can be used to revive a
destroyed split if it was destroyed by mistake, but the migration will fail if
you attempt to create a new split without a prefix.
`

var createExperimentWeights string
var createExperimentOwner string

func init() {
	createExperimentCmd.Flags().StringVar(&createExperimentOwner, "owner", "", "Who owns this feature flag?")
	createExperimentCmd.Flags().StringVar(&createExperimentWeights, "weights", "control: 50, treatment: 50", "Variant weights to use")
	createExperimentCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix experiment with app_name (supports existing legacy splits)")
	createCmd.AddCommand(createExperimentCmd)
}

var createExperimentCmd = &cobra.Command{
	Use:   "experiment name --owner <OWNER>",
	Short: "Create or update an experiment's configuration",
	Long:  createExperimentDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createExperiment(args[0], createExperimentWeights, createExperimentOwner)
	},
}

func createExperiment(name, weights string, owner string) error {
	schema, err := schema.Read()
	if err != nil {
		return err
	}

	err = validations.NonPrefixedExperiment("name", &name)
	if err != nil {
		return err
	}

	appName, err := getAppName()
	if err != nil {
		return err
	}

	err = validations.ValidateOwnerName(owner, ownershipFilename)
	if err != nil {
		return err
	}

	err = validations.AutoPrefixAndValidateSplit("name", &name, appName, schema, noPrefix, force)
	if err != nil {
		// if this errors, we know this is a create (not an update), so maybe prefix
		if !noPrefix {
			name = fmt.Sprintf("%s.%s", appName, name)
		}
	}

	weightsMap, err := splits.WeightsFromString(weights)
	if err != nil {
		return err
	}

	split, err := splits.New(&name, weightsMap, &owner)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(split)
	if err != nil {
		return err
	}

	err = mgr.CreateMigration()
	if err != nil {
		return err
	}

	return nil
}
