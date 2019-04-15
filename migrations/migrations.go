package migrations

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/server"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// IMigration represents a migration
type IMigration interface {
	Create() error
	Run() error
}

// Runner runs outstanding migrations
type Runner struct {
	Server server.IServer
}

// NewRunner returns a Runner ready to use
func NewRunner() (*Runner, error) {
	testTrack, err := server.New()
	if err != nil {
		return nil, err
	}

	return &Runner{Server: testTrack}, nil
}

var migrationFilenameRegex = regexp.MustCompile(`^(\d{13}(?:v\d{3})?)_[a-z\d_]+.yml$`)

// RunOutstanding runs all outstanding migrations
func (r *Runner) RunOutstanding() error {
	migrationVersions, err := r.getMigrationVersions()
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir("testtrack/migrate")
	if err != nil {
		return err
	}

	migrationsByVersion := map[string]IMigration{}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue // Skip hidden files
		}
		matches := migrationFilenameRegex.FindStringSubmatch(file.Name())
		if matches == nil {
			return fmt.Errorf("found foreign file testtrack/migrate/%s in migrations", file.Name())
		}
		migrationVersion := matches[1]
		fileBytes, err := ioutil.ReadFile(path.Join("testtrack/migrate", file.Name()))

		var migrationFile serializers.MigrationFile
		err = yaml.Unmarshal(fileBytes, &migrationFile)
		if err != nil {
			return err
		}

		if migrationFile.FeatureCompletion != nil {
			migrationsByVersion[migrationVersion] = FeatureCompletionFromFile(&migrationVersion, *migrationFile.FeatureCompletion, r.Server)
		} else {
			return fmt.Errorf("testtrack/migrate/%s didn't match a known migration type", file.Name())
		}
	}

	for _, version := range *migrationVersions {
		delete(migrationsByVersion, version.Version)
	}

	versions := make([]string, len(migrationsByVersion))
	i := 0
	for version := range migrationsByVersion {
		versions[i] = version
		i++
	}

	sort.Strings(versions)

	for _, version := range versions {
		err := migrationsByVersion[version].Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) getMigrationVersions() (*[]serializers.MigrationVersion, error) {
	migrationVersions := make([]serializers.MigrationVersion, 0)

	err := r.Server.Get("api/v2/migrations", &migrationVersions)
	if err != nil {
		return nil, err
	}

	return &migrationVersions, nil
}

func generateMigrationVersion() (*string, error) {
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
