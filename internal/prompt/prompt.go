package prompt

import (
	"reflect"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/stretchr/testify/mock"
)

// Prompt can be used to request user input.
type Prompt interface {
	AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error
}

type prompt struct {
	opts []survey.AskOpt
}

// New returns a prompt which just directly calls survey.AskOne. The provided
// opts are prepended to options passed to the AskOne method.
func New(opts ...survey.AskOpt) Prompt {
	return &prompt{opts: opts}
}

// AskOne implements Prompt.
func (pr *prompt) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return survey.AskOne(p, response, append(pr.opts, opts...)...)
}

// FakePrompt is a mock implementation of a Prompt.
type FakePrompt struct {
	mock.Mock
}

// AskOne implements Prompt.
func (f *FakePrompt) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	args := f.Called(p, response, opts)
	return args.Error(0)
}

// Stubber can stub prompts.
type Stubber struct {
	p *FakePrompt
}

// NewStubber creates a new *Stubber which can stub out calls to returned
// *FakePrompt's AskOne method.
func NewStubber() (*Stubber, *FakePrompt) {
	s := &Stubber{p: &FakePrompt{}}
	return s, s.p
}

// StubOne makes the next prompt return value.
func (s *Stubber) StubOne(value interface{}) {
	s.p.On("AskOne", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			_ = core.WriteAnswer(args.Get(1), "", value)
		}).
		Return(nil).
		Once()
}

// StubOneDefault makes the next prompt return its default value.
func (s *Stubber) StubOneDefault() {
	s.p.On("AskOne", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			fieldValue := reflect.ValueOf(args.Get(0)).Elem().FieldByName("Default")
			defaultValue := fieldValue.Interface()
			_ = core.WriteAnswer(args.Get(1), "", defaultValue)
		}).
		Return(nil).
		Once()
}

// StubOneError makes the next prompt return err.
func (s *Stubber) StubOneError(err error) {
	s.p.On("AskOne", mock.Anything, mock.Anything, mock.Anything).
		Return(err).
		Once()
}

// StubOne makes the next prompts return the provided values.
func (s *Stubber) StubMany(values ...interface{}) {
	for _, value := range values {
		s.StubOne(value)
	}
}
