package cmds

import (
	"github.com/spf13/cobra"
)

var destroyDoc = `
Immediately destroy a resource in the local TestTrack and write a migration
file so the change can be applied in other environments via the build/deploy
pipeline.
`

func init() {
	rootCmd.AddCommand(destroyCmd)
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a TestTrack resource",
	Long:  destroyDoc,
}
