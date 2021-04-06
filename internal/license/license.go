// Package license provides an adapter to fetch license texts from the GitHub
// Licenses API.
package license

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v28/github"
	log "github.com/sirupsen/logrus"
)

// ErrNotFound is returned by the Adapter if a license cannot be
// found via the GitHub Licenses API.
var ErrNotFound = errors.New("license not found")

// Info holds information about a single license.
type Info struct {
	Key  string
	Name string
	Body string
}

func toInfo(license *github.License) *Info {
	if license == nil {
		return nil
	}

	info := Info{}

	if license.Key != nil {
		info.Key = *license.Key
	}

	if license.Name != nil {
		info.Name = *license.Name
	}

	if license.Body != nil {
		info.Body = *license.Body
	}

	return &info
}

// GitHubLicensesService is the interface of the GitHub Licenses API Service.
type GitHubLicensesService interface {
	Get(ctx context.Context, licenseName string) (*github.License, *github.Response, error)
	List(ctx context.Context) ([]*github.License, *github.Response, error)
}

// Client fetches license information from the GitHub Licenses API Service.
type Client struct {
	GitHubLicensesService
}

// NewClient creates a new *Client which will use httpClient for making http
// requests. If httpClient is nil, http.DefaultClient will be used instead.
func NewClient(httpClient *http.Client) *Client {
	githubClient := github.NewClient(httpClient)

	return &Client{githubClient.Licenses}
}

// GetLicense fetches the info for the license with name. Will return
// ErrNotFound if the license is not recognized.
func (c *Client) GetLicense(ctx context.Context, name string) (*Info, error) {
	log.WithField("license", name).Debugf("fetching license info")

	license, _, err := c.Get(ctx, name)
	if err != nil {
		var errResp *github.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response.StatusCode == 404 {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return toInfo(license), nil
}

// ListLicenses lists the infos for all available licenses. These do not
// include the license body but only the metadata. Use Get to fetch the
// body of a particular license.
func (c *Client) ListLicenses(ctx context.Context) ([]*Info, error) {
	log.Debug("fetching license infos")

	licenses, _, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]*Info, len(licenses))
	for i, license := range licenses {
		infos[i] = toInfo(license)
	}

	return infos, nil
}
