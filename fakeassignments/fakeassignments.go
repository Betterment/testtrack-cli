package fakeassignments

import (
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

// Read reads or creates the assignment file
func Read() (*map[string]string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(user.HomeDir + "/.testtrack/assignments.yml"); os.IsNotExist(err) {
		err := os.MkdirAll(user.HomeDir+"/.testtrack", 0755)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(user.HomeDir+"/.testtrack/assignments.yml", []byte("{}"), 0644)
		if err != nil {
			return nil, err
		}
	}
	assignmentsBytes, err := ioutil.ReadFile(user.HomeDir + "/.testtrack/assignments.yml")
	if err != nil {
		return nil, err
	}
	var assignments map[string]string
	err = yaml.Unmarshal(assignmentsBytes, &assignments)
	if err != nil {
		return nil, err
	}
	return &assignments, nil
}

// Write dumps the assignment file to disk
func Write(assignments *map[string]string) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(assignments)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(user.HomeDir+"/.testtrack/assignments.yml", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
