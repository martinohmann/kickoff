package project

import (
	"fmt"
	"io"
	"regexp"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/kickoff"
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
	Source      kickoff.File
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
		"%s created %s skipped %s overwritten",
		color.GreenString("%d", s[ActionTypeCreate]),
		color.YellowString("%d", s[ActionTypeSkipUser]+s[ActionTypeSkipExisting]),
		color.RedString("%d", s[ActionTypeOverwrite]),
	)
}

// Result holds stats about project creation and the actions that were
// performed.
type Result struct {
	Stats   Stats
	Actions []Action
}

func (p *Project) writeSummary(w io.Writer) {
	tw := cli.NewTableWriter(w)
	tw.SetTablePadding(" ")

	for _, action := range p.result.Actions {
		var (
			status    string
			dirSuffix string
		)

		source := action.Source
		dest := action.Destination

		if source.Mode().IsDir() {
			dirSuffix = "/"
		}

		switch action.Type {
		case ActionTypeSkipUser:
			status = color.YellowString("! skip ") + color.HiBlackString("(user)")
		case ActionTypeSkipExisting:
			status = color.YellowString("! skip ") + color.HiBlackString("(exists)")
		case ActionTypeOverwrite:
			status = color.RedString("✓ overwrite")
		default:
			status = color.GreenString("✓ create")
		}

		var destPath string

		if source.Path() != dest.RelPath() {
			destPath = fmt.Sprintf(
				"%s %s",
				color.HiBlackString("=❯"),
				colorizePath(dest.RelPath()+dirSuffix),
			)
		}

		tw.Append(
			color.HiBlackString("❯"),
			colorizePath(source.Path()+dirSuffix),
			destPath,
			status,
		)
	}

	tw.Render()
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Project %s created in %s\n", colorBold.Sprint(p.name), colorBold.Sprint(p.targetDir))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "File statistics: %s\n", p.result.Stats)
}

func colorizePath(path string) string {
	return highlightRegexp.ReplaceAllString(path, color.CyanString(`$1`))
}
