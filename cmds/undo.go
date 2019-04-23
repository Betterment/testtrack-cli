package cmds

import (
	"github.com/Betterment/testtrack-cli/migrationrunners"
	"github.com/spf13/cobra"
)

var undoDoc = `
Unapplies the most recent migration, deletes it from the local TestTrack, and
deletes the migration file.

Undo is for before you've merged a migration so you can remove it from your
local server. After a migration is in a shared branch, you shouldn't use undo.
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
	runner, err := migrationrunners.New()
	if err != nil {
		return err
	}

	return runner.Undo()
}
