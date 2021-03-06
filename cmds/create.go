package cmds

import (
	"github.com/spf13/cobra"
)

var createDoc = `
Create a resource in the local schema and write a migration file so the change
can be applied in other environments via the build/deploy pipeline.
`

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a TestTrack resource",
	Long:  createDoc,
}
