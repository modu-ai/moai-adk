package codexadapter

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestPreToolUseStderrIsBlockingReason — AC-REQ-4a.
//
// Baseline: probe/run-exit2.jsonl, where the command never ran and the model
// reported "blocked by a workspace hook: T83-BLOCKED-BY-EXIT2 …".
func TestPreToolUseStderrIsBlockingReason(t *testing.T) {
	t.Parallel()

	rule, ok := ClassifyStderr(hook.EventPreToolUse)
	if !ok {
		t.Fatal("PreToolUse has no stderr classification")
	}
	if rule.Class != StderrBlockingReason {
		t.Errorf("class = %s, want %s", rule.Class, StderrBlockingReason)
	}
	if rule.Evidence != EvidenceMeasured {
		t.Errorf("evidence = %s, want %s", rule.Evidence, EvidenceMeasured)
	}
}

// TestStopStderrIsContinuationPrompt — AC-REQ-4b.
//
// Baseline: probe/run-stop-exit2.jsonl, where the turn continued and the model
// obeyed the stderr text while the text itself never appeared in the stream.
func TestStopStderrIsContinuationPrompt(t *testing.T) {
	t.Parallel()

	rule, ok := ClassifyStderr(hook.EventStop)
	if !ok {
		t.Fatal("Stop has no stderr classification")
	}
	if rule.Class != StderrContinuationPrompt {
		t.Errorf("class = %s, want %s", rule.Class, StderrContinuationPrompt)
	}
	if rule.Evidence != EvidenceMeasured {
		t.Errorf("evidence = %s, want %s", rule.Evidence, EvidenceMeasured)
	}
}

// TestTwoClassesAreDistinct guards the pairing itself: were both events to
// collapse onto one class, AC-REQ-4a/4b would pass individually while the
// distinction they exist to protect was gone.
func TestTwoClassesAreDistinct(t *testing.T) {
	t.Parallel()

	pre, _ := ClassifyStderr(hook.EventPreToolUse)
	stop, _ := ClassifyStderr(hook.EventStop)
	if pre.Class == stop.Class {
		t.Fatalf("PreToolUse and Stop share class %s; the measured behaviors differ", pre.Class)
	}
}

// TestUnmeasuredClassIsAnnotated — AC-REQ-4c.
func TestUnmeasuredClassIsAnnotated(t *testing.T) {
	t.Parallel()

	rule, ok := ClassifyStderr(hook.EventUserPromptSubmit)
	if !ok {
		t.Fatal("UserPromptSubmit has no stderr classification")
	}
	if rule.Evidence != EvidenceDeclared {
		t.Errorf("evidence = %s, want %s — exit 2 was never exercised on this event",
			rule.Evidence, EvidenceDeclared)
	}
}

// TestExcludedEventsHaveNoClass — AC-REQ-4c.
//
// Events SPEC §B excludes from adaptation must not appear in the table; a
// classification there would assert a contract nothing measured.
func TestExcludedEventsHaveNoClass(t *testing.T) {
	t.Parallel()

	for _, e := range []hook.EventType{
		hook.EventPreCompact,
		hook.EventPostCompact,
		hook.EventPermissionRequest,
		hook.EventSubagentStart,
		hook.EventSubagentStop,
	} {
		if _, ok := ClassifyStderr(e); ok {
			t.Errorf("%s carries a stderr classification but is excluded from adaptation", e)
		}
	}
}

// TestClassifyStderrDoesNotDefault asserts an unknown event yields no rule
// rather than a silent fallback.
func TestClassifyStderrDoesNotDefault(t *testing.T) {
	t.Parallel()

	if _, ok := ClassifyStderr(hook.EventType("NotAnEvent")); ok {
		t.Fatal("unknown event received a stderr classification; want none")
	}
}
