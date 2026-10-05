package escalation_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

var fixedNow = time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

// lineAt returns the 1-based line n of data.
func lineAt(t *testing.T, data []byte, n int) string {
	t.Helper()
	lines := strings.Split(string(data), "\n")
	if n < 1 || n > len(lines) {
		t.Fatalf("line %d out of range (%d lines)", n, len(lines))
	}
	return lines[n-1]
}

// opRecord returns an operational record of class with contract_ref ref.
func opRecord(card, class, ref string, parts ...string) escalation.Record {
	return escalation.Record{
		Card: card, Spec: "SPEC-A-001", Kind: escalation.KindOperational, Class: class,
		Fingerprint: escalation.Fingerprint(class, parts...), ContractRef: ref,
		HeadSHA: escalationtest.HeadSHA, Observation: class + " observed",
		Options: []string{"Raise the limit", "Stop the run"},
	}
}

// AC-AE-022 (REQ-AE-018, REQ-AE-019): every written record and a
// hand-authored revoke record parse as YAML frontmatter with every §I.1 field
// and type, a fingerprint equal to the file name's, a status in the closed
// set, the contract_ref form of its class, and the three body sections in
// order; the needs-decision reading counts the revoke record as not open.
func TestRecordFrontmatterParses(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	cdata := w.Read(".moai/specs/SPEC-A-001/contract.yaml")

	neverLine := escalation.ContractItemLine(cdata, "internal/x/**", "ownership", "never")
	retriesLine := escalation.ContractLine(cdata, "budget", "audit_retries")
	if neverLine == 0 || retriesLine == 0 {
		t.Fatalf("line mapping failed: never=%d audit_retries=%d", neverLine, retriesLine)
	}

	contractRec := escalation.Record{
		Card: "t9001", Spec: "SPEC-A-001", Kind: escalation.KindContract, Class: escalation.ClassOwnershipMove,
		Fingerprint: escalation.Fingerprint(escalation.ClassOwnershipMove, "internal/x/y.go", "internal/x/**"),
		ContractRef: "contract.yaml:" + strconv.Itoa(neverLine), EscalateOn: "ownership-move",
		HeadSHA: escalationtest.HeadSHA, NotObserved: []string{"session scratchpad root"},
		Observation: "Write internal/x/y.go", Options: []string{"Revert", "Amend the contract"},
	}
	written := []escalation.Record{
		contractRec,
		opRecord("t9001", escalation.ClassBudgetExceeded, "config:workflow.autonomy.escalation.budget_default.operations", "operations"),
		opRecord("t9001", escalation.ClassSameDiagnosticRepeat, "rule:same-diagnostic-3", "go test ./x", "FAIL x"),
		opRecord("t9001", escalation.ClassAuditFailAtRetryCap, "contract.yaml:"+strconv.Itoa(retriesLine), "plan-audit"),
		opRecord("t9001", escalation.ClassDetectionDisarmed, "disarm:signature-invalid", strings.Repeat("a", 64)),
	}
	var paths []string
	for _, r := range written {
		p, err := escalation.WriteRecord(w.Root, r, fixedNow)
		if err != nil {
			t.Fatalf("WriteRecord %s: %v", r.Class, err)
		}
		paths = append(paths, p)
	}
	revokeFP := escalation.Fingerprint(escalation.ClassRevokeOperator, "operator")
	revokePath := escalation.RecordPath(w.Root, "t9001", escalation.ClassRevokeOperator, revokeFP, 0)
	writeFile(t, revokePath, "---\n"+
		"schema_version: 1\ncard: t9001\nspec: SPEC-A-001\nkind: revoke\nclass: revoke-operator\n"+
		"fingerprint: "+revokeFP+"\ncontract_ref: \"\"\nescalate_on: \"\"\nstatus: open\ndecider: human\n"+
		"occurrences: 1\nhead_sha: "+escalationtest.HeadSHA+"\ndetected_at: \"2026-03-04T05:06:07Z\"\n"+
		"updated_at: \"2026-03-04T05:06:07Z\"\nnot_observed: []\n---\n\n## Observation\n\nrevoked\n\n"+
		"## Options\n\n1. Re-sign\n2. Abandon\n\n## Not observed\n\n- none\n")
	paths = append(paths, revokePath)

	refForm := map[string]*regexp.Regexp{
		escalation.ClassOwnershipMove:        regexp.MustCompile(`^contract\.yaml:\d+$`),
		escalation.ClassBudgetExceeded:       regexp.MustCompile(`^(config:workflow\.autonomy\.escalation\.budget_default\.\w+|contract\.yaml:\d+)$`),
		escalation.ClassSameDiagnosticRepeat: regexp.MustCompile(`^rule:same-diagnostic-3$`),
		escalation.ClassAuditFailAtRetryCap:  regexp.MustCompile(`^contract\.yaml:\d+$`),
		escalation.ClassDetectionDisarmed:    regexp.MustCompile(`^disarm:(contract-absent|signature-invalid|terminal-status|card-mismatch|state-tamper)$`),
		escalation.ClassRevokeOperator:       regexp.MustCompile(`^$`),
	}
	nameRe := regexp.MustCompile(`^(.+)-([0-9a-f]{16})(-\d+)?\.md$`)
	fields := map[string]string{
		"schema_version": "int", "card": "string", "spec": "string", "kind": "string", "class": "string",
		"fingerprint": "string", "contract_ref": "string", "escalate_on": "string", "status": "string",
		"decider": "string", "occurrences": "int", "head_sha": "string", "detected_at": "string",
		"updated_at": "string", "not_observed": "list",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		end := strings.Index(text[4:], "\n---\n")
		if !strings.HasPrefix(text, "---\n") || end < 0 {
			t.Fatalf("%s: no frontmatter", filepath.Base(p))
		}
		var fm map[string]any
		if err := yaml.Unmarshal([]byte(text[4:4+end+1]), &fm); err != nil {
			t.Fatalf("%s: YAML: %v", filepath.Base(p), err)
		}
		for k, typ := range fields {
			v, ok := fm[k]
			if !ok {
				t.Errorf("%s: field %s missing", filepath.Base(p), k)
				continue
			}
			switch typ {
			case "int":
				if _, ok := v.(int); !ok {
					t.Errorf("%s: %s = %#v, want int", filepath.Base(p), k, v)
				}
			case "string":
				if _, ok := v.(string); !ok {
					t.Errorf("%s: %s = %#v, want string", filepath.Base(p), k, v)
				}
			case "list":
				if _, ok := v.([]any); !ok {
					t.Errorf("%s: %s = %#v, want list", filepath.Base(p), k, v)
				}
			}
		}
		m := nameRe.FindStringSubmatch(filepath.Base(p))
		if m == nil || m[1] != fm["class"] || m[2] != fm["fingerprint"] {
			t.Errorf("%s: file name does not carry class %v and fingerprint %v", filepath.Base(p), fm["class"], fm["fingerprint"])
		}
		if st := fm["status"]; st != "open" && st != "resolved" {
			t.Errorf("%s: status %v", filepath.Base(p), st)
		}
		ref, _ := fm["contract_ref"].(string)
		if re := refForm[fm["class"].(string)]; re == nil || !re.MatchString(ref) {
			t.Errorf("%s: contract_ref %q has the wrong form for its class", filepath.Base(p), ref)
		}
		o := strings.Index(text, "## Observation")
		op := strings.Index(text, "## Options")
		no := strings.Index(text, "## Not observed")
		if o < 0 || op < o || no < op {
			t.Errorf("%s: body sections missing or out of order", filepath.Base(p))
		} else if items := regexp.MustCompile(`(?m)^\d+\. `).FindAllString(text[op:no], -1); len(items) < 2 {
			t.Errorf("%s: %d options, want at least 2", filepath.Base(p), len(items))
		}
	}

	// The contract_ref lines point at the tripped contract lines.
	if l := lineAt(t, cdata, neverLine); !strings.Contains(l, "internal/x/**") {
		t.Errorf("never line %d = %q", neverLine, l)
	}
	if l := lineAt(t, cdata, retriesLine); !strings.Contains(l, "audit_retries") {
		t.Errorf("audit_retries line %d = %q", retriesLine, l)
	}

	// A card holding only an (even open) revoke record is not needs-decision.
	for _, p := range paths[:len(paths)-1] {
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	nd, err := escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || nd {
		t.Errorf("NeedsDecision with only a revoke record = %v, %v; want false", nd, err)
	}
}

// AC-AE-023 (REQ-AE-019, REQ-AE-020): a repeat trip while open increments
// occurrences; a trip after resolution writes <class>-<fp>-2.md and leaves the
// resolved record's status and decider bytes unchanged; a growing budget count
// increments the one budget record.
func TestRecordDedupAndRetripAfterResolve(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	trip := escalation.Record{
		Card: "t9001", Spec: "SPEC-A-001", Kind: escalation.KindContract, Class: escalation.ClassOwnershipMove,
		Fingerprint: escalation.Fingerprint(escalation.ClassOwnershipMove, "internal/bar/x.go", "ownership.write"),
		ContractRef: "contract.yaml:17", EscalateOn: "ownership-move", HeadSHA: escalationtest.HeadSHA,
		Observation: "Write internal/bar/x.go", Options: []string{"Revert", "Amend"},
	}
	first, err := escalation.WriteRecord(w.Root, trip, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	second, err := escalation.WriteRecord(w.Root, trip, fixedNow.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if second != first || len(records(t, w)) != 1 {
		t.Fatalf("second trip wrote %s (records %v), want the same file", second, records(t, w))
	}
	r := mustParse(t, first)
	if r.Occurrences != 2 || r.DetectedAt != "2026-03-04T05:06:07Z" || r.UpdatedAt != "2026-03-04T05:07:07Z" {
		t.Errorf("after second trip: occurrences=%d detected=%s updated=%s", r.Occurrences, r.DetectedAt, r.UpdatedAt)
	}
	body := func(p string) string {
		data, _ := os.ReadFile(p)
		return string(data[strings.Index(string(data), "## Observation"):])
	}
	if !strings.Contains(body(first), "Write internal/bar/x.go") {
		t.Error("increment lost the record body")
	}

	// Resolve it by hand, as a decider would.
	data, _ := os.ReadFile(first)
	resolved := strings.Replace(strings.Replace(string(data), "status: open", "status: resolved", 1),
		"decider: \"\"", "decider: fixture-decider", 1)
	if err := os.WriteFile(first, []byte(resolved), 0o644); err != nil {
		t.Fatal(err)
	}
	third, err := escalation.WriteRecord(w.Root, trip, fixedNow.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if want := escalation.RecordPath(w.Root, "t9001", trip.Class, trip.Fingerprint, 2); third != want {
		t.Fatalf("third trip wrote %s, want %s", third, want)
	}
	after, _ := os.ReadFile(first)
	if !bytes.Equal(after, []byte(resolved)) {
		t.Error("resolved record bytes changed")
	}
	if n := mustParse(t, third); n.Status != escalation.StatusOpen || n.Occurrences != 1 || n.Decider != "" {
		t.Errorf("re-trip record = %+v", n)
	}
	// A fourth trip increments the open re-trip, not a third file.
	if p, _ := escalation.WriteRecord(w.Root, trip, fixedNow); p != third || mustParse(t, third).Occurrences != 2 {
		t.Errorf("fourth trip wrote %s", p)
	}

	// Budget: a higher observed count increments the one record.
	b := opRecord("t9001", escalation.ClassBudgetExceeded, "contract.yaml:40", "operations")
	b.Observation = "operations observed 41, limit 40"
	bp, err := escalation.WriteRecord(w.Root, b, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	b.Observation = "operations observed 42, limit 40"
	bp2, err := escalation.WriteRecord(w.Root, b, fixedNow)
	if err != nil || bp2 != bp || mustParse(t, bp).Occurrences != 2 {
		t.Errorf("budget recurrence wrote %s (%v), occurrences %d", bp2, err, mustParse(t, bp).Occurrences)
	}
	if n := len(records(t, w)); n != 3 {
		t.Errorf("records = %v, want 3 files", records(t, w))
	}
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustParse(t *testing.T, p string) escalation.Record {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	r, err := escalation.ParseRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
