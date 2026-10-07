package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The tree run of the github-flow sweep guard (SPEC-GITHUB-FLOW-DEFAULT-001 design D-9 / D-27,
// AC-GFD-011 and AC-GFD-021; card t1453 M4 step 1).
//
// The guard sweeps the scoped doc / rule / agent / skill / script surface for live text that
// still names `develop` as this project's base. It is ARMED only when this repository's
// git_strategy workflow resolves to github-flow (M5 flips the value); until then it prints a
// "disarmed" line and asserts only what is true on either branch model — that it visited
// the surface (the numeric floors), and that the allow list is well formed and has no stale
// entry. It never t.Skip()s: a skipped guard is indistinguishable from a green one.
//
// Two environment hooks let an operator demonstrate the engine on the real tree without
// editing any doc: MOAI_SWEEP_GUARD_ARMED=1 forces the armed assertions, and
// MOAI_SWEEP_GUARD_REPORT=<path> writes every finding as TSV.

const (
	sweepArmEnv    = "MOAI_SWEEP_GUARD_ARMED"
	sweepReportEnv = "MOAI_SWEEP_GUARD_REPORT"

	// sweepAllowCap is the shared cap of the allow list, strong and weak entries together
	// (design D-9 / D-27). Raising it needs a decision row in design.md.
	sweepAllowCap = 40
)

// sweepTotalFloor and the per-subtree floors are floor(0.9 x N) of the population counted
// with `git ls-files` at this card's M4-1 start (HEAD f33d1f19e); design D-9 derives its own
// table (1525 files, floor 1372) the same way at plan time. The two differ because the tree
// grew (skills +2, templates +2, scripts +4) — the rule is the contract, the numbers are
// its output.
const sweepTotalFloor = 1379

// sweepSurface is the scoped surface (design D-9, ten subtrees; `.moai/reports/**` and
// `.moai/specs/**` are out of scope — history and SPEC bodies).
func sweepSurface() []sweepSubtree {
	return []sweepSubtree{
		{Name: "root-docs", Files: []string{"AGENTS.md", "AGENTS.local.md", "CLAUDE.md", "README.md", "README.ko.md", "README.ja.md", "README.zh.md"}, Floor: 6},
		{Name: "rules", Dir: ".claude/rules", Exts: []string{".md"}, Floor: 101},
		{Name: "agents", Dir: ".claude/agents", Exts: []string{".md"}, Floor: 19},
		{Name: "skills", Dir: ".claude/skills", Exts: []string{".md"}, Floor: 231},
		{Name: "output-styles", Dir: ".claude/output-styles", Exts: []string{".md"}, Floor: 2},
		{Name: "docs-site", Dir: "docs-site/content", Exts: []string{".md"}, Floor: 558},
		{Name: "moai-docs", Dir: ".moai/docs", Exts: []string{".md"}, Floor: 33},
		{Name: "templates", Dir: "internal/template/templates", Exts: []string{".md", ".md.tmpl"}, Floor: 369},
		{Name: "codemaps", Dir: ".moai/project/codemaps", Exts: []string{".md"}, NonRecursive: true, Floor: 5},
		{Name: "scripts", Dir: "scripts", Exts: []string{".sh", ".py"}, Floor: 52},
	}
}

// sweepExpectedSubtrees is the literal set of subtrees the guard must sweep. It is asserted as
// set equality, separately from the floors: the two smallest subtrees (output-styles, codemaps)
// fit inside the total floor's slack, so dropping one from sweepSurface() would otherwise move
// no assertion at all.
var sweepExpectedSubtrees = []string{"agents", "codemaps", "docs-site", "moai-docs", "output-styles", "root-docs", "rules", "scripts", "skills", "templates"}

// sweepGuardAllow is the real allow list. It is empty at M4 step 1: the entries are reviewed
// against the swept tree at M4 start, after the t1399 rename lands, and every entry is one
// whole line with a reviewed reason.
var sweepGuardAllow = []sweepAllow{}

// sweepGuardCeilings are the ratchet ceilings (plan-audit D18 and design D-9's P-C family).
// They were read off the real tree at this card's M4 step 1 (HEAD f33d1f19e plus the guard
// files; the armed demonstration printed marker-exempted lines=11 (H1 0) and P-C lines=131)
// and may only fall: a marker that silences more live text than the baseline, a document-wide
// H1 marker, or a new token-less live sentence fails the armed run. M4/M5 tighten them to
// the swept tree's values at the armed run; raising one needs a decision row in design.md.
var sweepGuardCeilings = sweepCeilings{MarkerLines: 11, MarkerH1Lines: 0, PCLines: 131}

// sweepRun is one walk over the surface.
type sweepRun struct {
	Visited  map[string]int
	Area     map[string]string // file -> subtree name
	Findings []sweepFinding
}

// sweepWalk reads every in-scope file under root and classifies it. A subtree whose
// directory cannot be read is an error, never a silent zero.
func sweepWalk(root string, subtrees []sweepSubtree, allow []sweepAllow) (*sweepRun, error) {
	run := &sweepRun{Visited: map[string]int{}, Area: map[string]string{}}
	visit := func(s sweepSubtree, rel string) error {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		run.Visited[s.Name]++
		run.Area[rel] = s.Name
		run.Findings = append(run.Findings, sweepScan(rel, string(raw), allow)...)
		return nil
	}
	for _, s := range subtrees {
		run.Visited[s.Name] += 0
		if len(s.Files) > 0 {
			for _, rel := range s.Files {
				if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
					continue // a listed file that a checkout lacks is a floor matter, not an error
				}
				if err := visit(s, rel); err != nil {
					return nil, err
				}
			}
			continue
		}
		base := filepath.Join(root, filepath.FromSlash(s.Dir))
		info, err := os.Stat(base)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("subtree %s: %s is not a readable directory: %v", s.Name, s.Dir, err)
		}
		err = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if s.NonRecursive && path != base {
					return filepath.SkipDir
				}
				return nil
			}
			for _, ext := range s.Exts {
				if strings.HasSuffix(d.Name(), ext) {
					rel, err := filepath.Rel(root, path)
					if err != nil {
						return err
					}
					return visit(s, filepath.ToSlash(rel))
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("subtree %s: %w", s.Name, err)
		}
	}
	return run, nil
}

