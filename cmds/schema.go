package cmds

import (
	"github.com/spf13/cobra"
)

var schemaDoc = `
Manage your testtrack schema by dumping cumulative migration state to a schema
file or loading a schema file TestTrack server.
`

func init() {
	rootCmd.AddCommand(schemaCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Dump or load your testtrack schema file",
	Long:  schemaDoc,
}
