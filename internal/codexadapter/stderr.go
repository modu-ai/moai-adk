package codexadapter

import "github.com/modu-ai/moai-adk/internal/hook"

// StderrClass names what a hook's stderr text means when it exits 2.
//
// The two classes were measured to behave differently, so an adapter that
// treated stderr uniformly would be wrong in one direction: PreToolUse stderr
// is shown to the model as the blocking reason, while Stop stderr is injected
// as an instruction and its text never appears in the event stream at all.
type StderrClass string

const (
	// StderrBlockingReason is surfaced to the model as why the call was refused.
	StderrBlockingReason StderrClass = "blocking-reason"

	// StderrContinuationPrompt is fed to the model as an instruction to keep
	// going. The text itself is not displayed.
	StderrContinuationPrompt StderrClass = "continuation-prompt"
)

// EvidenceTier records how a classification is known.
//
// It is carried per row so a declared-but-unobserved classification cannot be
// mistaken for a measured one at the point of use.
type EvidenceTier string

const (
	// EvidenceMeasured — observed in a recorded codex-cli run.
	EvidenceMeasured EvidenceTier = "measured"

	// EvidenceDeclared — present in the Codex binary's strings, never exercised.
	EvidenceDeclared EvidenceTier = "declared-not-measured"
)

// StderrRule is one event's stderr treatment.
type StderrRule struct {
	Class    StderrClass
	Evidence EvidenceTier
}

// stderrClasses covers only events this milestone adapts. Events excluded by
// SPEC §B carry no classification here: giving them one would assert a
// contract nothing measured or exercised.
var stderrClasses = map[hook.EventType]StderrRule{
	// Measured: the command never ran and the stderr text was surfaced to the
	// model verbatim ("blocked by a workspace hook: …").
	hook.EventPreToolUse: {StderrBlockingReason, EvidenceMeasured},

	// Measured: the turn continued and the model followed the stderr text,
	// which never appeared in the event stream.
	hook.EventStop: {StderrContinuationPrompt, EvidenceMeasured},

	// Declared only. The Codex binary carries a "did not write a blocking
	// reason to stderr" string for this event, but exit 2 was never exercised
	// on it.
	hook.EventUserPromptSubmit: {StderrBlockingReason, EvidenceDeclared},
}

// ClassifyStderr returns the stderr treatment for an event.
//
// The second return reports whether a classification exists at all; callers
// must not fall back to a default, since guessing between "shown as a reason"
// and "injected as a prompt" changes what the model receives.
func ClassifyStderr(event hook.EventType) (StderrRule, bool) {
	rule, ok := stderrClasses[event]
	return rule, ok
}
