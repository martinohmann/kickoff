package project

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/project"
)

var (
	bold            = color.New(color.Bold)
	highlightRegexp = regexp.MustCompile(`(\{\{[^{]+\}\}|\.skel$)`)
)

func (o *CreateOptions) printConfig(config *project.Config) error {
	bold.Fprint(o.Out, "Project configuration:\n\n")

	tw := cli.NewTableWriter(o.Out)
	tw.Append(bold.Sprint("Name"), color.CyanString(config.Name), bold.Sprint("Owner"), config.Owner)
	tw.Append(bold.Sprint("Directory"), color.CyanString(homedir.Collapse(o.ProjectDir)), bold.Sprint("Host"), config.Host)
	tw.Render()

	bold.Fprintln(o.Out, "\nSkeletons")
	fmt.Fprint(o.Out, color.CyanString(strings.Join(o.SkeletonNames, " ")), "\n\n")

	if len(config.Skeleton.Values) > 0 || len(config.Values) > 0 {
		values, err := yaml.Marshal(config.Skeleton.Values)
		if err != nil {
			return err
		}

		overrides, err := yaml.Marshal(config.Values)
		if err != nil {
			return err
		}

		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Skeleton values", "Value overrides")
		tw.Append(string(values), string(overrides))
		tw.Render()
	}

	if config.Gitignore != nil || config.License != nil {
		gitignore := "-"
		if config.Gitignore != nil {
			gitignore = strings.Join(config.Gitignore.Names, " ")
		}

		license := "-"
		if config.License != nil {
			license = config.License.Name
		}

		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("License", "Gitignore")
		tw.Append(license, gitignore)
		tw.Render()

		fmt.Fprintln(o.Out)
	}

	return nil
}

func (o *CreateOptions) printPlan(plan *project.Plan) {
	bold.Fprint(o.Out, "The following file operations will be performed:\n\n")

	tw := cli.NewTableWriter(o.Out)
	tw.SetTablePadding(" ")

	for _, op := range plan.Operations {
		var (
			status    string
			dirSuffix string
		)

		source := op.Source
		dest := op.Dest

		if source.Mode.IsDir() {
			dirSuffix = "/"
		}

		switch op.Type {
		case project.OpSkipUser:
			status = color.YellowString("! skip ") + color.HiBlackString("(user)")
		case project.OpSkipExisting:
			status = color.YellowString("! skip ") + color.HiBlackString("(exists)")
		case project.OpOverwrite:
			status = color.RedString("✓ overwrite")
		default:
			status = color.GreenString("✓ create")
		}

		origin := "<generated>"
		if ref := source.SkeletonRef; ref != nil {
			origin = ref.String()
		}

		tw.Append(
			color.CyanString(origin),
			color.HiBlackString("❯"),
			colorizePath(source.RelPath+dirSuffix),
			color.HiBlackString("=❯"),
			colorizePath(dest.RelPath()+dirSuffix),
			status,
		)
	}

	tw.Render()
	fmt.Fprintln(o.Out)
}

func (o *CreateOptions) printSummary(plan *project.Plan) {
	counts := plan.OpCounts

	fmt.Fprintf(o.Out, "%s Project %s created in %s. %s files created, %s skipped and %s overwritten\n",
		color.GreenString("✓"), bold.Sprint(o.ProjectName), bold.Sprint(homedir.Collapse(o.ProjectDir)),
		color.GreenString("%d", counts[project.OpCreate]),
		color.YellowString("%d", counts[project.OpSkipUser]+counts[project.OpSkipExisting]),
		color.RedString("%d", counts[project.OpOverwrite]),
	)
}

func colorizePath(path string) string {
	return highlightRegexp.ReplaceAllString(path, color.CyanString(`$1`))
}
