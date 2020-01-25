package paths

import (
	"os"
	"os/user"
)

// ConfigDir return the fake server's configuration directory
func ConfigDir() (*string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	configDir, ok := os.LookupEnv("TESTTRACK_CONFIG_DIR")
	if !ok {
		configDir = user.HomeDir + "/.testtrack"
	}

	return &configDir, nil
}
