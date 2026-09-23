package codexadapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// Decision is a decision-bearing handler result normalized before any
// harness-specific translation (SPEC-DUAL-HARNESS-HOOK-PARITY-001 REQ-HPR-006,
// design.md §D1).
type Decision string

const (
	// DecisionAllow: proceed.
	DecisionAllow Decision = "allow"
	// DecisionDeny: refuse, with a reason.
	DecisionDeny Decision = "deny"
	// DecisionNeedsInput: a human must decide.
	DecisionNeedsInput Decision = "needs_input"
	// DecisionRetryableError: the handler failed and a retry is safe.
	DecisionRetryableError Decision = "retryable_error"
	// DecisionFatalError: the handler failed on a decision-bearing event.
	DecisionFatalError Decision = "fatal_error"
)

// Decisions lists the normalized decision vocabulary in declaration order.
func Decisions() []Decision {
	return []Decision{DecisionAllow, DecisionDeny, DecisionNeedsInput, DecisionRetryableError, DecisionFatalError}
}

// Harness names the host a translation renders for.
type Harness string

const (
	// HarnessClaude is Claude Code.
	HarnessClaude Harness = "claude"
	// HarnessCodex is the Codex CLI.
	HarnessCodex Harness = "codex"
)

// Outcome is the host-facing shape a normalized decision is rendered as.
type Outcome string

const (
	// OutcomeNoOpinion renders `{}`: the hook expresses no decision.
	OutcomeNoOpinion Outcome = "no-opinion"
	// OutcomeAllow renders an explicit allow decision.
	OutcomeAllow Outcome = "allow"
	// OutcomeDeny renders an explicit deny (PreToolUse, PermissionRequest) or
	// block (Stop, UserPromptSubmit), always with a non-empty reason.
	OutcomeDeny Outcome = "deny"
	// OutcomeAsk renders the host-native approval prompt (Claude PreToolUse).
	OutcomeAsk Outcome = "ask"
	// OutcomeAdvisory renders no decision; the message goes to the advisory
	// record instead of the decision channel.
	OutcomeAdvisory Outcome = "advisory"
)

// Translation is one row of the translation table: how one normalized
// decision on one event is rendered for one harness.
type Translation struct {
	Event    hook.EventType
	Harness  Harness
	Decision Decision
	Outcome  Outcome
	// DiscardRecorded marks a conversion that must also be written to the
	// adapter's discard sink (RecordDiscards), so it is visible rather than
	// silent (design.md §D5, operator decision Q2).
	DiscardRecorded bool
}

// DecisionBearingEvents are the events whose hook output can change what the
// host does next.
func DecisionBearingEvents() []hook.EventType {
	return []hook.EventType{hook.EventPreToolUse, hook.EventPermissionRequest, hook.EventStop, hook.EventUserPromptSubmit}
}

// @MX:NOTE: [AUTO] decision translation table — a Codex deny or needs_input row must never render as allow or the empty object (REQ-HPR-007); AC-HPR-006 checks the rows
// @MX:SPEC: SPEC-DUAL-HARNESS-HOOK-PARITY-001
// translationTable is the data behind REQ-HPR-006: one row per event x
// harness x normalized decision. The never-loosens property of REQ-HPR-007 is
// checked against these rows (AC-HPR-006), not against prose.
//
// The live Codex output path (MapOutput) is switched onto these rows in M2c;
// until then the table is the declared contract and MapOutput keeps its
// card-t590 behavior.
var translationTable = buildTranslationTable()

