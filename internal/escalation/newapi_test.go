package escalation_test

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

const fooBase = "package foo\n\nfunc Existing() {}\n"

// gitArmed builds a git-backed contract-mode worktree whose card base is
// main: the contract and a Go package are committed on main, the work
// branch WT-x is checked out, and the card is armed by a first observation.
func gitArmed(t *testing.T, card, detector string) (*escalationtest.Worktree, config.AutonomySettings) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	isolateStore(t)
	w := escalationtest.NewGitWorktree(t, card)
	extra := ""
	if detector != "" {
		extra = "escalation:\n  new_api_detector: \"" + detector + "\""
	}
	w.WriteContractMode(extra)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	w.Write("internal/foo/foo.go", fooBase)
	w.Commit("base")
	w.Git("checkout", "-q", "-b", "WT-x")
	s := contractSettings(t, w)
	preWrite(s, w, "internal/fixture/a.go")
	if !cardLog(t, w).Armed() {
		t.Fatalf("card did not arm: %+v", cardLog(t, w).Entries)
	}
	return w, s
}

// newAPIRecords returns the class 4 record bodies.
func newAPIRecords(t *testing.T, w *escalationtest.Worktree) []string {
	t.Helper()
	_, raw := recordsOfClass(t, w, escalation.ClassNewArchitectureOrAPI)
	return raw
}

// AC-AE-011 (REQ-AE-009): each Go fixture HEAD adding exactly one of the five
// addition kinds writes a record whose kind and name match; a HEAD adding
// only an unexported function writes none.
func TestNewAPIAdditionsTrip(t *testing.T) {
	cases := []struct {
		name     string
		change   func(w *escalationtest.Worktree)
		kind     string
		addition string
	}{
		{"exported-function", func(w *escalationtest.Worktree) {
			w.Write("internal/foo/foo.go", fooBase+"\nfunc Added() {}\n")
		}, escalation.AdditionExportedDecl, "Added"},
		{"new-package", func(w *escalationtest.Worktree) {
			w.Write("internal/bar/bar.go", "package bar\n\nfunc helper() {}\n")
		}, escalation.AdditionPackage, "internal/bar"},
		{"cli-verb", func(w *escalationtest.Worktree) {
			w.Write("internal/foo/foo.go", fooBase+"\nvar cmd = &cobra.Command{Use: \"frobnicate\"}\n")
		}, escalation.AdditionCLIVerb, "frobnicate"},
		{"mcp-tool", func(w *escalationtest.Worktree) {
			w.Write("internal/foo/foo.go", fooBase+"\nvar tool = mcp.NewTool(\"frob_tool\")\n")
		}, escalation.AdditionMCPTool, "frob_tool"},
		{"config-key", func(w *escalationtest.Worktree) {
			w.Write("internal/foo/foo.go", fooBase+"\ntype cfg struct {\n\tFrob string `yaml:\"frob_key\"`\n}\n")
		}, escalation.AdditionConfigKey, "frob_key"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, s := gitArmed(t, "t9001", "")
			tc.change(w)
			w.Commit(tc.name)
			res := escalation.Checkpoint(s, w.Root)

			found := false
			for _, a := range res.Additions {
				found = found || (a.Kind == tc.kind && a.Name == tc.addition)
			}
			if !found {
				t.Fatalf("additions = %+v, want %s %s", res.Additions, tc.kind, tc.addition)
			}
			recs := newAPIRecords(t, w)
			matched := 0
			for _, r := range recs {
				if strings.Contains(r, "kind: "+tc.kind) && strings.Contains(r, "name: "+tc.addition) {
					matched++
				}
			}
			if matched != 1 {
				t.Errorf("records matching %s %s = %d; records:\n%s", tc.kind, tc.addition, matched, strings.Join(recs, "\n---\n"))
			}
			for _, r := range recs {
				if !strings.Contains(r, "escalate_on: new-architecture-or-api") {
					t.Errorf("record lacks escalate_on:\n%s", r)
				}
			}
		})
	}

	t.Run("unexported-only", func(t *testing.T) {
		w, s := gitArmed(t, "t9001", "")
		w.Write("internal/foo/foo.go", fooBase+"\nfunc added() {}\n")
		w.Commit("unexported")
		res := escalation.Checkpoint(s, w.Root)
		if len(res.Additions) != 0 || len(newAPIRecords(t, w)) != 0 {
			t.Errorf("unexported function reported: %+v / %d records", res.Additions, len(newAPIRecords(t, w)))
		}
	})
}

