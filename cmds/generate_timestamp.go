package cmds

import (
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
)

var generateTimestampDoc = `
Write the current UTC timestamp to the file 'testtrack/build_timestamp'.
This timestamp is used by TestTrack clients to request the version of the split
registry active at the given time.
`

func init() {
	rootCmd.AddCommand(generateTimestampCmd)
}

var generateTimestampCmd = &cobra.Command{
	Use:   "generate_timestamp",
	Short: "Write the current UTC timestamp to 'testtrack/build_timestamp'",
	Long:  generateTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTimestamp()
	},
}

func generateTimestamp() error {
	timestamp := []byte(time.Now().UTC().Format("2006-01-02T15:04:05Z"))

	return ioutil.WriteFile("testtrack/build_timestamp", timestamp, 0644)
}
