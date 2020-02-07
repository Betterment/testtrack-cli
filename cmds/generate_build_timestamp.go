package cmds

import (
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
)

var generateBuildTimestampDoc = `
Write the current UTC timestamp to the file 'testtrack/build_timestamp'.
This timestamp is used by TestTrack clients to request the version of the split
registry active at the given time.
`

func init() {
	rootCmd.AddCommand(generateBuildTimestampCmd)
}

var generateBuildTimestampCmd = &cobra.Command{
	Use:   "generate_build_timestamp",
	Short: "Write the current UTC timestamp to 'testtrack/build_timestamp'",
	Long:  generateBuildTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateBuildTimestamp()
	},
}

func generateBuildTimestamp() error {
	timestamp := []byte(time.Now().UTC().Format("2006-01-02T15:04:05Z"))

	return ioutil.WriteFile("testtrack/build_timestamp", timestamp, 0644)
}
