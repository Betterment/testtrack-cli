package paths

import (
	"os"
	"os/user"
)

// FakeServerConfigDir return the fake server's configuration directory
func FakeServerConfigDir() (*string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	configDir, ok := os.LookupEnv("TESTTRACK_FAKE_SERVER_CONFIG_DIR")
	if !ok {
		configDir = user.HomeDir + "/.testtrack"
	}

	return &configDir, nil
}
