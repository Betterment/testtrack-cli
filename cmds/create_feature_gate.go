package cmds

import (
	"errors"
	"fmt"
	"os"

	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createFeatureGateDoc = `
Creates a feature gate in TestTrack.

Example:

testtrack create feature_gate my_feature_enabled

Feature gates will default to having two variants: true and false, and having a
weighting of 100% false.

You can specify a default with the default flag, or select your own variants
with the weights flag.

Weights are specified as a string and must sum to 100:

--weights "variant_1: 25, variant_2: 25, variant_3: 50"
`

var defaultVariant, weights string

func init() {
	createFeatureGateCmd.Flags().StringVar(&defaultVariant, "default", "false", "Default variant for your feature flag")
	createFeatureGateCmd.Flags().StringVar(&weights, "weights", "", "Variant weights to use (overrides default)")
	createCmd.AddCommand(createFeatureGateCmd)
}

var createFeatureGateCmd = &cobra.Command{
	Use:   "feature_gate name",
	Short: "Create or update a feature_gate's configuration",
	Long:  createFeatureGateDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createFeatureGate(args[0], defaultVariant, weights)
	},
}

func createFeatureGate(name, defaultVariant, weights string) error {
	appName, ok := os.LookupEnv("TESTTRACK_APP_NAME")
	if !ok {
		return errors.New("TESTTRACK_APP_NAME must be set")
	}

	err := validations.NonPrefixedFeatureGate("name", &name)
	if err != nil {
		return err
	}

	name = fmt.Sprintf("%s.%s", appName, name)

	if len(weights) == 0 {
		switch defaultVariant {
		case "true":
			weights = "false: 0, true: 100"
		case "false":
			weights = "false: 100, true: 0"
		default:
			return fmt.Errorf("default %s must be either 'true' or 'false'", defaultVariant)
		}
	}

	weightsMap, err := splits.WeightsFromString(weights)
	if err != nil {
		return err
	}

	if len(*weightsMap) != 2 {
		return fmt.Errorf("weights %v must contain exactly two variants, true and false", *weightsMap)
	}

	if _, ok := (*weightsMap)["true"]; !ok {
		return fmt.Errorf("weights %v are missing true variant", *weightsMap)
	}

	if _, ok := (*weightsMap)["false"]; !ok {
		return fmt.Errorf("weights %v are missing false variant", *weightsMap)
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
