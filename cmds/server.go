package cmds

import (
	"github.com/Betterment/testtrack-cli/fakeserver"
	"github.com/spf13/cobra"
)

var serverDoc = `
Run a fake TestTrack server for local development, backed by shema.yml files
and nonsense.
`

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a fake TestTrack server for local development",
	Long:  serverDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fakeserver.Start()
		return nil
	},
}
