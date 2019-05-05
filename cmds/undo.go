package cmds

import (
	"github.com/Betterment/testtrack-cli/migrationrunners"
	"github.com/spf13/cobra"
)

var undoDoc = `
Unapplies the most recent migration from your schema, and deletes the migration
file.

Undo is for before you've merged a migration so you can remove it from your
local machine. After a migration is in a shared branch, you shouldn't use undo.
`

func init() {
	rootCmd.AddCommand(undoCmd)
}

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Unapply and delete the latest migration",
	Long:  undoDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return undo()
	},
}

func undo() error {
	runner, err := migrationrunners.New(nil)
	if err != nil {
		return err
	}

	return runner.Undo()
}
