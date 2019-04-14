package cmds

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var version string
var build string

func init() {
	if _, ok := os.LookupEnv("TESTTRACK_CLI_URL"); !ok {
		godotenv.Load()
	}
}

var rootCmd = &cobra.Command{
	Use:     "testtrack",
	Short:   "TestTrack Split Config Management",
	Long:    fmt.Sprintf("CLI for managing TestTrack experiments and feature gates\n\nVersion: %s\nBuild: %s", version, build),
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if err, ok := err.(*ExitStatusAwareError); ok {
			os.Exit(err.ExitStatus())
		}
		os.Exit(1)
	}
}
