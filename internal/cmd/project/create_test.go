package project

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/prompt"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	configPath := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		WithRepository("other", "../../testdata/repos/repo2").
		WithProjectOwner("hansdampf").
		Create()

	streams, _, _, _ := cli.NewTestIOStreams()

	f := cmdutil.NewFactoryWithConfigPath(streams, configPath)
	f.HTTPClient = func() *http.Client { return http.DefaultClient }

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.github.com/gitignore/templates",
		httpmock.NewStringResponder(200, `["go", "hugo"]`))
	httpmock.RegisterResponder("GET", "https://api.github.com/gitignore/templates/go",
		httpmock.NewStringResponder(200, `{"name":"go","source":"the-gitignore-template"}`))
	httpmock.RegisterResponder("GET", "https://api.github.com/licenses",
		httpmock.NewStringResponder(200, `[{"key":"mit","name":"MIT License"},{"key":"unlicense","name":"The Unlicense"}]`))
	httpmock.RegisterResponder("GET", "https://api.github.com/licenses/mit",
		httpmock.NewStringResponder(200, `{"key":"mit","name":"MIT License","body":"the-mit-license"}`))
	httpmock.RegisterResponder("GET", "https://api.github.com/licenses/unlicense",
		httpmock.NewStringResponder(200, `{"key":"unlicense","name":"The Unlicense","body":"the-unlicense"}`))

	t.Run("empty skeleton does not create project dir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "myproject")

		stubPrompt(f)

		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"myproject", "default:minimal", "-d", dir, "--owner", "johndoe"})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())
		require.NoDirExists(t, dir)
	})

	t.Run("creates project from skeleton", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "myproject")

		stubber, fakePrompt := stubPrompt(f)

		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{"myproject", "default:advanced", "-d", dir, "--owner", "johndoe"})
		cmd.SetOut(io.Discard)

		// confirm apply
		stubber.StubOne(true)

		require.NoError(t, cmd.Execute())

		fakePrompt.AssertExpectations(t)

		require.DirExists(t, dir)
	})

	t.Run("asks for required inputs", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "myproject")

		stubber, fakePrompt := stubPrompt(f)

		cmd := NewCreateCmd(f)
		cmd.SetOut(io.Discard)
		cmd.SetArgs([]string{"-d", dir})

		// skeleton names
		stubber.StubOne([]string{"default:advanced"})

		// project name
		stubber.StubOne("myproject")

		// confirm apply
		stubber.StubOne(true)

		require.NoError(t, cmd.Execute())

		fakePrompt.AssertExpectations(t)

		require.DirExists(t, dir)
		require.DirExists(t, filepath.Join(dir, ".git"))
		require.FileExists(t, filepath.Join(dir, "foobar", "somefile.yaml"))
	})

	t.Run("overwrite files if --overwrite flag is provided", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "myproject")

		stubber, fakePrompt := stubPrompt(f)

		// create a project
		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{
			"myproject", "default:advanced", "-d", dir,
			"--owner", "johndoe", "--license", "mit",
		})
		cmd.SetOut(io.Discard)

		// confirm apply
		stubber.StubOne(true)

		require.NoError(t, cmd.Execute())
		assertFileContains(t, filepath.Join(dir, "LICENSE"), "the-mit-license")

		// create a project in the same dir again but with different license, without passing overwrite
		cmd = NewCreateCmd(f)
		cmd.SetArgs([]string{
			"myproject", "default:advanced", "-d", dir,
			"--owner", "johndoe", "--license", "unlicense",
		})
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())
		assertFileContains(t, filepath.Join(dir, "LICENSE"), "the-mit-license")

		// create a project in the same dir again, this time overwrite all existing files
		cmd = NewCreateCmd(f)
		cmd.SetArgs([]string{
			"myproject", "default:advanced", "-d", dir,
			"--owner", "johndoe", "--license", "unlicense", "--overwrite",
		})
		cmd.SetOut(io.Discard)

		// confirm apply
		stubber.StubOne(true)

		require.NoError(t, cmd.Execute())
		assertFileContains(t, filepath.Join(dir, "LICENSE"), "the-unlicense")

		fakePrompt.AssertExpectations(t)
	})

	t.Run("interactive mode prompts for every config option", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "myproject")

		stubber, fakePrompt := stubPrompt(f)

		cmd := NewCreateCmd(f)
		cmd.SetArgs([]string{
			"myproject", "default:advanced",
			"--license", "unlicense", "--gitignore", "go",
			"--set", "filename=barbaz", "--init-git",
			"--skip-file", "optional-file",
			"--interactive",
		})
		cmd.SetOut(io.Discard)

		// skeleton name, project name and dir
		stubber.StubOneDefault()
		stubber.StubOneDefault()
		stubber.StubOne(dir)

		// project host, project owner
		stubber.StubOneDefault()
		stubber.StubOneDefault()

		// license, gitignore templates
		stubber.StubOneDefault()
		stubber.StubOneDefault()

		// git init
		stubber.StubOne(true)

		// edit values
		stubber.StubOne(true)
		stubber.StubOneDefault()

		// confirm apply
		stubber.StubOne(true)

		require.NoError(t, cmd.Execute())

		fakePrompt.AssertExpectations(t)

		require.DirExists(t, dir)
		require.DirExists(t, filepath.Join(dir, ".git"))
		require.FileExists(t, filepath.Join(dir, "LICENSE"))
		require.FileExists(t, filepath.Join(dir, ".gitignore"))
		require.FileExists(t, filepath.Join(dir, "barbaz", "somefile.yaml"))
		require.NoFileExists(t, filepath.Join(dir, "optional-file"))
	})
}

func stubPrompt(f *cmdutil.Factory) (*prompt.Stubber, *prompt.FakePrompt) {
	stubber, fakePrompt := prompt.NewStubber()
	f.Prompt = fakePrompt
	return stubber, fakePrompt
}

func assertFileContains(t *testing.T, path, expectedContent string) {
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(contents))
}
