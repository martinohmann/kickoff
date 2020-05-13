package skeleton

// fileTemplates is a mapping between filenames and the contents for these
// files when generating a new skeleton.
var fileTemplates = map[string]string{
	ConfigFileName: `---
# Refer to the .kickoff.yaml documentation at https://kickoff.run/skeletons/configuration
# for a complete list of available skeleton configuration options.
#
# ---
# description: |
#   Some optional description of the skeleton that might be helpful to users.
# values:
#   myVar: 'myValue'
#   other:
#     someVar: false
`,
	"README.md.skel": `# {{.Project.Name}}

{{ if .License -}}
![GitHub](https://img.shields.io/github/license/{{.Project.Owner}}/{{.Project.Name}}?color=orange)

## License

The source code of {{.Project.Name}} is released under the {{.License.Name}}. See the bundled
LICENSE file for details.
{{- end }}
`,
}
