package cmds

import (
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
)

var generateTimestampDoc = `
Write the current timestamp to the file 'build_timestamp.txt' in a TestTrack project.
This timestamp can be passed as a param by the TestTrack client when calling the
split registry endpoint from the TestTrack server.
`

func init() {
	rootCmd.AddCommand(generateTimestampCmd)
}

var generateTimestampCmd = &cobra.Command{
	Use:   "generate_timestamp",
	Short: "Write the current timestamp to 'testtrack/build_timestamp.txt'",
	Long:  generateTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTimestamp()
	},
}

func generateTimestamp() error {
	timestamp := []byte(time.Now().Format("2006-01-02T15:04:05Z"))
	err := ioutil.WriteFile("testtrack/build_timestamp.txt", timestamp, 0644)

	return err
}
