package cmds

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var generateTimestampDoc = `
Write the current UTC timestamp to the file 'testtrack/build_timestamp.txt' in a
TestTrack project. This timestamp can be passed as a param by the TestTrack
client when calling the split registry endpoint from the TestTrack server.
`

func init() {
	rootCmd.AddCommand(generateTimestampCmd)
}

var generateTimestampCmd = &cobra.Command{
	Use:   "generate_timestamp",
	Short: "Write the current UTC timestamp to 'testtrack/build_timestamp.txt'",
	Long:  generateTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTimestamp()
	},
}

func generateTimestamp() error {
	const buildTimestampPath = "testtrack/build_timestamp.txt"
	timestamp := []byte(time.Now().UTC().Format("2006-01-02T15:04:05Z"))

	err := ioutil.WriteFile(buildTimestampPath, timestamp, 0644)
	if e, ok := err.(*os.PathError); ok {
		if e.Path == buildTimestampPath {
			log.Fatal("Testtrack Directory Not Found: Make sure you are in a TestTrack project")
		}
	} else if err != nil {
		log.Fatal(err)
	}

	return nil
}
