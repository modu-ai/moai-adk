package hook

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// userPromptRow renders a user transcript row whose message content is the
// spawn prompt text — the row an auditor transcript carries as its first
// message, naming the SPEC the audit was spawned for.
func userPromptRow(prompt string) string {
	b, _ := json.Marshal(map[string]any{
		"type":    "user",
		"message": map[string]any{"role": "user", "content": prompt},
	})
	return string(b)
}

// driftTranscriptWithPrompt writes an auditor transcript whose first user row
// is the spawn prompt and whose assistant rows are served by the wrong model —
// the t1240 re-audit shape: the model drifted AND the final message carries no
// parseable verdict line, so the verdict-line SPEC attribution fails.
func driftTranscriptWithPrompt(t *testing.T, agentType, agentID, spawnPrompt string) string {
	t.Helper()
	return writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), agentID,
		[]string{
			userPromptRow(spawnPrompt),
			assistantRow("glm-5.3-flash"),
			assistantRow("glm-5.3-flash"),
		}, map[string]string{"agentType": agentType, "model": "opus"})
}

// spawnWithPrompt builds a PreToolUse spawn whose prompt names a SPEC — the
// field the spawn-side attribution matches against.
func spawnWithPrompt(root, subagentType, prompt string) *HookInput {
	raw, _ := json.Marshal(map[string]string{"subagent_type": subagentType, "prompt": prompt})
	return &HookInput{CWD: root, ToolName: "Agent", ToolInput: raw, HookEventName: "PreToolUse"}
}

// jsonMarshal is a tiny indirection so the test file does not import
// encoding/json twice under two names.

// AC (card t1323) — the refusal a re-audit leaves attributes to the AUDITED
// SPEC (extracted from the auditor transcript's spawn prompt), and the
// spawn-side deny matches that attribution: an unrelated milestone's spawn is
// not blocked, while a spawn for the refused SPEC still is. The observed
// defect (worker-70, card t1240): two plan-auditor re-audit spawns left
// plan-auditor--unknown-spec--served.json in the spawning session's tree, and
// an unrelated M6 manager-develop spawn was denied by it.
func TestServedModel_RefusalAttributesToAuditedSpec(t *testing.T) {
	root := newServedGateTree(t, "true", "")
	const audited = "SPEC-T1240-001"
	const unrelated = "SPEC-M6-006"

	// (1) The audit: a plan-auditor re-audit whose spawn prompt names the
	// audited SPEC, served by the wrong model, final message without a
	// parseable verdict line.
	transcript := driftTranscriptWithPrompt(t, "plan-auditor", "g-1240", "plan-audit re-run for "+audited+": audit the plan artifacts")
	out := runServedStop(t, nil, servedStopInput(root, "plan-auditor", "g-1240", transcript, "audit complete; report attached"))
	if out == nil || out.Decision != "" {
		t.Fatalf("stop output = %+v, want no decision", out)
	}

	// (2) The refusal attributes to the audited SPEC — not unknown-spec.
	refusals := rejectionsOfKind(t, root, auditreceipt.KindServed)
	if len(refusals) != 1 {
		t.Fatalf("served refusals = %+v, want exactly 1", refusals)
	}
	if refusals[0].SpecID != audited {
		t.Fatalf("refusal SpecID = %q, want the audited %q (attribution defect)", refusals[0].SpecID, audited)
	}
	if !strings.Contains(refusals[0].Cause, "glm-5.3-flash") {
		t.Fatalf("refusal cause %q does not name the served model", refusals[0].Cause)
	}

	// (3) Negative — an unrelated milestone's phase-entry spawn is not blocked.
	if ok, reason := denied(runSpawn(t, spawnWithPrompt(root, "manager-develop",
		"M6 work for "+unrelated+": implement the milestone"))); ok {
		t.Fatalf("unrelated-spec spawn denied: %q", reason)
	}

	// (4) The gate keeps its teeth — a spawn for the refused SPEC is denied.
	if ok, reason := denied(runSpawn(t, spawnWithPrompt(root, "manager-develop",
		"implement "+audited+" phase entry"))); !ok || !strings.HasPrefix(reason, servedModelViolation) {
		t.Fatalf("refused-SPEC spawn: denied=%v reason=%q, want a %s deny", ok, reason, servedModelViolation)
	}
}

// AC (card t1323, negative) — refusals recorded in one tree do not block
// spawns attributed to another tree, even in the same shared store.
func TestServedModel_RefusalIsTreeScoped(t *testing.T) {
	other := newServedGateTree(t, "true", "")
	r := auditreceipt.Rejection{AgentType: "plan-auditor", SpecID: "SPEC-T1240-001", Kind: auditreceipt.KindServed,
		Cause: auditreceipt.CauseServedModelDrift + ": expected opus, served [glm-5.3-flash]", TreeRoot: other}
	if err := auditreceipt.WriteRejectionIn(other, &r); err != nil {
		t.Fatalf("WriteRejectionIn: %v", err)
	}

	// A different gate-on tree sharing the same primary store: a spawn here
	// must not see the other tree's refusal.
	mine := newServedGateTree(t, "true", "")
	if ok, reason := denied(runSpawn(t, spawnWithPrompt(mine, "manager-develop", "work for SPEC-1323-001"))); ok {
		t.Fatalf("spawn in an unrelated tree denied: %q", reason)
	}
	_ = filepath.Join // keep filepath linked if assertions above change
}
