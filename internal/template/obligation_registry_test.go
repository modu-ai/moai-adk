package template

import (
	"path/filepath"
	"strings"
	"testing"
)

// repoRootForRegistry is the module root, two levels above this package.
func repoRootForRegistry(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func registryResolver(t *testing.T) SourceResolver {
	t.Helper()
	root := repoRootForRegistry(t)
	cmds, err := IndexCLICommands(filepath.Join(root, "internal", "cli"))
	if err != nil {
		t.Fatalf("index cli commands: %v", err)
	}
	tests, err := IndexTestFunctions(filepath.Join(root, "internal"))
	if err != nil {
		t.Fatalf("index tests: %v", err)
	}
	return SourceResolver{Commands: cmds, TemplatesRoot: filepath.Join(root, "internal", "template", "templates"), Tests: tests}
}

func mustLoadRegistry(t *testing.T) *ObligationRegistry {
	t.Helper()
	reg, err := LoadObligationRegistry()
	if err != nil {
		t.Fatalf("load the embedded registry: %v", err)
	}
	return reg
}

// TestObligationCoverageRegistry is AC-HPR-018 on the whole-catalog registry
// (M2h, operator decision Q1): every required row resolves against the real
// CLI command set, the real template tree, and the real test index; and
// removing one row's Codex path or check fails naming that row.
func TestObligationCoverageRegistry(t *testing.T) {
	reg := mustLoadRegistry(t)
	r := registryResolver(t)

	if v := CheckObligationCoverage(reg, r); len(v) != 0 {
		t.Fatalf("the embedded registry has coverage violations:\n%v", v)
	}

	// The catalog carries every family the SPEC registers.
	var m1, unverified, unsupported int
	for _, o := range reg.Obligations {
		marker, value, _ := ParseApplicationPath(o.CodexPath)
		switch {
		case marker == MarkerBlocked && strings.HasPrefix(value, "M1"):
			m1++
		case marker == MarkerUnverified:
			unverified++
		}
		if cm, _, _ := ParseApplicationPath(o.ClaudePath); cm == MarkerUnsupported {
			unsupported++
		}
	}
	if m1 == 0 || unverified != 4 || unsupported == 0 {
		t.Fatalf("catalog families: blocked:M1=%d unverified=%d claude UNSUPPORTED=%d; want M1 rows, 4 non-Stop chains, and the Claude interrupt source", m1, unverified, unsupported)
	}

	for _, mutate := range []struct {
		name  string
		apply func(o *Obligation)
		want  string
	}{
		{"remove the Codex path", func(o *Obligation) { o.CodexPath = "" }, "codex_path is empty"},
		{"remove the check", func(o *Obligation) { o.Check = "" }, "check is empty"},
		{"point the check at a missing test", func(o *Obligation) { o.Check = "TestNoSuchParityCheck" }, "names no existing test"},
		{"point the Codex path at an unregistered command", func(o *Obligation) { o.CodexPath = "moai hook no-such-event --harness codex" }, "does not resolve"},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			mut := &ObligationRegistry{Obligations: append([]Obligation(nil), reg.Obligations...)}
			target := mut.Obligations[0].ID
			mutate.apply(&mut.Obligations[0])
			v := CheckObligationCoverage(mut, r)
			if len(v) != 1 || v[0].ID != target || !strings.Contains(v[0].Problem, mutate.want) {
				t.Fatalf("want one violation for %s mentioning %q, got %v", target, mutate.want, v)
			}
		})
	}
}

