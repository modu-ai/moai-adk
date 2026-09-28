package template

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// go test -json actions the aggregate reads (acceptance.md §B rule P).
const (
	ActionPass = "pass"
	ActionFail = "fail"
	ActionSkip = "skip"
)

// Per-obligation and overall verdicts (SPEC-DUAL-HARNESS-HOOK-PARITY-001
// AC-HPR-019). Only VerdictPass is a pass; every other value keeps the
// aggregate below PASS.
const (
	VerdictPass         = "PASS"
	VerdictFail         = "FAIL"
	VerdictNotRun       = "NOT_RUN"
	VerdictUnsupported  = "UNSUPPORTED"
	VerdictBlocked      = "BLOCKED"
	VerdictUnverified   = "UNVERIFIED"
	VerdictUnattributed = "UNATTRIBUTED"
	VerdictInsufficient = "INSUFFICIENT_EVIDENCE"
)

// Evidence levels (acceptance.md §B rule 3): registration proves only that a
// handler is configured; only effect-verified evidence satisfies an
// obligation.
const (
	EvidenceRegistered     = "registered"
	EvidenceFired          = "fired"
	EvidenceEffectVerified = "effect-verified"
)

// ParityRecord is one verdict record with its attribution (acceptance.md §B
// rule 7, REQ-HPR-024). A Stop gate the §D3.8 cap recorded carries Verdict
// "UNVERIFIED".
type ParityRecord struct {
	Verdict       string `json:"verdict"`
	EvidenceLevel string `json:"evidence_level"`
	Commit        string `json:"commit"`
	TreeDigest    string `json:"tree_digest"`
	ClaudeVersion string `json:"claude_version"`
	CodexVersion  string `json:"codex_version"`
	Uname         string `json:"uname"`
}

// ObligationEvidence is what the aggregate reads for one obligation: the
// go-test action of its check and its verdict record.
type ObligationEvidence struct {
	Action string
	Record *ParityRecord
}

// ObligationVerdict is one obligation's aggregated verdict.
type ObligationVerdict struct {
	ID      string
	Verdict string
	Reason  string
}

// ParityAggregate is the aggregate over the whole registry.
type ParityAggregate struct {
	Overall       string
	PerObligation []ObligationVerdict
}

// @MX:NOTE: [AUTO] parity aggregate (AC-HPR-019) — only an effect-verified, fully attributed, passing required obligation is PASS; a skip, empty run, NOT_RUN, UNSUPPORTED, blocked, unverified, or unattributed row keeps the whole verdict below PASS
// @MX:SPEC: SPEC-DUAL-HARNESS-HOOK-PARITY-001
// AggregateParityVerdict computes each required obligation's verdict from its
// registry paths, its check's go-test action, and its verdict record, taking
// the weakest, and reports PASS overall only when every required obligation is
// PASS. An empty registry is not PASS.
func AggregateParityVerdict(reg *ObligationRegistry, evidence map[string]ObligationEvidence) ParityAggregate {
	agg := ParityAggregate{Overall: VerdictPass}
	if reg == nil || len(reg.Obligations) == 0 {
		agg.Overall = VerdictNotRun
		return agg
	}
	for _, o := range reg.Obligations {
		if !o.Required {
			continue
		}
		v := obligationVerdict(o, evidence[o.ID])
		agg.PerObligation = append(agg.PerObligation, v)
		if v.Verdict != VerdictPass {
			agg.Overall = VerdictFail
		}
	}
	return agg
}

func obligationVerdict(o Obligation, ev ObligationEvidence) ObligationVerdict {
	out := ObligationVerdict{ID: o.ID}
	for _, p := range []string{o.ClaudePath, o.CodexPath} {
		marker, ref, _ := ParseApplicationPath(p)
		switch marker {
		case MarkerUnsupported:
			out.Verdict, out.Reason = VerdictUnsupported, ref
			return out
		case MarkerBlocked:
			out.Verdict, out.Reason = VerdictBlocked, ref
			return out
		case MarkerUnverified:
			out.Verdict, out.Reason = VerdictUnverified, ref
			return out
		}
	}
	switch ev.Action {
	case ActionPass:
	case ActionFail:
		out.Verdict, out.Reason = VerdictFail, o.Check+" failed"
		return out
	case ActionSkip:
		out.Verdict, out.Reason = VerdictNotRun, o.Check+" skipped"
		return out
	default:
		out.Verdict, out.Reason = VerdictNotRun, o.Check+" has no go-test record (empty run)"
		return out
	}
	r := ev.Record
	if r == nil {
		out.Verdict, out.Reason = VerdictUnattributed, "no verdict record"
		return out
	}
	switch strings.ToUpper(r.Verdict) {
	case VerdictPass:
	case VerdictNotRun:
		out.Verdict, out.Reason = VerdictNotRun, "the verdict record reads NOT_RUN"
		return out
	case VerdictUnsupported:
		out.Verdict, out.Reason = VerdictUnsupported, "the verdict record reads UNSUPPORTED"
		return out
	case VerdictUnverified:
		out.Verdict, out.Reason = VerdictUnverified, "the verdict record reads unverified"
		return out
	default:
		out.Verdict, out.Reason = VerdictFail, "the verdict record reads "+r.Verdict
		return out
	}
	for field, val := range map[string]string{"commit": r.Commit, "tree_digest": r.TreeDigest, "claude_version": r.ClaudeVersion, "codex_version": r.CodexVersion, "uname": r.Uname} {
		if strings.TrimSpace(val) == "" {
			out.Verdict, out.Reason = VerdictUnattributed, "attribution field "+field+" is empty"
			return out
		}
	}
	if r.EvidenceLevel != EvidenceEffectVerified {
		out.Verdict, out.Reason = VerdictInsufficient, fmt.Sprintf("evidence level %q is not %s", r.EvidenceLevel, EvidenceEffectVerified)
		return out
	}
	out.Verdict = VerdictPass
	return out
}

// ReadGoTestActions reads a `go test -json` stream and returns one action per
// top-level test (rule P): a skip or fail in the test or any of its subtests
// wins over a pass. A test with no record is absent, which the aggregate reads
// as an empty run.
func ReadGoTestActions(r io.Reader) (map[string]string, error) {
	rank := map[string]int{ActionPass: 1, ActionSkip: 2, ActionFail: 3}
	out := map[string]string{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var ev struct {
			Action string `json:"Action"`
			Test   string `json:"Test"`
		}
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil || ev.Test == "" {
			continue
		}
		if rank[ev.Action] == 0 {
			continue
		}
		top, _, _ := strings.Cut(ev.Test, "/")
		if rank[ev.Action] > rank[out[top]] {
			out[top] = ev.Action
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read go test -json: %w", err)
	}
	return out, nil
}
