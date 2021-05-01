package prompt

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/editor"
)

type Editor struct {
	survey.Renderer
	Message         string
	Default         string
	Help            string
	Command         string
	HideDefault     bool
	AppendDefault   bool
	FilenamePattern string
}

type EditorTemplateData struct {
	Editor
	Answer     string
	ShowAnswer bool
	ShowHelp   bool
	Config     *survey.PromptConfig
}

var EditorQuestionTemplate = survey.EditorQuestionTemplate

// PromptAgain implements survey.PromptAgainer.
func (e *Editor) PromptAgain(config *survey.PromptConfig, invalid interface{}, err error) (interface{}, error) {
	initialValue := invalid.(string)
	return e.prompt(initialValue, config)
}

// Prompt implements survey.Prompt.
func (e *Editor) Prompt(config *survey.PromptConfig) (interface{}, error) {
	initialValue := ""
	if e.Default != "" && e.AppendDefault {
		initialValue = e.Default
	}
	return e.prompt(initialValue, config)
}

func (e *Editor) prompt(initialValue string, config *survey.PromptConfig) (interface{}, error) {
	data := EditorTemplateData{
		Editor: *e,
		Config: config,
	}

	if err := e.Render(EditorQuestionTemplate, data); err != nil {
		return "", err
	}

	rr := e.NewRuneReader()
	rr.SetTermMode()
	defer rr.RestoreTermMode()

	cursor := e.NewCursor()
	cursor.Hide()
	defer cursor.Show()

	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			return "", err
		}

		if r == '\r' || r == '\n' {
			break
		}

		if r == terminal.KeyInterrupt {
			return "", terminal.InterruptErr
		}

		if r == terminal.KeyEndTransmission {
			break
		}

		if string(r) == config.HelpInput && e.Help != "" {
			data := EditorTemplateData{
				Editor:   *e,
				ShowHelp: true,
				Config:   config,
			}

			if err := e.Render(EditorQuestionTemplate, data); err != nil {
				return "", err
			}
		}
		continue
	}

	stdio := e.Stdio()

	editor := &editor.Editor{
		IOStreams: cli.IOStreams{
			In:     stdio.In,
			Out:    stdio.Out,
			ErrOut: stdio.Err,
		},
		Command: e.Command,
	}

	edited, err := editor.Edit([]byte(initialValue), e.FilenamePattern)
	if err != nil {
		cursor.Show()
		return "", err
	}

	if len(edited) == 0 && !e.AppendDefault {
		return e.Default, nil
	}

	return string(edited), nil
}

// Cleanup implements survey.Prompt.
func (e *Editor) Cleanup(config *survey.PromptConfig, val interface{}) error {
	data := EditorTemplateData{
		Editor:     *e,
		Answer:     "<Received>",
		ShowAnswer: true,
		Config:     config,
	}

	return e.Render(EditorQuestionTemplate, data)
}
