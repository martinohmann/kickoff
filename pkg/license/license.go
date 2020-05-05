// Package license provides an adapter to fetch license texts from the GitHub
// Licenses API.
package license

import (
	"context"
	"errors"

	"github.com/apex/log"
	"github.com/google/go-github/v28/github"
)

var (
	// ErrNotFound is returned by the Adapter if a license cannot be
	// found via the GitHub Licenses API.
	ErrNotFound = errors.New("license not found")

	// DefaultAdapter is the default adapter for the GitHub Licenses API.
	DefaultAdapter = NewAdapter(github.NewClient(nil).Licenses)
)

// GitHubLicensesService is the interface of the GitHub Licenses API Service.
type GitHubLicensesService interface {
	Get(ctx context.Context, licenseName string) (*github.License, *github.Response, error)
	List(ctx context.Context) ([]*github.License, *github.Response, error)
}

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

// Adapter adapts to the GitHub Licences API.
type Adapter struct {
	service GitHubLicensesService
}

// NewAdapter creates a new *Adapter for service.
func NewAdapter(service GitHubLicensesService) *Adapter {
	return &Adapter{
		service: service,
	}
}

// Get fetches the info for the license with name. Will return
// ErrNotFound if the license is not recognized.
func (f *Adapter) Get(ctx context.Context, name string) (*Info, error) {
	log.WithField("license", name).Debugf("fetching license info from GitHub")

	license, _, err := f.service.Get(ctx, name)
	if err != nil {
		errResp, ok := err.(*github.ErrorResponse)
		if ok && errResp.Response.StatusCode == 404 {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return toInfo(license), nil
}

// List lists the infos for all available licenses. These do not include the
// license body but only the metadata. Use Get to fetch the body of a
// particular license.
func (f *Adapter) List(ctx context.Context) ([]*Info, error) {
	log.Debug("fetching license infos from GitHub")

	licenses, _, err := f.service.List(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]*Info, len(licenses))
	for i, license := range licenses {
		infos[i] = toInfo(license)
	}

	return infos, nil
}

// Get fetches the info for the license with name using the DefaultAdapter.
// Will return ErrNotFound if the license is not recognized.
func Get(ctx context.Context, name string) (*Info, error) {
	return DefaultAdapter.Get(ctx, name)
}

// List lists the infos for all available licenses using the DefaultAdapter.
// These do not include the license body but only the metadata. Use Get to
// fetch the body of a particular license.
func List(ctx context.Context) ([]*Info, error) {
	return DefaultAdapter.List(ctx)
}