// AC-AE-011 negative fixtures (REQ-AE-022): off writes no class 4 record; an
// unresolvable card base writes none and lists the card base under
// not_observed; a Python HEAD reports each addition under a supported
// sub-kind or not_observed, with the CLI-verb, MCP-tool, and config-key
// sub-kinds listed not_observed.
func TestNewAPINotObservedCases(t *testing.T) {
	t.Run("off", func(t *testing.T) {
		w, s := gitArmed(t, "t9001", "off")
		if s.NewAPIDetector != "off" {
			t.Fatalf("fixture detector = %q", s.NewAPIDetector)
		}
		w.Write("internal/foo/foo.go", fooBase+"\nfunc Added() {}\n")
		w.Commit("exported")
		res := escalation.Checkpoint(s, w.Root)
		if len(newAPIRecords(t, w)) != 0 || len(res.Additions) != 0 {
			t.Errorf("off wrote class 4: %+v", res)
		}
	})

	t.Run("no-integration-branch", func(t *testing.T) {
		w, s := gitArmed(t, "t9001", "")
		w.Git("branch", "-m", "main", "trunk")
		w.Write("internal/foo/foo.go", fooBase+"\nfunc Added() {}\n")
		w.Commit("exported")
		res := escalation.Checkpoint(s, w.Root)
		if len(newAPIRecords(t, w)) != 0 {
			t.Errorf("records written without a card base")
		}
		if !slices.ContainsFunc(res.NotObserved, func(s string) bool { return strings.Contains(s, "card base") }) {
			t.Errorf("not_observed = %v, want the card base", res.NotObserved)
		}
		no := notObservedEntries(t, w)
		if len(no) == 0 || !strings.Contains(strings.Join(no[len(no)-1].NotObserved, " "), "card base") {
			t.Errorf("card log carries no card-base not-observed line: %+v", no)
		}
	})

	t.Run("python", func(t *testing.T) {
		w, s := gitArmed(t, "t9001", "")
		w.Write("tool/cli.py", "def main():\n    return 0\n")
		w.Write("pyproject.toml", "[project.scripts]\ntool = \"tool.cli:main\"\n")
		w.Commit("python")
		res := escalation.Checkpoint(s, w.Root)
		for _, sub := range []string{escalation.AdditionCLIVerb, escalation.AdditionMCPTool, escalation.AdditionConfigKey} {
			want := sub + " (python)"
			if !slices.Contains(res.NotObserved, want) {
				t.Errorf("not_observed = %v, want %q", res.NotObserved, want)
			}
		}
		// The new package directory is observed; the module's main is either
		// an observed exported declaration or listed not-observed.
		kinds := map[string]bool{}
		for _, a := range res.Additions {
			kinds[a.Kind+":"+a.Name] = true
		}
		if !kinds[escalation.AdditionPackage+":tool"] {
			t.Errorf("additions = %+v, want new package tool", res.Additions)
		}
		declObserved := kinds[escalation.AdditionExportedDecl+":main"]
		declListed := slices.ContainsFunc(res.NotObserved, func(s string) bool {
			return strings.Contains(s, escalation.AdditionExportedDecl) && strings.Contains(s, "tool/cli.py")
		})
		t.Logf("python main: observed as exported declaration=%v, listed not-observed=%v", declObserved, declListed)
		if declObserved == declListed {
			t.Errorf("main: observed=%v listed-not-observed=%v, want exactly one", declObserved, declListed)
		}
		for _, r := range newAPIRecords(t, w) {
			if !strings.Contains(r, "cli-verb (python)") {
				t.Errorf("record not_observed lacks the python sub-kinds:\n%s", r)
			}
		}
	})
}
