package cmdutil

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
)

// LongDesc formats the long description for a *cobra.Command.
func LongDesc(s string) string {
	return trim(heredoc.Doc(s))
}

// Examples formats examples for a *cobra.Command. Each line is indented by 2
// spaces so that it aligns nicely with usage and flags.
func Examples(s string) string {
	return indent(trim(s), "  ")
}

func indent(s, indent string) string {
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		lines[i] = indent + trim(line)
	}

	return strings.Join(lines, "\n")
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
