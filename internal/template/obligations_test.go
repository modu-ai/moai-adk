package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeCoverageResolver resolves tests and paths from fixed sets.
type fakeCoverageResolver struct {
	tests map[string]bool
	paths map[string]bool // "<harness>:<path>"
}

func (f fakeCoverageResolver) TestExists(name string) bool { return f.tests[name] }
func (f fakeCoverageResolver) PathExists(harness, path string) bool {
	return f.paths[harness+":"+path]
}

const obligationFixture = `obligations:
  - id: stop-goal-continuation
    required: true
    claude_path: moai hook stop-goal
    codex_path: moai hook stop --harness codex
    check: TestStopChainEffectParityGolden
    ac: AC-HPR-002
  - id: claude-interrupt-cancellation
    required: true
    claude_path: "UNSUPPORTED:design.md §D6 — Claude Stop does not fire on user interrupt"
    codex_path: moai hook interrupt --harness codex
    check: TestGoalCancellationPrecedence
    ac: AC-HPR-013
  - id: standing-policy-row
    required: true
    claude_path: moai hook pre-tool
    codex_path: "blocked:M1"
    check: TestStandingPolicyParity
    ac: AC-POL-01
  - id: non-stop-multi-handler-chains
    required: true
    claude_path: moai hook post-tool
    codex_path: "unverified:spec.md §F exclusion"
    check: TestNonStopChainsRecordedUnverified
    ac: AC-HPR-018
`

func fixtureResolver() fakeCoverageResolver {
	return fakeCoverageResolver{
		tests: map[string]bool{
			"TestStopChainEffectParityGolden":     true,
			"TestGoalCancellationPrecedence":      true,
			"TestStandingPolicyParity":            true,
			"TestNonStopChainsRecordedUnverified": true,
		},
		paths: map[string]bool{
			"claude:moai hook stop-goal":                true,
			"codex:moai hook stop --harness codex":      true,
			"codex:moai hook interrupt --harness codex": true,
			"claude:moai hook pre-tool":                 true,
			"claude:moai hook post-tool":                true,
		},
	}
}

func mustParseObligations(t *testing.T, doc string) *ObligationRegistry {
	t.Helper()
	reg, err := ParseObligations([]byte(doc))
	if err != nil {
		t.Fatalf("ParseObligations: %v", err)
	}
	return reg
}

// TestObligationCoverage is AC-HPR-018's coverage leg
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001 REQ-HPR-020/021, design.md §D4). M2a
// covers the schema, loader, and check; the whole-catalog registry file and
// the production resolver land in M2h.
func TestObligationCoverage(t *testing.T) {
	t.Run("a complete registry passes", func(t *testing.T) {
		reg := mustParseObligations(t, obligationFixture)
		if len(reg.Obligations) != 4 {
			t.Fatalf("parsed %d obligations, want 4", len(reg.Obligations))
		}
		if v := CheckObligationCoverage(reg, fixtureResolver()); len(v) != 0 {
			t.Fatalf("complete registry reported violations: %v", v)
		}
	})

	mutations := []struct {
		name   string
		id     string
		mutate func(o *Obligation)
		want   string
	}{
		{"remove the Codex path", "stop-goal-continuation", func(o *Obligation) { o.CodexPath = "" }, "codex_path"},
		{"remove the Claude path", "stop-goal-continuation", func(o *Obligation) { o.ClaudePath = "" }, "claude_path"},
		{"remove the check", "stop-goal-continuation", func(o *Obligation) { o.Check = "" }, "check"},
		{"name a check that does not exist", "standing-policy-row", func(o *Obligation) { o.Check = "TestNoSuchCheck" }, "TestNoSuchCheck"},
		{"name a Codex path that does not resolve", "claude-interrupt-cancellation", func(o *Obligation) { o.CodexPath = "moai hook no-such" }, "moai hook no-such"},
	}
	for _, m := range mutations {
		t.Run(m.name+" fails naming the obligation", func(t *testing.T) {
			reg := mustParseObligations(t, obligationFixture)
			found := false
			for i := range reg.Obligations {
				if reg.Obligations[i].ID == m.id {
					m.mutate(&reg.Obligations[i])
					found = true
				}
			}
			if !found {
				t.Fatalf("mutation target %s absent", m.id)
			}
			v := CheckObligationCoverage(reg, fixtureResolver())
			if len(v) != 1 || v[0].ID != m.id || !strings.Contains(v[0].Problem, m.want) {
				t.Fatalf("want one violation for %s mentioning %q, got %v", m.id, m.want, v)
			}
			if !strings.Contains(v[0].String(), m.id) {
				t.Fatalf("violation text does not name the obligation: %q", v[0].String())
			}
		})
	}

	t.Run("an optional obligation is not held to coverage", func(t *testing.T) {
		reg := mustParseObligations(t, obligationFixture)
		reg.Obligations[0].Required = false
		reg.Obligations[0].CodexPath = ""
		if v := CheckObligationCoverage(reg, fixtureResolver()); len(v) != 0 {
			t.Fatalf("optional obligation reported: %v", v)
		}
	})

	t.Run("an empty registry never passes", func(t *testing.T) {
		reg := mustParseObligations(t, "obligations: []\n")
		if v := CheckObligationCoverage(reg, fixtureResolver()); len(v) == 0 {
			t.Fatal("an empty registry passed coverage vacuously")
		}
	})
}

