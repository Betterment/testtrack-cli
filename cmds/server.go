package cmds

import (
	"github.com/Betterment/testtrack-cli/fakeserver"
	"github.com/spf13/cobra"
)

var serverDoc = `
Run a fake TestTrack server for local development, backed by schema.yml files
and nonsense.
`

var listenOn string

const defaultListenOn = "127.0.0.1:8297"

func init() {
	serverCmd.Flags().StringVarP(&listenOn, "listen", "l", defaultListenOn, "IP address and port to listen on")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a fake TestTrack server for local development",
	Long:  serverDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fakeserver.Start(listenOn)
		return nil
	},
}
