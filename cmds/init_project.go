package cmds

import (
	"os"

	"github.com/spf13/cobra"
)

var initProjectDoc = `
Sets up your project for testtrack CLI usage, with a directory structure for
split config migrations that you'll commit alongside your code.
`

func init() {
	rootCmd.AddCommand(initProjectCmd)
}

var initProjectCmd = &cobra.Command{
	Use:   "init_project",
	Short: "Set up a project for testtrack",
	Long:  initProjectDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initProject()
	},
}

func initProject() error {
	err := os.MkdirAll("testtrack/migrate", 0755)
	if err != nil {
		return err
	}
	return nil
}
