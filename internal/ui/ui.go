package ui

import "context"

// SelectItem represents a selectable option in Selector or Checkbox components.
type SelectItem struct {
	Label string
	Value string
	Desc  string
}

// WizardResult holds user selections from the init wizard.
type WizardResult struct {
	ProjectName string
	Language    string
	Framework   string
	Features    []string
	UserName    string
	ConvLang    string
}

// Selector provides fuzzy-search-enabled single selection.
type Selector interface {
	Select(label string, items []SelectItem) (string, error)
}

// Checkbox provides toggle-based multi-selection.
type Checkbox interface {
	MultiSelect(label string, items []SelectItem) ([]string, error)
}

// Prompt provides text input and confirmation prompts.
type Prompt interface {
	Input(label string, opts ...InputOption) (string, error)
	Confirm(label string, defaultVal bool) (bool, error)
}

// InputOption configures optional behavior for text input prompts.
type InputOption func(*inputConfig)

// inputConfig holds internal configuration for Input prompts.
type inputConfig struct {
	placeholder string
	validate    func(string) error
	defaultVal  string
}

// WithPlaceholder sets placeholder text displayed when the input is empty.
func WithPlaceholder(p string) InputOption {
	return func(c *inputConfig) {
		c.placeholder = p
	}
}

// WithValidation sets a validation function for the input value.
func WithValidation(fn func(string) error) InputOption {
	return func(c *inputConfig) {
		c.validate = fn
	}
}

// WithDefault sets a default value for the input.
func WithDefault(d string) InputOption {
	return func(c *inputConfig) {
		c.defaultVal = d
	}
}

// Wizard runs an interactive multi-step project initialization flow.
type Wizard interface {
	Run(ctx context.Context) (*WizardResult, error)
}

// Progress provides determinate progress bars and indeterminate spinners.
type Progress interface {
	Start(title string, total int) ProgressBar
	Spinner(title string) Spinner
}

// ProgressBar is a determinate progress indicator.
type ProgressBar interface {
	Increment(n int)
	SetTitle(title string)
	Done()
}

// Spinner is an indeterminate progress indicator.
type Spinner interface {
	SetTitle(title string)
	Stop()
}

// ErrCancelled is returned when the user cancels an interactive operation.
var ErrCancelled = errCancelled{}

type errCancelled struct{}

func (errCancelled) Error() string { return "operation cancelled" }

// ErrNoItems is returned when a Selector or Checkbox receives an empty item list.
var ErrNoItems = errNoItems{}

type errNoItems struct{}

func (errNoItems) Error() string { return "no items to select from" }

// ErrHeadlessNoDefaults is returned when headless mode runs without defaults.
var ErrHeadlessNoDefaults = errHeadlessNoDefaults{}

type errHeadlessNoDefaults struct{}

func (errHeadlessNoDefaults) Error() string {
	return "headless mode requires defaults for all wizard fields"
}
