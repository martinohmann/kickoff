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

func TestAdapter_Get(t *testing.T) {
	svc := &mockService{}
	adapter := NewAdapter(svc)

	license := &github.License{
		Key:  github.String("foo"),
		Name: github.String("Foo License"),
		Body: github.String("The Foo License Text"),
	}

	svc.On("Get", mock.Anything, "foo").Return(license, &github.Response{}, nil)

	info, err := adapter.Get("foo")
	require.NoError(t, err)

	expected := &Info{
		Key:  "foo",
		Name: "Foo License",
		Body: "The Foo License Text",
	}

	assert.Equal(t, expected, info)
}

func TestAdapter_Get_NotFound(t *testing.T) {
	svc := &mockService{}
	adapter := NewAdapter(svc)

	svc.On("Get", mock.Anything, "foo").Return(nil, &github.Response{}, &github.ErrorResponse{
		Response: &http.Response{StatusCode: 404},
	})

	_, err := adapter.Get("foo")
	require.Error(t, err)
	assert.Equal(t, ErrLicenseNotFound, err)
}

func TestAdapter_List(t *testing.T) {
	svc := &mockService{}
	adapter := NewAdapter(svc)

	licenses := []*github.License{
		{
			Key:  github.String("foo"),
			Name: github.String("Foo License"),
			Body: github.String("The Foo License Text"),
		},
	}

	svc.On("List", mock.Anything).Return(licenses, &github.Response{}, nil)

	infos, err := adapter.List()
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
