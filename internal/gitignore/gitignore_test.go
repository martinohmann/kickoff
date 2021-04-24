package gitignore

import (
	"context"
	"errors"
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

func (s *mockService) Get(ctx context.Context, name string) (gitignore *github.Gitignore, resp *github.Response, err error) {
	args := s.Called(ctx, name)
	if args.Get(0) != nil {
		gitignore = args.Get(0).(*github.Gitignore)
	}
	if args.Get(1) != nil {
		resp = args.Get(1).(*github.Response)
	}
	return gitignore, resp, args.Error(2)
}

func (s *mockService) List(ctx context.Context) (gitignores []string, resp *github.Response, err error) {
	args := s.Called(ctx)
	if args.Get(0) != nil {
		gitignores = args.Get(0).([]string)
	}
	if args.Get(1) != nil {
		resp = args.Get(1).(*github.Response)
	}
	return gitignores, resp, args.Error(2)
}

func TestClient_GetTemplate(t *testing.T) {
	t.Run("fetches template", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		gitignore := &github.Gitignore{
			Name:   github.String("Go"),
			Source: github.String("the-source\n"),
		}

		svc.On("List", mock.Anything).Return([]string{"Go"}, &github.Response{}, nil)
		svc.On("Get", mock.Anything, "Go").Return(gitignore, &github.Response{}, nil)

		template, err := client.GetTemplate(context.Background(), "Go")
		require.NoError(t, err)

		expected := &Template{
			Query:   "Go",
			Names:   []string{"Go"},
			Content: []byte("### Go ###\nthe-source\n"),
		}

		assert.Equal(t, expected, template)
	})

	t.Run("fetches all templates from comma separated query", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		svc.On("List", mock.Anything).Return([]string{"Go", "Java", "Python"}, &github.Response{}, nil)
		svc.On("Get", mock.Anything, "Go").Return(&github.Gitignore{
			Name:   github.String("Go"),
			Source: github.String("go-source\n\n"),
		}, &github.Response{}, nil)
		svc.On("Get", mock.Anything, "Python").Return(&github.Gitignore{
			Name:   github.String("Python"),
			Source: github.String("\npython-source\n"),
		}, &github.Response{}, nil)

		template, err := client.GetTemplate(context.Background(), "Go,python")
		require.NoError(t, err)

		expected := &Template{
			Query:   "Go,python",
			Names:   []string{"Go", "Python"},
			Content: []byte("### Go ###\ngo-source\n\n### Python ###\npython-source\n"),
		}

		assert.Equal(t, expected, template)
	})

	t.Run("normalizes names before fetching templates", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		svc.On("List", mock.Anything).Return([]string{"Go", "Java", "Python"}, &github.Response{}, nil)

		_, err := client.GetTemplate(context.Background(), "Go,nonexistent,python")
		require.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("empty query returns 404 without making request", func(t *testing.T) {
		client := &Client{}

		_, err := client.GetTemplate(context.Background(), " ")
		require.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("error on initial list", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		svc.On("List", mock.Anything).Return(nil, &github.Response{}, errors.New("whoops"))

		_, err := client.GetTemplate(context.Background(), "go")
		require.EqualError(t, err, "whoops")
	})

	t.Run("converts github 404 error response", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		svc.On("List", mock.Anything).Return([]string{"foo"}, &github.Response{}, nil)
		svc.On("Get", mock.Anything, "foo").Return(nil, &github.Response{}, &github.ErrorResponse{
			Response: &http.Response{StatusCode: 404},
		})

		_, err := client.GetTemplate(context.Background(), "foo")
		require.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("error on get", func(t *testing.T) {
		svc := &mockService{}
		client := &Client{svc}

		svc.On("List", mock.Anything).Return([]string{"foo"}, &github.Response{}, nil)
		svc.On("Get", mock.Anything, "foo").Return(nil, &github.Response{}, errors.New("whoops"))

		_, err := client.GetTemplate(context.Background(), "foo")
		require.EqualError(t, err, "whoops")
	})
}
