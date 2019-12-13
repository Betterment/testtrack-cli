package cmds

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var generateTimestampDoc = `
Write the current timestamp to the file '~/.testtrack/build_timestamp.txt'.
This timestamp can be passed as a param by the Testtrack client when calling the
split registry endpoint from the Testtrack server.
`

func init() {
	rootCmd.AddCommand(generateTimestampCmd)
}

var generateTimestampCmd = &cobra.Command{
	Use:   "generate_timestamp",
	Short: "Write the current timestamp to '~/.testtrack/build_timestamp.txt'",
	Long:  generateTimestampDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateTimestamp()
	},
}

func generateTimestamp() error {
	usr, _ := user.Current()
	TimestampDir := filepath.Join(usr.HomeDir, ".testtrack")
	TimestampFilePath := filepath.Join(TimestampDir, "build_timestamp.txt")

	if _, err := os.Stat(TimestampDir); os.IsNotExist(err) {
		err := os.Mkdir(TimestampDir, 0755)
		if err != nil {
			return err
		}
	}

	timestamp := []byte(time.Now().Format("2006-01-02T15:04:05Z"))
	err := ioutil.WriteFile(TimestampFilePath, timestamp, 0644)

	return err
}
