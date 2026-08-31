// Package guardstate owns the guard-liveness STATE MODEL
// (SPEC-GUARD-STATE-MODEL-001, card t347): what firing expectations are
// declared, what vocabulary measures them, and — from M2 onward — how an entry
// is decided into exactly one classification.
//
// It is deliberately a separate package from guardliveness, which owns the
// SURFACING half and whose own doc states that "everything that produces a
// result" belongs here. The seam between them is the three-clause contract in
// guardliveness's package doc, consumed through its Producer interface.
//
// M1 ships this file's half only: the manifest schema, the closed
// measured-quantity vocabulary, and the census. Nothing here classifies.
package guardstate

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// ManifestPath is the manifest's home, relative to the repository root.
//
// It lives OUTSIDE .moai/config/ deliberately: `moai update` deletes that root
// wholesale (CleanMoaiManagedPaths), so a manifest kept there would be a
// liveness record that itself silently stops on the next update — the exact
// defect this SPEC exists to catch, one layer down (REQ-GSM-001).
const ManifestPath = ".moai/guards/liveness.yaml"

// Measure is the measured quantity an entry declares. The vocabulary is closed
// at exactly three values (REQ-GSM-003), and the three admit three different
// conclusion sets so that a single number is never asked to measure both
// whether a guard fired and whether a firing caught anything.
type Measure string

const (
	// MeasureFiredAtAll counts any run record, whatever its conclusion. It
	// answers "did this subject run?" and nothing else.
	MeasureFiredAtAll Measure = "fired-at-all"

	// MeasureFiredWithEffect excludes runs whose conclusion is `skipped` or
	// `cancelled`. It answers "did a run get as far as doing work?".
	MeasureFiredWithEffect Measure = "fired-with-effect"

	// MeasureVerdictRendered additionally requires a terminal `success` or
	// `failure`. It answers "did a run reach a verdict?" — and is the right
	// choice only where a subject that runs always reaches one.
	MeasureVerdictRendered Measure = "verdict-rendered"
)

// Measures is the closed vocabulary, in declaration order.
func Measures() []Measure {
	return []Measure{MeasureFiredAtAll, MeasureFiredWithEffect, MeasureVerdictRendered}
}

// Admits reports whether a run with the given conclusion qualifies under this
// measured quantity. An empty conclusion means the run reached no conclusion —
// still in progress, or reported null — which is what separates
// fired-with-effect from verdict-rendered.
func (m Measure) Admits(conclusion string) bool {
	switch m {
	case MeasureFiredAtAll:
		return true
	case MeasureFiredWithEffect:
		return conclusion != "skipped" && conclusion != "cancelled"
	case MeasureVerdictRendered:
		return conclusion == "success" || conclusion == "failure"
	default:
		return false
	}
}

// Valid reports whether the value is one of the three.
func (m Measure) Valid() bool {
	for _, known := range Measures() {
		if m == known {
			return true
		}
	}
	return false
}

// The named rejections. Each is named rather than folded into one
// "malformed entry" error because AC-GSM-002 (b) and AC-GSM-003 (b) require the
// rejection to be identifiable: an entry that is defaulted rather than refused
// converts a skipped-only subject into a green.
var (
	ErrMissingKind    = errors.New("guardstate: entry declares no kind")
	ErrMissingLocator = errors.New("guardstate: entry declares no locator")
	ErrMissingEvents  = errors.New("guardstate: entry declares no expected events")
	ErrMissingWindow  = errors.New("guardstate: entry declares no expectation window")
	ErrMissingMeasure = errors.New("guardstate: entry declares no measured quantity")
	ErrUnknownMeasure = errors.New("guardstate: measured quantity is outside the closed vocabulary")
	ErrInvalidWindow  = errors.New("guardstate: expectation window is not interpretable")
)

// Entry is one declared firing expectation.
//
// Every field is DATA rather than shape: `Kind` is a free string and `Locator`
// is a free string, so an entry of a second, non-workflow kind is accepted by
// this same struct and this same parser with no schema change (REQ-GSM-001,
// AC-GSM-005 (b)). Nothing here is named for a forge workflow.
type Entry struct {
	// Kind names what sort of subject this is. State-table row 1 decides on
	// it — an entry missing it is undecidable — which is why it is one of the
	// five required fields rather than an optional annotation.
	Kind string `yaml:"kind"`

	// Locator addresses the subject within its kind. For a workflow it is the
	// repository-relative path; the field carries no assumption beyond that it
	// identifies the subject to whatever reader its kind has.
	Locator string `yaml:"locator"`

	// Events are the triggering events under which firing is expected.
	Events []string `yaml:"events"`

	// Window is the expectation window, as `Nd` or `Nh`. Rows 3 and 4 read it.
	Window string `yaml:"window"`

	// Measure is the single measured quantity. Exactly one, per REQ-GSM-003.
	Measure Measure `yaml:"measure"`

	// ExpectedWhen declares the condition under which firing is expected at
	// all, for a subject legitimately quiet the rest of the time (REQ-GSM-004).
	// Empty means firing is expected unconditionally.
	//
	// It is optional and load-bearing at once: rows 5 and 6 are the two
	// branches of the test that reads it, so an entry that omits it where it
	// belongs does not leave a cell empty — it makes row 5 collapse into row 6
	// and reports a correctly-quiet subject as an anomaly on every sweep.
	ExpectedWhen string `yaml:"expected_when,omitempty"`
}

// IsConditional reports whether the entry declares a condition on its firing.
func (e Entry) IsConditional() bool { return e.ExpectedWhen != "" }

var windowPattern = regexp.MustCompile(`^([1-9][0-9]*)([dh])$`)

// WindowDuration interprets the declared window.
func (e Entry) WindowDuration() (time.Duration, error) {
	m := windowPattern.FindStringSubmatch(e.Window)
	if m == nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidWindow, e.Window)
	}
	var n int
	if _, err := fmt.Sscanf(m[1], "%d", &n); err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidWindow, e.Window)
	}
	unit := time.Hour
	if m[2] == "d" {
		unit = 24 * time.Hour
	}
	return time.Duration(n) * unit, nil
}

// Validate refuses an incomplete or out-of-vocabulary entry by name. It never
// defaults: a defaulted field is an expectation nobody declared, and the
// classifier downstream cannot tell it from one that was.
func (e Entry) Validate() error {
	if e.Kind == "" {
		return ErrMissingKind
	}
	if e.Locator == "" {
		return ErrMissingLocator
	}
	if len(e.Events) == 0 {
		return ErrMissingEvents
	}
	if e.Window == "" {
		return ErrMissingWindow
	}
	if _, err := e.WindowDuration(); err != nil {
		return err
	}
	if e.Measure == "" {
		return ErrMissingMeasure
	}
	if !e.Measure.Valid() {
		return fmt.Errorf("%w: %q", ErrUnknownMeasure, e.Measure)
	}
	return nil
}

// Manifest is the declared watched set.
type Manifest struct {
	Entries []Entry `yaml:"entries"`
}

// ParseManifest decodes and validates. Validation runs per entry here as well
// as on the constructor path, so a malformed entry cannot reach the classifier
// by arriving through the file instead of through code.
func ParseManifest(src []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(src, &m); err != nil {
		return nil, fmt.Errorf("guardstate: parse manifest: %w", err)
	}
	for _, e := range m.Entries {
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("guardstate: entry %q: %w", e.Locator, err)
		}
	}
	return &m, nil
}

// LoadManifest reads and parses the manifest at path.
func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("guardstate: read manifest: %w", err)
	}
	return ParseManifest(raw)
}
