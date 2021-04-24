package license

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v28/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	mock.Mock
}

func (s *mockService) Get(ctx context.Context, licenseName string) (license *github.License, resp *github.Response, err error) {
	args := s.Called(ctx, licenseName)
	if args.Get(0) != nil {
		license = args.Get(0).(*github.License)
	}
	if args.Get(1) != nil {
		resp = args.Get(1).(*github.Response)
	}
	return license, resp, args.Error(2)
}

func (s *mockService) List(ctx context.Context) (licenses []*github.License, resp *github.Response, err error) {
	args := s.Called(ctx)
	if args.Get(0) != nil {
		licenses = args.Get(0).([]*github.License)
	}
	if args.Get(1) != nil {
		resp = args.Get(1).(*github.Response)
	}
	return licenses, resp, args.Error(2)
}

func TestClient_GetLicense(t *testing.T) {
	svc := &mockService{}
	client := &Client{svc}

	license := &github.License{
		Key:  github.String("foo"),
		Name: github.String("Foo License"),
		Body: github.String("The Foo License Text"),
	}

	svc.On("Get", mock.Anything, "foo").Return(license, &github.Response{}, nil)

	info, err := client.GetLicense(context.Background(), "foo")
	require.NoError(t, err)

	expected := &Info{
		Key:  "foo",
		Name: "Foo License",
		Body: "The Foo License Text",
	}

	assert.Equal(t, expected, info)
}

func TestClient_GetLicense_NotFound(t *testing.T) {
	svc := &mockService{}
	client := &Client{svc}

	svc.On("Get", mock.Anything, "foo").Return(nil, &github.Response{}, &github.ErrorResponse{
		Response: &http.Response{StatusCode: 404},
	})

	_, err := client.GetLicense(context.Background(), "foo")
	require.EqualError(t, err, NotFoundError("foo").Error())
}

func TestClient_ListLicenses(t *testing.T) {
	svc := &mockService{}
	client := &Client{svc}

	licenses := []*github.License{
		{
			Key:  github.String("foo"),
			Name: github.String("Foo License"),
			Body: github.String("The Foo License Text"),
		},
	}

	svc.On("List", mock.Anything).Return(licenses, &github.Response{}, nil)

	infos, err := client.ListLicenses(context.Background())
	require.NoError(t, err)

	expected := []*Info{
		{
			Key:  "foo",
			Name: "Foo License",
			Body: "The Foo License Text",
		},
	}

	assert.Equal(t, expected, infos)
}
