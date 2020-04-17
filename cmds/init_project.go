package cmds

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/Betterment/testtrack-cli/schema"
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
	Short: "Set up a project for testtrack CLI",
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

	keepfile, err := os.Create("testtrack/migrate/.keep")
	if err != nil {
		log.Fatal(err)
	}
	keepfile.Close()

	if err := ioutil.WriteFile("testtrack/.gitignore", []byte("build_timestamp\n"), 0644); err != nil {
		log.Fatal(err)
	}

	_, err = schema.Generate()
	return err
}