func buildTranslationTable() []Translation {
	type cell struct {
		outcome Outcome
		discard bool
	}
	row := func(o Outcome) cell { return cell{outcome: o} }
	recorded := func(o Outcome) cell { return cell{outcome: o, discard: true} }

	// Per harness, per event: allow, deny, needs_input, retryable_error, fatal_error.
	spec := map[Harness]map[hook.EventType][5]cell{
		HarnessClaude: {
			hook.EventPreToolUse: {row(OutcomeAllow), row(OutcomeDeny), row(OutcomeAsk), row(OutcomeAdvisory), row(OutcomeDeny)},
			// No decision on PermissionRequest leaves Claude's own permission
			// dialog in front of the user, which is the native needs_input.
			hook.EventPermissionRequest: {row(OutcomeAllow), row(OutcomeDeny), row(OutcomeNoOpinion), row(OutcomeAdvisory), row(OutcomeDeny)},
			hook.EventStop:              {row(OutcomeNoOpinion), row(OutcomeDeny), row(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
			hook.EventUserPromptSubmit:  {row(OutcomeNoOpinion), row(OutcomeDeny), row(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
		},
		HarnessCodex: {
			// Codex refuses permissionDecision:allow without updatedInput, so an
			// allow is the no-opinion object (card t590). needs_input is a
			// fail-closed deny, recorded (Q2).
			hook.EventPreToolUse:        {row(OutcomeNoOpinion), row(OutcomeDeny), recorded(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
			hook.EventPermissionRequest: {row(OutcomeAllow), row(OutcomeDeny), recorded(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
			hook.EventStop:              {row(OutcomeNoOpinion), row(OutcomeDeny), recorded(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
			hook.EventUserPromptSubmit:  {row(OutcomeNoOpinion), row(OutcomeDeny), recorded(OutcomeDeny), row(OutcomeAdvisory), row(OutcomeDeny)},
		},
	}

	var out []Translation
	for _, h := range []Harness{HarnessClaude, HarnessCodex} {
		for _, ev := range DecisionBearingEvents() {
			cells := spec[h][ev]
			for i, d := range Decisions() {
				out = append(out, Translation{Event: ev, Harness: h, Decision: d, Outcome: cells[i].outcome, DiscardRecorded: cells[i].discard})
			}
		}
	}
	return out
}

// TranslationTable returns a copy of the translation table.
func TranslationTable() []Translation {
	return append([]Translation(nil), translationTable...)
}

// Lookup returns the translation for one harness, event, and decision.
func Lookup(h Harness, ev hook.EventType, d Decision) (Translation, bool) {
	for _, row := range translationTable {
		if row.Harness == h && row.Event == ev && row.Decision == d {
			return row, true
		}
	}
	return Translation{}, false
}

// HostResolvesAsAllow reports whether the host proceeds as if allowed when it
// receives outcome o on event ev. It is deliberately conservative: an outcome
// whose host resolution is unmeasured is counted as allow (research.md H1,
// H2), so the never-loosens check cannot pass on an unproven assumption.
func HostResolvesAsAllow(h Harness, ev hook.EventType, o Outcome) bool {
	switch o {
	case OutcomeAllow:
		return true
	case OutcomeDeny, OutcomeAsk:
		return false
	}
	// OutcomeNoOpinion and OutcomeAdvisory carry no decision.
	if h == HarnessClaude && ev == hook.EventPermissionRequest {
		// Claude shows its own permission dialog when a PermissionRequest
		// hook expresses no decision.
		return false
	}
	return true
}

// Render produces the hook stdout bytes for outcome o on event ev. A deny or
// ask without a usable reason is refused: Codex rejects a blank-reason deny,
// which would turn it into a no-op.
func Render(ev hook.EventType, o Outcome, reason string) ([]byte, error) {
	needsReason := o == OutcomeDeny || o == OutcomeAsk
	if needsReason && strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("render %s on %s: a reason is required", o, ev)
	}
	var v any
	switch o {
	case OutcomeNoOpinion, OutcomeAdvisory:
		v = map[string]any{}
	case OutcomeAllow, OutcomeDeny, OutcomeAsk:
		switch ev {
		case hook.EventPreToolUse:
			hso := map[string]any{"hookEventName": string(ev), "permissionDecision": string(o)}
			if needsReason {
				hso["permissionDecisionReason"] = reason
			}
			v = map[string]any{"hookSpecificOutput": hso}
		case hook.EventPermissionRequest:
			if o == OutcomeAsk {
				return nil, fmt.Errorf("render ask on %s: no ask shape exists", ev)
			}
			dec := map[string]any{"behavior": string(o)}
			if o == OutcomeDeny {
				dec["message"] = reason
			}
			v = map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": string(ev), "decision": dec}}
		case hook.EventStop, hook.EventUserPromptSubmit:
			switch o {
			case OutcomeDeny:
				v = map[string]any{"decision": "block", "reason": reason}
			default:
				return nil, fmt.Errorf("render %s on %s: no such shape", o, ev)
			}
		default:
			return nil, fmt.Errorf("render %s: %s is not a decision-bearing event", o, ev)
		}
	default:
		return nil, fmt.Errorf("render: unknown outcome %q", o)
	}
	return json.Marshal(v)
}