// TestNonStopChainsRecordedUnverified pins the four non-Stop multi-handler
// chains spec.md §F excludes: each is registered, required, and carries an
// unverified Codex marker, so the aggregate cannot read it as PASS.
func TestNonStopChainsRecordedUnverified(t *testing.T) {
	reg := mustLoadRegistry(t)
	want := map[string]bool{"chain-post-tool-use": false, "chain-session-start": false, "chain-subagent-stop": false, "chain-user-prompt-submit": false}
	for _, o := range reg.Obligations {
		if _, ok := want[o.ID]; !ok {
			continue
		}
		marker, _, _ := ParseApplicationPath(o.CodexPath)
		if !o.Required || marker != MarkerUnverified {
			t.Errorf("%s: required=%v codex marker=%q, want a required row marked unverified", o.ID, o.Required, marker)
		}
		want[o.ID] = true
	}
	for id, seen := range want {
		if !seen {
			t.Errorf("%s is not registered", id)
		}
	}
	verdicts := AggregateParityVerdict(reg, nil).PerObligation
	for _, v := range verdicts {
		if strings.HasPrefix(v.ID, "chain-") && v.Verdict != VerdictUnverified {
			t.Errorf("%s aggregates %s, want %s", v.ID, v.Verdict, VerdictUnverified)
		}
	}
}

// TestStandingPolicyParity is the registered check for the M1
// standing-policy obligations while M1 is open: each is required, blocked on
// M1, and aggregates as BLOCKED — reported honestly as not PASS until M1
// replaces this check with the real one.
func TestStandingPolicyParity(t *testing.T) {
	reg := mustLoadRegistry(t)
	n := 0
	for _, o := range reg.Obligations {
		if o.Check != "TestStandingPolicyParity" {
			continue
		}
		n++
		marker, value, _ := ParseApplicationPath(o.CodexPath)
		if !o.Required || marker != MarkerBlocked || !strings.HasPrefix(value, "M1") {
			t.Errorf("%s: required=%v codex=%q, want a required row blocked on M1", o.ID, o.Required, o.CodexPath)
		}
	}
	if n == 0 {
		t.Fatal("no M1 standing-policy row is registered")
	}
	for _, v := range AggregateParityVerdict(reg, nil).PerObligation {
		if strings.HasPrefix(v.ID, "policy-") && v.Verdict != VerdictBlocked {
			t.Errorf("%s aggregates %s, want %s", v.ID, v.Verdict, VerdictBlocked)
		}
	}
}

// fullAttribution is a complete attribution for an effect-verified record.
func fullAttribution() ParityRecord {
	return ParityRecord{Verdict: VerdictPass, EvidenceLevel: EvidenceEffectVerified,
		Commit: "abc123", TreeDigest: "d1", ClaudeVersion: "2.1.282", CodexVersion: "codex-cli 0.157.0", Uname: "darwin arm64"}
}

