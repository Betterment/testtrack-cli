package cmds

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var version string
var build string
var noPrefix bool
var force bool

func init() {
	_, urlSet := os.LookupEnv("TESTTRACK_CLI_URL")
	_, appNameSet := os.LookupEnv("TESTTRACK_APP_NAME")
	if !urlSet && !appNameSet {
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

func getAppName() (string, error) {
	urlString, ok := os.LookupEnv("TESTTRACK_CLI_URL")
	if ok {
		url, err := url.Parse(urlString)
		if err != nil {
			return "", err
		}
		if url.User != nil {
			return url.User.Username(), nil
		}
	}

	appName, ok := os.LookupEnv("TESTTRACK_APP_NAME")
	if !ok {
		return "", errors.New("TESTTRACK_APP_NAME must be set")
	}
	return appName, nil
}
