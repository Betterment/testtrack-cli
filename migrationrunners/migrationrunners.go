package migrationrunners

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/Betterment/testtrack-cli/featurecompletions"
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/migrations"
	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/servers"
	"gopkg.in/yaml.v2"
)

// Runner runs sets of migrations
type Runner struct {
	server servers.IServer
}

// New returns a Runner ready to use
func New() (*Runner, error) {
	server, err := servers.New()
	if err != nil {
		return nil, err
	}

	return &Runner{server: server}, nil
}

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

	migrationsByVersion := map[string]migrations.IMigration{}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue // Skip hidden files
		}

		migrationVersion, err := migrations.ExtractVersionFromFilename(file.Name())
		if err != nil {
			return err
		}

		fileBytes, err := ioutil.ReadFile(path.Join("testtrack/migrate", file.Name()))
		if err != nil {
			return err
		}

		var migrationFile serializers.MigrationFile
		err = yaml.Unmarshal(fileBytes, &migrationFile)
		if err != nil {
			return err
		}

		if migrationFile.FeatureCompletion != nil {
			migrationsByVersion[migrationVersion] = featurecompletions.FromFile(&migrationVersion, migrationFile.FeatureCompletion)
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
		mgr := migrationmanagers.NewWithServer(migrationsByVersion[version], r.server)
		err := mgr.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) getMigrationVersions() (*[]serializers.MigrationVersion, error) {
	migrationVersions := make([]serializers.MigrationVersion, 0)

	err := r.server.Get("api/v2/migrations", &migrationVersions)
	if err != nil {
		return nil, err
	}

	return &migrationVersions, nil
}