// sweepGuardArmed reports whether the armed assertions apply, and why.
func sweepGuardArmed(root string) (bool, string) {
	if os.Getenv(sweepArmEnv) == "1" {
		return true, "forced by " + sweepArmEnv + "=1"
	}
	cfg := config.LoadGitFlowIntegrationConfig(root)
	return cfg.Workflow == config.WorkflowGitHubFlow,
		fmt.Sprintf("git_strategy workflow=%q (armed only when %q)", cfg.Workflow, config.WorkflowGitHubFlow)
}

// sweepSummary is the by-area tally the guard prints and the demonstration reports.
type sweepSummary struct {
	Strong, Weak, Exempt, Allowed int
	ByArea                        map[string][3]int // strong, weak, exempt
	StrongByFile                  map[string]int
}

func sweepSummarize(run *sweepRun) sweepSummary {
	s := sweepSummary{ByArea: map[string][3]int{}, StrongByFile: map[string]int{}}
	for _, f := range run.Findings {
		a := s.ByArea[run.Area[f.File]]
		switch f.Class {
		case sweepStrong:
			s.Strong++
			a[0]++
			s.StrongByFile[f.File]++
		case sweepWeak:
			s.Weak++
			a[1]++
		case sweepExempt:
			s.Exempt++
			a[2]++
			if f.Rule == "allow" {
				s.Allowed++
			}
		}
		s.ByArea[run.Area[f.File]] = a
	}
	return s
}

