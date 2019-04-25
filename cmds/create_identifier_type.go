package cmds

import (
	"github.com/Betterment/testtrack-cli/identifiertypes"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createIdentifierTypeDoc = `
Creates an identifier type in TestTrack.

Example:

testtrack create identifier_type myapp_user_id

You will likely only have one or a handful of identifier types across your
whole app ecosystem, wherever visitors identify themselves (e.g. by logging
in).
`

func init() {
	createCmd.AddCommand(createIdentifierTypeCmd)
}

var createIdentifierTypeCmd = &cobra.Command{
	Use:   "identifier_type name",
	Short: "Create an identifier_type",
	Long:  createIdentifierTypeDoc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createIdentifierType(args[0])
	},
}

func createIdentifierType(name string) error {
	err := validations.SnakeCaseParam("name", &name)
	if err != nil {
		return err
	}

	identifierType, err := identifiertypes.New(&name)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(identifierType)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
