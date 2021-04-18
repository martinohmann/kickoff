package cmdutil

import (
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCompletion(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../testdata/repos/repo1").
		WithRepository("other", "../testdata/repos/repo2").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, _, _ := cli.NewTestIOStreams()

	f := NewFactoryWithConfigPath(streams, configFile.Name())
	f.HTTPClient = func() *http.Client { return http.DefaultClient }

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://www.toptal.com/developers/gitignore/api/list",
		httpmock.NewStringResponder(200, "hugo\ngo"))

	assert.Equal(t, []string{"go", "hugo"}, GitignoreNames(f))

	httpmock.RegisterResponder("GET", "https://api.github.com/licenses",
		httpmock.NewStringResponder(200, `[{"key":"mit","name":"MIT License"},{"key":"unlicense","name":"The Unlicense"}]`))

	assert.Equal(t, []string{"mit", "unlicense"}, LicenseNames(f))

	assert.Equal(t, []string{"default", "other"}, RepositoryNames(f))
	assert.Equal(t, []string{"default:advanced", "default:minimal", "other:minimal"}, SkeletonNames(f))
	assert.Equal(t, []string{"README.md.skel", "optional-file.skel", "{{.Values.filename}}/somefile.yaml"}, SkeletonFilenames(f, "default:advanced"))
	assert.Equal(t, []string{"panic", "fatal", "error", "warning", "info", "debug", "trace"}, LogLevelNames())
}
