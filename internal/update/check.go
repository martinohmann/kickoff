package update

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-github/v28/github"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"golang.org/x/term"
)

// Info describes an update.
type Info struct {
	IsUpdate         bool
	CurrentVersion   string
	LatestVersion    string
	LatestReleaseURL string
}

// Check checks for updates if the last check was older than checkInterval. The
// last update check time is tracked in the file at statePath. Returns nil if
// no check was performed.
func Check(ctx context.Context, statePath, currentVersion string, checkInterval time.Duration) (*Info, error) {
	if !shouldCheck(statePath, checkInterval) {
		return nil, nil
	}

	info, err := check(ctx, currentVersion)
	if err != nil {
		return nil, err
	}

	if err := writeState(statePath); err != nil {
		return nil, err
	}

	return info, nil
}

func shouldCheck(statePath string, checkInterval time.Duration) bool {
	if os.Getenv(kickoff.EnvKeyNoUpdateCheck) != "" {
		return false
	}

	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}

	state, err := getState(statePath)
	if err != nil {
		return true
	}

	return time.Since(state.CheckedAt) > checkInterval
}

func check(ctx context.Context, current string) (*Info, error) {
	client := github.NewClient(http.DefaultClient)

	release, _, err := client.Repositories.GetLatestRelease(ctx, "martinohmann", "kickoff")
	if err != nil {
		return nil, err
	}

	latest := release.GetTagName()

	currentVersion, err := semver.NewVersion(current)
	if err != nil {
		return nil, err
	}

	latestVersion, err := semver.NewVersion(latest)
	if err != nil {
		return nil, err
	}

	info := &Info{
		IsUpdate:         currentVersion.LessThan(latestVersion),
		CurrentVersion:   current,
		LatestVersion:    latest,
		LatestReleaseURL: release.GetHTMLURL(),
	}

	return info, nil
}

type updateState struct {
	CheckedAt time.Time `json:"checkedAt"`
}

func getState(path string) (*updateState, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state updateState

	if err := json.Unmarshal(buf, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func writeState(path string) error {
	state := &updateState{CheckedAt: time.Now()}

	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf, 0644)
}
