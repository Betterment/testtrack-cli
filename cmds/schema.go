package cmds

import (
	"github.com/spf13/cobra"
)

var schemaDoc = `
Manage your testtrack schema by generating fresh from migrations, loading the
schema state into a TestTrack server, or linking your schema for 'testtrack
server' to use.
`

func init() {
	rootCmd.AddCommand(schemaCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage your testtrack schema file",
	Long:  schemaDoc,
}
