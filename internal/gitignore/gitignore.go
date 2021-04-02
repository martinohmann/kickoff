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

	log "github.com/sirupsen/logrus"
)

const defaultBaseURL = "https://www.toptal.com/developers/gitignore/api"

// ErrNotFound is returned if a gitignore template could not be found.
var ErrNotFound = errors.New("gitignore template not found")

// Client can fetch gitignore templates.
type Client struct {
	*http.Client
	BaseURL string
}

// NewClient creates a new *Client which will use httpClient for making http
// requests. If httpClient is nil, http.DefaultClient will be used instead.
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		Client:  httpClient,
		BaseURL: defaultBaseURL,
	}
}

// GetTemplate fetches the gitignore template for query. The query can be a
// comma-separated list of gitignore templates (e.g. "go,python") which are
// combined into a single gitignore template. Will return an error if the
// http connection fails or if the response status code is not 200. Will
// return ErrNotFound if any of the requested gitignore templates cannot be
// found.
func (c *Client) GetTemplate(ctx context.Context, query string) (string, error) {
	log.WithField("query", query).Debug("fetching gitignore template")

	req, err := http.NewRequest("GET", c.buildRequestURL(fmt.Sprintf("/%s", query)), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", ErrNotFound
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("received status code %d while fetching gitignore template", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	return strings.TrimSpace(string(body)), err
}

// ListTemplates obtains a list of available gitignore templates. Will
// return an error if the http connection fails or the response status code
// is not 200.
func (c *Client) ListTemplates(ctx context.Context) ([]string, error) {
	log.Debug("fetching gitignore template list")

	req, err := http.NewRequest("GET", c.buildRequestURL("/list"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d while listing gitignore templates", resp.StatusCode)
	}

	gitignores := make([]string, 0)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		gitignores = append(gitignores, strings.Split(scanner.Text(), ",")...)
	}

	return gitignores, nil
}

func (c *Client) buildRequestURL(path string) string {
	return fmt.Sprintf("%s%s", c.BaseURL, path)
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	httpClient := c.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return httpClient.Do(req)
}
