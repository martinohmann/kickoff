package project

import (
	"fmt"
	"io"
	"regexp"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
)

var highlightRegexp = regexp.MustCompile(`(\{\{[^{]+\}\}|\.skel$)`)

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

// Logger logs project actions.
type Logger interface {
	// Log produces a log entry for given action.
	Log(action Action)
	// Stats returns the stats for actions that were passed to Log.
	Stats() Stats
	// Flush flushes all log entries to the underlying writer.
	Flush()
}

type logger struct {
	w     io.Writer
	tw    cli.TableWriter
	stats Stats
}

// NewLogger creates a new Logger which flushes formatted log entries to w.
func NewLogger(w io.Writer) Logger {
	tw := cli.NewTableWriter(w)
	tw.SetTablePadding(" ")

	return &logger{
		w:     w,
		tw:    tw,
		stats: make(Stats),
	}
}

func (l *logger) Log(action Action) {
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

	l.tw.Append(
		fmt.Sprintf("❯ %s", color.BlueString(sourceType)),
		colorizePath(source.Path()+pathSuffix),
		color.GreenString("=❯"),
		colorizePath(dest.RelPath()+pathSuffix),
		actionSuffix,
	)

	l.stats[action.Type]++
}

func (l *logger) Flush() {
	l.tw.Render()
	fmt.Fprintln(l.w)
}

func (l *logger) Stats() Stats {
	return l.stats
}

func colorizePath(path string) string {
	return highlightRegexp.ReplaceAllString(path, color.YellowString(`$1`))
}
