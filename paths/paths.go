package paths

import (
	"os"
	"os/user"
)

// HomeDir return home directory path
func HomeDir() (*string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	testTrackHomeDir, ok := os.LookupEnv("TESTTRACK_HOME_DIR")
	if !ok {
		testTrackHomeDir = user.HomeDir + "/.testtrack"
	}

	return &testTrackHomeDir, nil
}
