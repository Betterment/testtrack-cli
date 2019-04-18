package migrations

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/pkg/errors"
)

// IMigration represents a migration
type IMigration interface {
	Validate() error
	File() *serializers.MigrationFile
	Filename() *string
	SyncPath() string
	Serializable() interface{}
	MigrationVersion() *string
}

var migrationFilenameRegex = regexp.MustCompile(`^(\d{13}(?:v\d{3})?)_[a-z\d_]+.yml$`)

// GenerateMigrationVersion returns a new timestamp-derived migration version
func GenerateMigrationVersion() (*string, error) {
	t := time.Now().UTC()
	nowEpochSeconds := t.Unix()
	todayEpochSeconds := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	secondsIntoToday := nowEpochSeconds - todayEpochSeconds

	baseVersion := fmt.Sprintf("%04d%02d%02d%05d", t.Year(), t.Month(), t.Day(), secondsIntoToday)

	matches, err := filepath.Glob(fmt.Sprintf("testtrack/migrate/%s*", baseVersion))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return &baseVersion, nil
	}

	sort.Strings(matches)

	lastMatch := matches[len(matches)-1]
	matches = migrationFilenameRegex.FindStringSubmatch(filepath.Base(lastMatch))
	if matches == nil {
		return nil, fmt.Errorf("Failed to parse migration filename %s", lastMatch)
	}

	var i int
	if len(matches[1]) == 13 {
		i = 1
	} else if len(matches[1]) == 17 {
		i, err = strconv.Atoi(matches[1][14:17])
		if err != nil {
			return nil, errors.Wrap(err, "couldn't parse file version")
		}
		i++
	} else {
		return nil, errors.New("unexpected file version length")
	}

	longVersion := fmt.Sprintf("%sv%03d", baseVersion, i)
	return &longVersion, nil
}

// ExtractVersionFromFilename returns the migration version from a filename
func ExtractVersionFromFilename(filename string) (string, error) {
	matches := migrationFilenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return "", fmt.Errorf("found foreign file testtrack/migrate/%s in migrations", filename)
	}
	return matches[1], nil
}
