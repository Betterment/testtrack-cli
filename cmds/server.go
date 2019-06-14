package cmds

import (
	"github.com/Betterment/testtrack-cli/fakeserver"
	"github.com/spf13/cobra"
)

var serverDoc = `
Run a fake TestTrack server for local development, backed by schema.yml files
and nonsense.
`

var port int
var logRequests bool

const defaultPort = 8297

func init() {
	serverCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to listen on")
	serverCmd.Flags().BoolVarP(&logRequests, "log-requests", "", false, "Log requests to stderr")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a fake TestTrack server for local development",
	Long:  serverDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fakeserver.Start(port, logRequests)
		return nil
	},
}
