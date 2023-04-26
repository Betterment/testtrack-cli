package cmds

import (
	"fmt"

	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/splits"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createFeatureGateDoc = `
Creates or updates a feature gate split.

Example:

testtrack create feature_gate my_feature_enabled

Feature gates will default to having two variants: true and false, and having a
weighting of 100% false.

You can specify a default with the default flag, or select your own variants
with the weights flag.

Optional weights are specified as a string, must have the variants true and
false, and must sum to 100:

--weights "true: 25, false: 75"

Do not use --no-prefix to create a new split. It can be used to revive a
destroyed split if it was destroyed by mistake, but the migration will fail if
you attempt to create a new split without a prefix.
`

var createFeatureGateDefault, createFeatureGateWeights, createFeatureGateOwner string

func init() {
	createFeatureGateCmd.Flags().StringVar(&createFeatureGateOwner, "owner", "", "Who owns this feature flag?")
	createFeatureGateCmd.Flags().StringVar(&createFeatureGateDefault, "default", "false", "Default variant for your feature flag")
	createFeatureGateCmd.Flags().StringVar(&createFeatureGateWeights, "weights", "", "Variant weights to use (overrides default)")
	createFeatureGateCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix feature gate with app_name (supports existing legacy splits)")
	createCmd.AddCommand(createFeatureGateCmd)
}

var createFeatureGateCmd = &cobra.Command{
	Use:   "feature_gate name",
	Short: "Create or update a feature_gate's configuration",
	Long:  createFeatureGateDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createFeatureGate(args[0], createFeatureGateDefault, createFeatureGateWeights, createFeatureGateOwner)
	},
}

func createFeatureGate(name, defaultVariant, weights string, owner string) error {
	schema, err := schema.Read()
	if err != nil {
		return err
	}

	err = validations.NonPrefixedFeatureGate("name", &name)
	if err != nil {
		return err
	}

	appName, err := getAppName()
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

	err = validations.ValidateOwnerName(owner)
	if err != nil {
		return err
	}

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

	split, err := splits.New(&name, weightsMap, &createFeatureGateOwner)
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
