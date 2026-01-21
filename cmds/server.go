package cmds

import (
	"github.com/Betterment/testtrack-cli/fakeserver"
	"github.com/spf13/cobra"
)

var serverDoc = `
Run a fake TestTrack server for local development, backed by schema.{json,yml} files
and nonsense.
`

var port int

const defaultPort = 8297

func init() {
	serverCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to listen on")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a fake TestTrack server for local development",
	Long:  serverDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fakeserver.Start(port)
		return nil
	},
}