// TestObligationCoverageSchema pins the registry row schema and the loader's
// refusals.
func TestObligationCoverageSchema(t *testing.T) {
	t.Run("markers parse with their reference", func(t *testing.T) {
		cases := map[string]struct {
			marker PathMarker
			value  string
		}{
			"moai hook stop-goal":      {MarkerNone, "moai hook stop-goal"},
			"UNSUPPORTED:evidence-ref": {MarkerUnsupported, "evidence-ref"},
			"blocked:M1":               {MarkerBlocked, "M1"},
			"unverified:spec §F":       {MarkerUnverified, "spec §F"},
		}
		for in, want := range cases {
			m, v, err := ParseApplicationPath(in)
			if err != nil || m != want.marker || v != want.value {
				t.Errorf("ParseApplicationPath(%q) = %q, %q, %v; want %q, %q", in, m, v, err, want.marker, want.value)
			}
		}
		for _, bad := range []string{"UNSUPPORTED:", "blocked:  ", "unverified:"} {
			if _, _, err := ParseApplicationPath(bad); err == nil {
				t.Errorf("ParseApplicationPath(%q) accepted a marker without a reference", bad)
			}
		}
	})

	refusals := map[string]string{
		"unknown field":            "obligations:\n  - id: a\n    required: true\n    claude_path: x\n    codex_path: y\n    check: TestA\n    ac: AC-1\n    requires: [z]\n",
		"missing id":               "obligations:\n  - required: true\n    claude_path: x\n    codex_path: y\n    check: TestA\n    ac: AC-1\n",
		"missing required":         "obligations:\n  - id: a\n    claude_path: x\n    codex_path: y\n    check: TestA\n    ac: AC-1\n",
		"duplicate id":             "obligations:\n  - id: a\n    required: true\n    claude_path: x\n    codex_path: y\n    check: TestA\n    ac: AC-1\n  - id: a\n    required: false\n    claude_path: x\n    codex_path: y\n    check: TestA\n    ac: AC-1\n",
		"marker without reference": "obligations:\n  - id: a\n    required: true\n    claude_path: x\n    codex_path: \"blocked:\"\n    check: TestA\n    ac: AC-1\n",
		"missing ac":               "obligations:\n  - id: a\n    required: true\n    claude_path: x\n    codex_path: y\n    check: TestA\n",
	}
	for name, doc := range refusals {
		t.Run("loader refuses "+name, func(t *testing.T) {
			if _, err := ParseObligations([]byte(doc)); err == nil {
				t.Fatalf("ParseObligations accepted a registry with %s", name)
			}
		})
	}

	t.Run("the test index finds test functions statically", func(t *testing.T) {
		root := t.TempDir()
		pkg := filepath.Join(root, "internal", "p")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package p\n\nimport \"testing\"\n\nfunc TestPresent(t *testing.T) {}\n\nfunc helperNotATest() {}\n"
		if err := os.WriteFile(filepath.Join(pkg, "p_test.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		// A non-test file declaring a Test-shaped function must not count.
		if err := os.WriteFile(filepath.Join(pkg, "p.go"), []byte("package p\n\nfunc TestLookalike() {}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		idx, err := IndexTestFunctions(root)
		if err != nil {
			t.Fatal(err)
		}
		if !idx.TestExists("TestPresent") {
			t.Error("TestPresent not indexed")
		}
		if idx.TestExists("TestLookalike") || idx.TestExists("helperNotATest") {
			t.Error("index accepted a function that is not a test")
		}
	})
}
