package hook

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/binlag"
)

// stubComparer installs a verdict in the ONE comparison seam and restores the
// real one afterwards. Both surfaces read this seam, which is what makes the
// single-implementation property observable rather than merely asserted.
func stubComparer(t *testing.T, v binlag.Verdict) {
	t.Helper()
	orig := binlag.Comparer
	t.Cleanup(func() { binlag.Comparer = orig })
	binlag.Comparer = func(context.Context, binlag.Request) binlag.Verdict { return v }
}

const (
	stubBinaryCommit = "aaaaaaaaabbbbbbbbbccccccccc1111111112222"
	stubSourceHead   = "ffffffffff99999999988888888877777777666"
)

func behindVerdict() binlag.Verdict {
	return binlag.Verdict{
		Status:       binlag.StatusBehind,
		BinaryCommit: stubBinaryCommit,
		SourceHead:   stubSourceHead,
	}
}

// handleForLagTest runs the session-start handler against a throwaway project
// directory and returns its output.
func handleForLagTest(t *testing.T, sessionID string) *HookOutput {
	t.Helper()
	projectDir := t.TempDir()
	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     sessionID,
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	return out
}

// AC-BLV-001 — the verdict reaches an observer with no diagnostic command
// invoked, and names both commits.
func TestSessionStart_LagAdvisoryReachesAdditionalContext(t *testing.T) {
	stubComparer(t, behindVerdict())

	out := handleForLagTest(t, "sess-lag-behind")
	if out.HookSpecificOutput == nil {
		t.Fatal("no hookSpecificOutput; the advisory had no surface to land on")
	}
	got := out.HookSpecificOutput.AdditionalContext
	for _, want := range []string{binlag.Short(stubBinaryCommit), binlag.Short(stubSourceHead)} {
		if !strings.Contains(got, want) {
			t.Errorf("additionalContext does not name %q:\n%s", want, got)
		}
	}
}

// AC-BLV-002 — the control group. A binary at HEAD produces silence, so the
// advisory cannot be satisfied by a build that always speaks.
func TestSessionStart_NoLagAdvisoryWhenBinaryMatchesHead(t *testing.T) {
	stubComparer(t, binlag.Verdict{
		Status:       binlag.StatusFresh,
		BinaryCommit: stubSourceHead,
		SourceHead:   stubSourceHead,
	})

	out := handleForLagTest(t, "sess-lag-fresh")
	got := ""
	if out.HookSpecificOutput != nil {
		got = out.HookSpecificOutput.AdditionalContext
	}
	if strings.Contains(got, "moai binary lag") {
		t.Errorf("lag advisory emitted for a binary that matches HEAD:\n%s", got)
	}
	// The pre-existing attribution string must survive untouched.
	if !strings.Contains(got, "moai session attribution") {
		t.Errorf("existing additionalContext content was lost:\n%s", got)
	}
}

// AC-BLV-008 — the advisory is judged on the SERIALIZED bytes, under the
// hookSpecificOutput.additionalContext key specifically. Reading the struct
// field would pass a `json:"-"` field, and searching the whole document would
// pass an advisory written to systemMessage instead.
func TestSessionStart_LagAdvisorySerializesUnderAdditionalContext(t *testing.T) {
	stubComparer(t, behindVerdict())

	out := handleForLagTest(t, "sess-lag-serialized")
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal handler output: %v", err)
	}

	var doc struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal handler output: %v", err)
	}
	if !strings.Contains(doc.HookSpecificOutput.AdditionalContext, "moai binary lag") {
		t.Fatalf("lag advisory absent from the serialized hookSpecificOutput.additionalContext:\n%s", raw)
	}
}

// AC-BLV-006 — a seam that never returns must not hold the session start.
//
// The stub deliberately IGNORES context cancellation. A stub that honoured it
// would let a `context.WithTimeout` wrapper release the handler too, so the
// criterion would stop distinguishing the two shapes. Under a stub that
// ignores cancellation, only a genuine caller-side timer+select join returns
// in time — which is exactly what REQ-BLV-006 asks for: the handler must not
// block whatever the seam does.
//
// The async branch must be switched on for this: TestMain sets
// deferredScansAsync=false for the whole test binary, and the inline branch
// has no bounded join at all.
func TestSessionStart_BlockingComparerDoesNotStallSessionStart(t *testing.T) {
	release := make(chan struct{})
	stubExited := make(chan struct{})

	orig := binlag.Comparer
	origAsync := deferredScansAsync
	deferredScanSeamMu.Lock()
	deferredScansAsync = true
	deferredScanSeamMu.Unlock()
	binlag.Comparer = func(context.Context, binlag.Request) binlag.Verdict {
		defer close(stubExited)
		<-release // ignores ctx on purpose
		return behindVerdict()
	}
	// Registered FIRST so it runs LAST: t.Cleanup is LIFO, and the seam may
	// only be restored once the abandoned goroutine has stopped reading it.
	// Swapping these two registrations reinstates the data race.
	t.Cleanup(func() {
		binlag.Comparer = orig
		deferredScanSeamMu.Lock()
		deferredScansAsync = origAsync
		deferredScanSeamMu.Unlock()
	})
	// Registered SECOND so it runs FIRST: release the deliberately-abandoned
	// goroutine and JOIN it before the restore above. The join belongs in
	// cleanup and nowhere earlier — waiting for the stub before the assertion
	// would delete the very property AC-BLV-006 measures.
	t.Cleanup(func() {
		close(release)
		select {
		case <-stubExited:
		case <-time.After(10 * time.Second):
			t.Errorf("stub comparer did not exit after release; restoring the seam would race with it")
		}
	})

	start := time.Now()
	out := handleForLagTest(t, "sess-lag-blocked")
	elapsed := time.Since(start)

	if elapsed > binaryLagJoinBound+2*time.Second {
		t.Errorf("Handle blocked %v on a non-returning comparer; expected a return near the %v bound",
			elapsed, binaryLagJoinBound)
	}
	got := ""
	if out.HookSpecificOutput != nil {
		got = out.HookSpecificOutput.AdditionalContext
	}
	if strings.Contains(got, "moai binary lag") {
		t.Errorf("an advisory was emitted although the comparison never finished:\n%s", got)
	}
}
