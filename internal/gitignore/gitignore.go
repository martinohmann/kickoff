package gitignore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v28/github"
	log "github.com/sirupsen/logrus"
)

// NotFoundError is returned if a gitignore template could not be found.
type NotFoundError string

func (e NotFoundError) Error() string {
	return fmt.Sprintf("gitignore template %q not found", string(e))
}

// Template holds the content of a gitignore template and the query that was
// used to obtain it.
type Template struct {
	Query   string
	Names   []string
	Content []byte
}

// GitHubGitignoresService is the interface of the GitHub Gitignores API Service.
type GitHubGitignoresService interface {
	Get(ctx context.Context, name string) (*github.Gitignore, *github.Response, error)
	List(ctx context.Context) ([]string, *github.Response, error)
}

// Client can fetch gitignore templates.
type Client struct {
	GitHubGitignoresService
}

// NewClient creates a new *Client which will use httpClient for making http
// requests. If httpClient is nil, http.DefaultClient will be used instead.
func NewClient(httpClient *http.Client) *Client {
	githubClient := github.NewClient(httpClient)

	return &Client{githubClient.Gitignores}
}

// GetTemplate fetches the gitignore template for query. The query can be a
// comma-separated list of gitignore templates (e.g. "go,python") which are
// combined into a single gitignore template. Will return an error if the http
// connection fails or if the response status code is not 200. Will return
// ErrNotFound if any of the requested gitignore templates cannot be found.
func (c *Client) GetTemplate(ctx context.Context, query string) (*Template, error) {
	log.WithField("query", query).Debug("fetching gitignore template")

	query = strings.TrimSpace(query)

	if query == "" {
		return nil, NotFoundError("")
	}

	gitignores, err := c.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	names := strings.Split(query, ",")

	normalizedNames := make([]string, 0, len(names))
	for _, name := range names {
		normalized, ok := matchCaseInsensitive(gitignores, strings.TrimSpace(name))
		if !ok {
			return nil, NotFoundError(name)
		}

		normalizedNames = append(normalizedNames, normalized)
	}

	buf := new(bytes.Buffer)

	for i, name := range normalizedNames {
		gitignore, _, err := c.Get(ctx, name)
		if err != nil {
			var errResp *github.ErrorResponse
			if errors.As(err, &errResp) && errResp.Response.StatusCode == 404 {
				return nil, NotFoundError(name)
			}

			return nil, err
		}

		source := strings.TrimSpace(gitignore.GetSource())

		buf.WriteString("### ")
		buf.WriteString(gitignore.GetName())
		buf.WriteString(" ###\n")
		buf.WriteString(source)
		buf.WriteRune('\n')

		if i < len(normalizedNames)-1 {
			buf.WriteRune('\n')
		}
	}

	t := &Template{
		Query:   query,
		Names:   normalizedNames,
		Content: buf.Bytes(),
	}

	return t, nil
}

// ListTemplates obtains a list of available gitignore templates.
func (c *Client) ListTemplates(ctx context.Context) ([]string, error) {
	log.Debug("fetching gitignore template list")

	gitignores, _, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	return gitignores, nil
}

func matchCaseInsensitive(haystack []string, needle string) (string, bool) {
	needle = strings.ToLower(needle)

	for _, item := range haystack {
		if strings.ToLower(item) == needle {
			return item, true
		}
	}

	return "", false
}