// TestParityVerdictAggregate is AC-HPR-019: the aggregate is PASS only when
// every required obligation is effect-verified with full attribution, reading
// both the go-test action and the verdict record and taking the weaker.
func TestParityVerdictAggregate(t *testing.T) {
	reg := &ObligationRegistry{Obligations: []Obligation{
		{ID: "a", Required: true, ClaudePath: "moai hook stop", CodexPath: "moai hook stop --harness codex", Check: "TestA", AC: "AC-1"},
		{ID: "b", Required: true, ClaudePath: "moai hook stop", CodexPath: "moai hook stop --harness codex", Check: "TestB", AC: "AC-2"},
		{ID: "opt", Required: false, ClaudePath: "x", CodexPath: "blocked:later", Check: "TestOpt", AC: "AC-3"},
	}}
	good := func() map[string]ObligationEvidence {
		ra, rb := fullAttribution(), fullAttribution()
		return map[string]ObligationEvidence{"a": {Action: ActionPass, Record: &ra}, "b": {Action: ActionPass, Record: &rb}}
	}

	if got := AggregateParityVerdict(reg, good()); got.Overall != VerdictPass {
		t.Fatalf("a fully effect-verified registry aggregates %s (%+v), want PASS", got.Overall, got.PerObligation)
	}

	for _, tc := range []struct {
		name string
		mut  func(ev map[string]ObligationEvidence, reg *ObligationRegistry)
		want string
	}{
		{"skip", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			e := ev["a"]
			e.Action = ActionSkip
			ev["a"] = e
		}, VerdictNotRun},
		{"empty run", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			e := ev["a"]
			e.Action = ""
			ev["a"] = e
		}, VerdictNotRun},
		{"go test failed", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			e := ev["a"]
			e.Action = ActionFail
			ev["a"] = e
		}, VerdictFail},
		{"NOT_RUN record", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) { ev["a"].Record.Verdict = VerdictNotRun }, VerdictNotRun},
		{"pass action with a NOT_RUN record", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			ev["a"].Record.Verdict = VerdictNotRun
		}, VerdictNotRun},
		{"UNSUPPORTED record", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			ev["a"].Record.Verdict = VerdictUnsupported
		}, VerdictUnsupported},
		{"UNSUPPORTED path", func(_ map[string]ObligationEvidence, reg *ObligationRegistry) {
			reg.Obligations[0].ClaudePath = "UNSUPPORTED:host cannot"
		}, VerdictUnsupported},
		{"blocked path", func(_ map[string]ObligationEvidence, reg *ObligationRegistry) {
			reg.Obligations[0].CodexPath = "blocked:M1"
		}, VerdictBlocked},
		{"unverified path", func(_ map[string]ObligationEvidence, reg *ObligationRegistry) {
			reg.Obligations[0].CodexPath = "unverified:excluded"
		}, VerdictUnverified},
		{"Stop gate capped unverified", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			ev["a"].Record.Verdict = VerdictUnverified
		}, VerdictUnverified},
		{"missing attribution field", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) { ev["a"].Record.CodexVersion = "" }, VerdictUnattributed},
		{"no record at all", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			e := ev["a"]
			e.Record = nil
			ev["a"] = e
		}, VerdictUnattributed},
		{"config-existence-only evidence", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			ev["a"].Record.EvidenceLevel = EvidenceRegistered
		}, VerdictInsufficient},
		{"fired but effect not verified", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) {
			ev["a"].Record.EvidenceLevel = EvidenceFired
		}, VerdictInsufficient},
		{"obligation absent from the evidence", func(ev map[string]ObligationEvidence, _ *ObligationRegistry) { delete(ev, "a") }, VerdictNotRun},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mreg := &ObligationRegistry{Obligations: append([]Obligation(nil), reg.Obligations...)}
			ev := good()
			tc.mut(ev, mreg)
			got := AggregateParityVerdict(mreg, ev)
			if got.Overall == VerdictPass {
				t.Fatalf("%s: aggregate read PASS", tc.name)
			}
			var a ObligationVerdict
			for _, v := range got.PerObligation {
				if v.ID == "a" {
					a = v
				}
			}
			if a.Verdict != tc.want {
				t.Fatalf("%s: obligation a aggregates %s (%s), want %s", tc.name, a.Verdict, a.Reason, tc.want)
			}
		})
	}

	t.Run("the embedded registry is not PASS from this SPEC alone", func(t *testing.T) {
		if got := AggregateParityVerdict(mustLoadRegistry(t), nil); got.Overall == VerdictPass {
			t.Fatal("the whole-catalog registry aggregated PASS with no evidence")
		}
	})
}

// TestReadGoTestActions checks rule P on a go test -json stream: a skip in a
// test or any of its subtests reads as skip, and a test with no record is
// absent (an empty run).
func TestReadGoTestActions(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"run","Test":"TestA"}`,
		`{"Action":"pass","Test":"TestA"}`,
		`{"Action":"run","Test":"TestB"}`,
		`{"Action":"skip","Test":"TestB/leg"}`,
		`{"Action":"pass","Test":"TestB"}`,
		`{"Action":"fail","Test":"TestC/leg"}`,
		`{"Action":"fail","Test":"TestC"}`,
		`{"Action":"skip","Test":"TestD"}`,
		`{"Action":"output","Output":"ok"}`,
	}, "\n")
	got, err := ReadGoTestActions(strings.NewReader(stream))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"TestA": ActionPass, "TestB": ActionSkip, "TestC": ActionFail, "TestD": ActionSkip}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	if _, ok := got["TestE"]; ok {
		t.Error("a test with no record must be absent")
	}
}
