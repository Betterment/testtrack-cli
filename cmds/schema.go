package cmds

import (
	"github.com/spf13/cobra"
)

var schemaDoc = `
Manage your testtrack schema by generating fresh from migrations or loading the
schema state into a TestTrack server.
`

func init() {
	rootCmd.AddCommand(schemaCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Regenerate or load your testtrack schema file",
	Long:  schemaDoc,
}
