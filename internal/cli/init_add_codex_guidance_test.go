package cli

// SPEC-UPDATE-ADD-CODEX-001 M3 — init --force redirect guidance (REQ-UAC-013,
// decision D4). When init runs --agent codex|both against an
// already-initialized project (the --force reinit path), init prints guidance
// naming `moai update --add-codex` as the sanctioned additive path BEFORE
// proceeding. The reinit itself is not blocked — the codex-add *purpose* is
// redirected, the reinit *capability* remains (redirect-not-block).

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestInitAddCodexGuidanceNamesAdditiveVerb pins the REQ-UAC-013 wording: the
// guidance must name `moai update --add-codex` (the AC grep anchor) and
// describe it as the sanctioned additive path.
func TestInitAddCodexGuidanceNamesAdditiveVerb(t *testing.T) {
	if !strings.Contains(addCodexReinitGuidance, "update --add-codex") {
		t.Errorf("guidance does not name the additive verb `update --add-codex`:\n%s", addCodexReinitGuidance)
	}
	if !strings.Contains(addCodexReinitGuidance, "additive") {
		t.Errorf("guidance does not describe the additive path:\n%s", addCodexReinitGuidance)
	}
}

// TestEmitAddCodexReinitGuidanceConditions covers the redirect-not-block
// matrix: the guidance prints ONLY for a codex|both selection on an
// already-initialized project; a claude selection and a fresh project stay
// silent, and a nil writer is safe.
func TestEmitAddCodexReinitGuidanceConditions(t *testing.T) {
	cases := []struct {
		name               string
		wiring             agentWiring
		alreadyInitialized bool
		wantGuidance       bool
	}{
		{"codex+initialized", agentWiringCodex, true, true},
		{"both+initialized", agentWiringBoth, true, true},
		{"codex+fresh", agentWiringCodex, false, false},
		{"claude+initialized", agentWiringClaude, true, false},
		{"claude+fresh", agentWiringClaude, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			emitAddCodexReinitGuidance(&out, c.wiring, c.alreadyInitialized)
			got := out.Len() > 0
			if got != c.wantGuidance {
				t.Errorf("guidance printed = %v, want %v (output: %q)", got, c.wantGuidance, out.String())
			}
		})
	}
	// Nil writer must be safe (defensive — callers pass cmd.ErrOrStderr).
	emitAddCodexReinitGuidance(nil, agentWiringBoth, true)
}

// TestInitAddCodexGuidanceRunsBeforeExecute is the placement guard: runInit
// must emit the guidance BEFORE executor.Execute, under the --force +
// codex|both + already-initialized conditions, so the reinit proceeds after
// the operator has seen the additive-path note (redirect-not-block).
func TestInitAddCodexGuidanceRunsBeforeExecute(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)
	// The call-site form is unique: the runInit call passes cmd.ErrOrStderr(),
	// while the definition and the emit function's own body use other forms.
	emitIdx := strings.Index(body, "emitAddCodexReinitGuidance(cmd.ErrOrStderr()")
	if emitIdx < 0 {
		t.Fatal("runInit does not call emitAddCodexReinitGuidance — the --force codex-add redirect guidance is not wired (REQ-UAC-013)")
	}
	execIdx := strings.Index(body, "executor.Execute(ctx, opts)")
	if execIdx < 0 {
		t.Fatal("executor.Execute call not found in init.go — test premise stale")
	}
	if emitIdx > execIdx {
		t.Error("guidance emitted AFTER executor.Execute — it must print BEFORE the reinit proceeds (REQ-UAC-013)")
	}
	before := body[max(0, emitIdx-450):emitIdx]
	after := body[emitIdx:min(len(body), emitIdx+250)]
	if !strings.Contains(before, `getBoolFlag(cmd, "force")`) {
		t.Errorf("emission is not guarded by --force:\n%s", before)
	}
	if !strings.Contains(before, "agentWiringSelection != agentWiringClaude") {
		t.Errorf("emission is not guarded by a codex|both selection:\n%s", before)
	}
	if !strings.Contains(after, "!probe.Valid") {
		t.Errorf("emission is not guarded by the already-initialized probe:\n%s", after)
	}
}