func TestGitHubFlowSweepGuard(t *testing.T) {
	root := filepath.Join("..", "..")
	surface := sweepSurface()

	if err := sweepValidateAllow(sweepGuardAllow, sweepAllowCap); err != nil {
		t.Fatalf("allow list invalid: %v", err)
	}
	run, err := sweepWalk(root, surface, sweepGuardAllow)
	if err != nil {
		t.Fatalf("sweep walk: %v", err)
	}
	armed, why := sweepGuardArmed(root)

	total := 0
	for _, s := range surface {
		total += run.Visited[s.Name]
	}
	sum := sweepSummarize(run)
	counts := sweepCountsOf(run.Findings)
	used := map[int]bool{}
	for _, f := range run.Findings {
		if f.Class == sweepExempt && f.Rule == "allow" {
			used[f.AllowIdx] = true
		}
	}

	if armed {
		t.Logf("sweep guard ARMED: %s", why)
	} else {
		t.Logf("sweep guard disarmed: %s — violation assertions are not made on this tree", why)
	}
	t.Logf("visited=%d (floor %d)", total, sweepTotalFloor)
	for _, s := range surface {
		t.Logf("subtree %-13s visited=%d floor=%d", s.Name, run.Visited[s.Name], s.Floor)
	}
	t.Logf("weak hits: allowed=%d remaining=%d; strong remaining=%d; allow entries used=%d/%d (cap %d)",
		sum.Allowed, sum.Weak, sum.Strong, len(used), len(sweepGuardAllow), sweepAllowCap)
	t.Logf("marker-exempted lines=%d (H1 %d) ceilings=%d/%d; P-C lines=%d ceiling=%d",
		counts.MarkerLines, counts.MarkerH1Lines, sweepGuardCeilings.MarkerLines, sweepGuardCeilings.MarkerH1Lines,
		counts.PCLines, sweepGuardCeilings.PCLines)

	if path := os.Getenv(sweepReportEnv); path != "" {
		if err := sweepWriteReport(path, run, sum, counts); err != nil {
			t.Errorf("write report %s: %v", path, err)
		} else {
			t.Logf("full finding list written to %s", path)
		}
	}

	// True on either branch model: the sweep must have swept, and the allow list must be live.
	var names []string
	for _, s := range surface {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != strings.Join(sweepExpectedSubtrees, ",") {
		t.Errorf("swept subtrees = %v, want exactly %v — a dropped subtree stays inside the total floor's slack", names, sweepExpectedSubtrees)
	}
	for _, p := range sweepVisitProblems(run.Visited, surface, sweepTotalFloor) {
		t.Errorf("visit: %s", p)
	}
	for i, e := range sweepGuardAllow {
		if !used[i] {
			t.Errorf("stale allow entry [%d] %s %q: no would-be violation matched that whole line", i, e.File, e.Literal)
		}
	}

	// The arming switch reads the repository's git_strategy workflow through the config loader.
	// M5 flips this repository's value, so the switch is exercised here on synthetic roots in
	// both directions — a guard that can never arm would stay green forever.
	t.Run("arming-switch", func(t *testing.T) {
		t.Setenv(sweepArmEnv, "")
		write := func(workflow string) string {
			dir := t.TempDir()
			sections := filepath.Join(dir, ".moai", "config", "sections")
			if err := os.MkdirAll(sections, 0o755); err != nil {
				t.Fatal(err)
			}
			if workflow != "" {
				body := "git_strategy:\n  mode: manual\n  manual:\n    workflow: " + workflow + "\n"
				if err := os.WriteFile(filepath.Join(sections, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			return dir
		}
		if on, _ := sweepGuardArmed(write("github-flow")); !on {
			t.Error("a github-flow root did not arm the guard")
		}
		if on, _ := sweepGuardArmed(write("git-flow")); on {
			t.Error("a git-flow root armed the guard")
		}
		if on, _ := sweepGuardArmed(write("")); on {
			t.Error("a root with no git-strategy.yaml armed the guard")
		}
		t.Setenv(sweepArmEnv, "1")
		if on, _ := sweepGuardArmed(write("git-flow")); !on {
			t.Error("the " + sweepArmEnv + " override did not arm the guard")
		}
	})

	// F-empty through the real walker: a root whose subtrees exist but hold nothing visits
	// zero files, and the same verdict function the guard uses must fail it; a root missing a
	// subtree is an error, not a silent zero.
	t.Run("F-empty-walk", func(t *testing.T) {
		empty := t.TempDir()
		for _, s := range surface {
			if s.Dir != "" {
				if err := os.MkdirAll(filepath.Join(empty, filepath.FromSlash(s.Dir)), 0o755); err != nil {
					t.Fatal(err)
				}
			}
		}
		er, err := sweepWalk(empty, surface, nil)
		if err != nil {
			t.Fatalf("walk of an empty tree: %v", err)
		}
		if probs := sweepVisitProblems(er.Visited, surface, sweepTotalFloor); len(probs) == 0 {
			t.Error("an empty sweep passed the visit verdict")
		}
		if _, err := sweepWalk(t.TempDir(), surface, nil); err == nil {
			t.Error("a root with no subtree directories was walked without error")
		}
	})

	if !armed {
		return
	}
	var viols []string
	for _, f := range run.Findings {
		if f.violation() {
			viols = append(viols, f.String())
		}
	}
	if len(viols) > 0 {
		show := viols
		if len(show) > 40 {
			show = show[:40]
		}
		t.Errorf("%d live develop-base line(s) remain (%d strong, %d weak); first %d:\n  %s",
			len(viols), sum.Strong, sum.Weak, len(show), strings.Join(show, "\n  "))
	}
	for _, p := range sweepRatchetProblems(counts, sweepGuardCeilings) {
		t.Errorf("ratchet: %s", p)
	}
}

func sweepWriteReport(path string, run *sweepRun, sum sweepSummary, counts sweepCounts) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# strong=%d weak=%d exempt=%d (allow %d) marker=%d marker-h1=%d P-C=%d\n",
		sum.Strong, sum.Weak, sum.Exempt, sum.Allowed, counts.MarkerLines, counts.MarkerH1Lines, counts.PCLines)
	areas := make([]string, 0, len(sum.ByArea))
	for a := range sum.ByArea {
		areas = append(areas, a)
	}
	sort.Strings(areas)
	for _, a := range areas {
		v := sum.ByArea[a]
		fmt.Fprintf(&b, "# area %-13s strong=%d weak=%d exempt=%d visited=%d\n", a, v[0], v[1], v[2], run.Visited[a])
	}
	type kv struct {
		File string
		N    int
	}
	var top []kv
	for f, n := range sum.StrongByFile {
		top = append(top, kv{f, n})
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].N != top[j].N {
			return top[i].N > top[j].N
		}
		return top[i].File < top[j].File
	})
	for i, e := range top {
		if i == 10 {
			break
		}
		fmt.Fprintf(&b, "# top-strong %2d %s %d\n", i+1, e.File, e.N)
	}
	for _, f := range run.Findings {
		fmt.Fprintf(&b, "%s\t%s\t%s:%d\t%s\n", f.Class, f.Rule, f.File, f.Line, strings.TrimSpace(f.Text))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
