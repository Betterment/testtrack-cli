package migrations

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/server"
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

var migrationFilenameRegex = regexp.MustCompile(`^(\d{13})_[a-z\d_]+.yml$`)

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

func generateMigrationVersion() string {
	t := time.Now().UTC()
	nowEpochSeconds := t.Unix()
	todayEpochSeconds := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	secondsIntoToday := nowEpochSeconds - todayEpochSeconds
	return fmt.Sprintf("%04d%02d%02d%05d", t.Year(), t.Month(), t.Day(), secondsIntoToday)
}
