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
`

var createExperimentWeights string

func init() {
	createExperimentCmd.Flags().StringVar(&createExperimentWeights, "weights", "control: 50, treatment: 50", "Variant weights to use")
	createExperimentCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix experiment with app_name to refer to legacy splits")
	createCmd.AddCommand(createExperimentCmd)
}

var createExperimentCmd = &cobra.Command{
	Use:   "experiment name",
	Short: "Create or update an experiment's configuration",
	Long:  createExperimentDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createExperiment(args[0], createExperimentWeights)
	},
}

func createExperiment(name, weights string) error {
	schema, err := schema.Read()
	if err != nil {
		return err
	}

	err = validations.SplitExistsInSchema("name", &name, schema)
	if err != nil && !noPrefix { // Bare name doesn't exist in schema
		appName, err := getAppName()
		if err != nil {
			return err
		}

		err = validations.NonPrefixedExperiment("name", &name)
		if err != nil {
			return err
		}

		name = fmt.Sprintf("%s.%s", appName, name)
	} else {
		err = validations.ExperimentSuffix("name", &name)
		if err != nil {
			return err
		}
	}

	weightsMap, err := splits.WeightsFromString(weights)
	if err != nil {
		return err
	}

	split, err := splits.New(&name, weightsMap)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(split)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
