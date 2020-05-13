// Package gitignore provides an interface to gitignore.io to fetch gitignore
// templates. These templates are used to optionally populate the .gitignore
// file of a new project.
package gitignore

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apex/log"
)

var (
	// apiBaseURL is a variable and not a constant so it can be replaced in
	// tests.
	apiBaseURL = "https://gitignore.io/api"

	// ErrNotFound is returned if a gitignore template could not be found.
	ErrNotFound = errors.New("gitignore not found")
)

// Get fetches the gitignore template for query from gitignore.io. The query
// can be a comma-separated list of gitignore templates (e.g. "go,python")
// which are combined into a single gitignore template. Will return an error if
// the http connection fails or if the response status code is not 200. Will
// return ErrNotFound if any of the requested gitignore templates cannot be
// found.
func Get(ctx context.Context, query string) (string, error) {
	log.WithField("query", query).Debug("fetching template from gitignore.io")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", apiBaseURL, query), nil)
	if err != nil {
		return "", err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", ErrNotFound
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("gitignore.io returned status code %d while fetching gitignore template", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	return strings.TrimSpace(string(body)), err
}

// List obtains a list of available gitignore templates from gitignore.io. Will
// return an error if the http connection fails or the response status code is
// not 200.
func List(ctx context.Context) ([]string, error) {
	log.Debug("fetching template list from gitignore.io")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/list", apiBaseURL), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gitignore.io returned status code %d while listing gitignore templates", resp.StatusCode)
	}

	gitignores := make([]string, 0)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		gitignores = append(gitignores, strings.Split(scanner.Text(), ",")...)
	}

	return gitignores, nil
}
