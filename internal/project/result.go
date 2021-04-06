package project

import (
	"fmt"
	"io"
	"regexp"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
)

var (
	highlightRegexp = regexp.MustCompile(`(\{\{[^{]+\}\}|\.skel$)`)
	colorBold       = color.New(color.Bold)
)

// ActionType defines the action that should be performed for a given project
// file, template or directory.
type ActionType uint8

const (
	ActionTypeCreate ActionType = iota
	ActionTypeOverwrite
	ActionTypeSkipExisting
	ActionTypeSkipUser
)

// Action defines an action that should be performed on project creation.
type Action struct {
	Type        ActionType
	Source      Source
	Destination Destination
}

// Stats is a map from action type to the number of times the action was
// performed.
type Stats map[ActionType]int

// String implements fmt.Stringer.
//
// Returns a string which contains counts of the actions in s.
func (s Stats) String() string {
	return fmt.Sprintf(
		"Created: %d, Overwritten: %d, Skipped: %d",
		s[ActionTypeCreate], s[ActionTypeOverwrite],
		s[ActionTypeSkipUser]+s[ActionTypeSkipExisting],
	)
}

// Result holds stats about project creation and the actions that were
// performed.
type Result struct {
	Stats   Stats
	Actions []Action
}

func writeSummary(w io.Writer, result *Result) {
	tw := cli.NewTableWriter(w)
	tw.SetTablePadding(" ")

	for _, action := range result.Actions {
		var (
			actionSuffix string
			sourceType   string
			pathSuffix   string
		)

		source := action.Source
		dest := action.Destination

		switch {
		case source.IsTemplate():
			sourceType = "template"
		case source.Mode().IsDir():
			sourceType = "dir"
			pathSuffix = "/"
		default:
			sourceType = "file"
		}

		switch action.Type {
		case ActionTypeSkipUser:
			actionSuffix = color.CyanString("(skip)")
		case ActionTypeSkipExisting:
			actionSuffix = color.YellowString("(skip existing)")
		case ActionTypeOverwrite:
			actionSuffix = color.RedString("(overwrite)")
		}

		tw.Append(
			fmt.Sprintf("❯ %s", color.BlueString(sourceType)),
			colorizePath(source.Path()+pathSuffix),
			color.GreenString("=❯"),
			colorizePath(dest.RelPath()+pathSuffix),
			actionSuffix,
		)
	}

	tw.Render()
	fmt.Fprintln(w)
	colorBold.Fprintf(w, "Project creation complete. %s.\n", result.Stats.String())
}

func colorizePath(path string) string {
	return highlightRegexp.ReplaceAllString(path, color.YellowString(`$1`))
}
