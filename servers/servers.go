package servers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// IServer is the interface of the TestTrack API client
type IServer interface {
	Get(path string, v interface{}) error
	Post(path string, body interface{}) (*http.Response, error)
}

// Server is the live implementation of the TestTrack API client
type Server struct {
	url *url.URL
}

// New returns a live TestTrack for use in API calls
func New() (IServer, error) {
	urlString, ok := os.LookupEnv("TESTTRACK_CLI_URL")
	if !ok {
		return nil, errors.New("TESTTRACK_CLI_URL must be set")
	}

	url, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	return &Server{url: url}, nil
}

// Get makes an authenticated GET to the TestTrack API
func (s *Server) Get(path string, v interface{}) error {
	url, err := s.urlFor(path)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("got %d status code", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, &v)
	if err != nil {
		return err
	}

	return nil
}

// Post makes an authenticated POST to the TestTrack API
func (s *Server) Post(path string, body interface{}) (*http.Response, error) {
	url, err := s.urlFor(path)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.Post(url, "application/json", bytes.NewReader(bodyBytes))
}

// Note that this operates on a copy to avoid mutating *s.url
func (s Server) urlFor(path string) (string, error) {
	s.url.Path = strings.TrimRight(s.url.Path, "/")

	return strings.Join([]string{
		s.url.String(),
		path,
	}, "/"), nil
}
