package cmds

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var generateTimestampDoc = `
Write the current timestamp to the file '~/.testtrack/build_timestamp.txt'.
This timestamp can be used by the Testtrack client as a timestamp param when
calling the split registry endpoint from the Testtrack server.
`

func init() {
	rootCmd.AddCommand(generateTimestampCmd)
}

var generateTimestampCmd = &cobra.Command{
	Use:   "generate_timestamp",
	Short: "Write the current timstamp to '~/.testtrack/build_timestamp.txt'",
	Long:  generateTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTimestamp()
	},
}

func generateTimestamp() error {
	homeDir, _ := os.UserHomeDir()
	TimestampDir := path.Join(homeDir, ".testtrack")
	TimestampFilePath := filepath.Join(TimestampDir, "build_timestamp.txt")

	if _, err := os.Stat(TimestampDir); os.IsNotExist(err) {
		err := os.Mkdir(TimestampDir, 0755)
		if err != nil {
			return err
		}
	}

	timestamp := []byte(time.Now().Format("2006-01-02T15:04:05Z"))
	err := ioutil.WriteFile(TimestampFilePath, timestamp, 0644)
	if err != nil {
		return err
	}

	return err
}
