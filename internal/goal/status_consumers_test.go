package goal

import (
	"context"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// goalPkgPath is the import path of this package, used by the consumer scan to
// recognise references that resolve here.
const goalPkgPath = "github.com/modu-ai/moai-adk/internal/goal"

// statusConsumerScanPatterns is the scan scope fixed by SPEC-DUAL-HARNESS-HOOK-PARITY-001
// design.md §D6: the non-test files of these three packages.
var statusConsumerScanPatterns = []string{"./internal/goal", "./internal/cli", "./internal/hook"}

// statusConsumerSite is one row of the design.md §D6 consumer table: a
// top-level declaration that reads or writes goal status, or calls ClearGoal,
// together with the exact set of goal identifiers it references. Pinning the
// set (not only the site) is what asserts the internal/hook rows unchanged:
// any new status reference there changes the set and fails the scan.
type statusConsumerSite struct {
	site     string   // "<repo-relative file>|<top-level declaration>"
	idents   []string // sorted goal identifiers referenced at the site
	handling string   // design.md §D6 "required handling of cancelled"
}

// expectedStatusConsumers is the design.md §D6 table in machine form.
var expectedStatusConsumers = []statusConsumerSite{
	{"internal/goal/schema.go|const StatusArmed", []string{"Status"}, "the status set; carries StatusCancelled"},
	{"internal/goal/schema.go|type Goal", []string{"Status"}, "field type only"},
	{"internal/goal/schema.go|NewGoal", []string{"Goal.Status", "StatusArmed"}, "writes armed on construction"},
	{"internal/goal/schema.go|IsKnownStatus", []string{"Status", "StatusArmed", "StatusCancelled", "StatusCeilingExit", "StatusCleared", "StatusSatisfied", "StatusUnsatisfiable"}, "the recognised-status predicate; lists cancelled"},
	{"internal/goal/evaluate.go|(*Eval).Evaluate", []string{"Goal.Status", "StatusArmed", "StatusCancelled", "StatusCeilingExit", "StatusCleared", "StatusSatisfied", "StatusUnsatisfiable"}, "cancelled joins the early-return set; writers are unreachable for it; unknown → diagnostic"},
	{"internal/goal/dashboard.go|buildDashboardModel", []string{"Goal.Status"}, "renders cancelled verbatim"},
	{"internal/cli/launcher_blockcap_infinite.go|injectStopHookBlockCapForGoal", []string{"Goal.Status", "StatusArmed"}, "acts only on armed; cancelled excluded"},
	{"internal/cli/handoff.go|newHandoffSaveCmd", []string{"Goal.Status", "StatusArmed"}, "embeds only armed; cancelled excluded"},
	{"internal/cli/goal.go|runGoalStatusAll", []string{"Goal.Status"}, "renders cancelled verbatim"},
	{"internal/cli/goal.go|runGoalClear", []string{"ClearGoal"}, "removes goal state; writes no status"},
	{"internal/cli/goal.go|printGoalHuman", []string{"Goal.Status"}, "renders cancelled verbatim"},
	{"internal/hook/session_start_compact.go|renderCompactReinject", []string{"Goal.Status", "StatusArmed"}, "acts only on armed; internal/hook untouched (REQ-7)"},
	{"internal/hook/handoff_inject.go|rearmEmbeddedGoal", []string{"Goal.Status", "StatusArmed"}, "writes armed; internal/hook untouched (REQ-7)"},
	{"internal/hook/stop_failure.go|disarmGoalOnUnrecoverable", []string{"ClearGoal"}, "failure path, not cancellation; internal/hook untouched (REQ-7)"},
}

// moduleRoot walks up from the test's working directory to the go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

// scanStatusConsumers type-checks the scan scope and returns, per top-level
// declaration, the goal identifiers it references: constants of type Status,
// the Status type itself, the Goal.Status field, and ClearGoal. A textual grep
// cannot tell g.Status on a goal from the dozens of unrelated .Status fields in
// internal/cli, which is why this is type-driven.
func scanStatusConsumers(t *testing.T, root string) map[string][]string {
	t.Helper()
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   root,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, statusConsumerScanPatterns...)
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}
	if len(pkgs) != len(statusConsumerScanPatterns) {
		t.Fatalf("packages.Load returned %d packages, want %d", len(pkgs), len(statusConsumerScanPatterns))
	}

	var statusField types.Object
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			t.Fatalf("package %s has load errors: %v", p.PkgPath, p.Errors)
		}
		if p.PkgPath != goalPkgPath {
			continue
		}
		obj := p.Types.Scope().Lookup("Goal")
		st, ok := obj.Type().Underlying().(*types.Struct)
		if !ok {
			t.Fatal("goal.Goal is not a struct")
		}
		for i := 0; i < st.NumFields(); i++ {
			if st.Field(i).Name() == "Status" {
				statusField = st.Field(i)
			}
		}
	}
	if statusField == nil {
		t.Fatal("goal.Goal.Status field not found; the scan would be vacuous")
	}

	classify := func(obj types.Object) string {
		if obj == nil {
			return ""
		}
		if obj == statusField {
			return "Goal.Status"
		}
		if obj.Pkg() == nil || obj.Pkg().Path() != goalPkgPath {
			return ""
		}
		switch o := obj.(type) {
		case *types.Const:
			if named, ok := o.Type().(*types.Named); ok && named.Obj().Name() == "Status" {
				return o.Name()
			}
		case *types.TypeName:
			if o.Name() == "Status" {
				return "Status"
			}
		case *types.Func:
			if o.Name() == "ClearGoal" {
				return "ClearGoal"
			}
		}
		return ""
	}

	sites := map[string]map[string]bool{}
	for _, p := range pkgs {
		for _, file := range p.Syntax {
			rel, err := filepath.Rel(root, p.Fset.Position(file.Pos()).Filename)
			if err != nil {
				t.Fatal(err)
			}
			rel = filepath.ToSlash(rel)
			for _, decl := range file.Decls {
				name := declName(decl)
				ast.Inspect(decl, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if !ok {
						return true
					}
					if label := classify(p.TypesInfo.Uses[id]); label != "" {
						key := rel + "|" + name
						if sites[key] == nil {
							sites[key] = map[string]bool{}
						}
						sites[key][label] = true
					}
					return true
				})
			}
		}
	}
	out := make(map[string][]string, len(sites))
	for k, set := range sites {
		var ids []string
		for id := range set {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		out[k] = ids
	}
	return out
}

// declName names a top-level declaration: "Func", "(*Recv).Method",
// "type Name", "const FirstName", or "var FirstName".
func declName(decl ast.Decl) string {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv == nil || len(d.Recv.List) == 0 {
			return d.Name.Name
		}
		recv := d.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			if id, ok := star.X.(*ast.Ident); ok {
				return "(*" + id.Name + ")." + d.Name.Name
			}
		}
		if id, ok := recv.(*ast.Ident); ok {
			return id.Name + "." + d.Name.Name
		}
		return d.Name.Name
	case *ast.GenDecl:
		kw := d.Tok.String()
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				return kw + " " + s.Name.Name
			case *ast.ValueSpec:
				if len(s.Names) > 0 {
					return kw + " " + s.Names[0].Name
				}
			}
		}
		return kw
	}
	return "?"
}

// diffStatusConsumers compares the expected table against a scan and returns
// one line per disagreement: an unlisted site, a listed site that no longer
// exists, or a site whose identifier set changed.
func diffStatusConsumers(expected []statusConsumerSite, observed map[string][]string) []string {
	want := map[string][]string{}
	for _, row := range expected {
		want[row.site] = row.idents
	}
	var problems []string
	for site, ids := range observed {
		w, ok := want[site]
		if !ok {
			problems = append(problems, "unlisted goal-status site "+site+" references "+strings.Join(ids, ","))
			continue
		}
		if strings.Join(w, ",") != strings.Join(ids, ",") {
			problems = append(problems, "site "+site+" references "+strings.Join(ids, ",")+", table lists "+strings.Join(w, ","))
		}
	}
	for site := range want {
		if _, ok := observed[site]; !ok {
			problems = append(problems, "listed site "+site+" no longer references goal status")
		}
	}
	sort.Strings(problems)
	return problems
}

// TestGoalStatusConsumersHandleCancelled is AC-HPR-013's consumer leg
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001, design.md §D6, operator decision Q3).
func TestGoalStatusConsumersHandleCancelled(t *testing.T) {
	t.Run("inventory matches the design table", func(t *testing.T) {
		observed := scanStatusConsumers(t, moduleRoot(t))
		for _, p := range diffStatusConsumers(expectedStatusConsumers, observed) {
			t.Error(p)
		}
	})

	t.Run("an unlisted site is named", func(t *testing.T) {
		// Mutation: drop one row from the table; the diff must name the site.
		dropped := expectedStatusConsumers[len(expectedStatusConsumers)-1]
		observed := map[string][]string{}
		for _, row := range expectedStatusConsumers {
			observed[row.site] = row.idents
		}
		problems := diffStatusConsumers(expectedStatusConsumers[:len(expectedStatusConsumers)-1], observed)
		if len(problems) != 1 || !strings.Contains(problems[0], dropped.site) {
			t.Fatalf("want exactly one problem naming %s, got %v", dropped.site, problems)
		}
	})

	t.Run("cancelled is a recognised status distinct from cleared", func(t *testing.T) {
		if StatusCancelled != "cancelled" {
			t.Fatalf("StatusCancelled = %q, want %q", StatusCancelled, "cancelled")
		}
		if StatusCancelled == StatusCleared {
			t.Fatal("cancelled must not reuse cleared (operator decision Q3)")
		}
		if !IsKnownStatus(StatusCancelled) {
			t.Fatal("IsKnownStatus(cancelled) = false")
		}
		if IsKnownStatus(Status("paused")) || IsKnownStatus(Status("")) {
			t.Fatal("IsKnownStatus accepts a status outside the vocabulary")
		}
	})

	// Every writer path in Evaluate (ceiling, wall-clock, stagnation,
	// unsatisfiable, satisfied) must be unreachable for a cancelled goal.
	writerFixtures := map[string]func() *Goal{
		"unmet condition": func() *Goal {
			return NewGoal("s", "g", []Condition{{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0}})
		},
		"turn ceiling": func() *Goal {
			g := NewGoal("s", "g", []Condition{{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0}})
			g.TurnsUsed = g.Ceiling.MaxTurns
			return g
		},
		"wall-clock bound": func() *Goal {
			g := NewGoal("s", "g", []Condition{{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0}})
			g.Ceiling.MaxDuration = 1
			g.CreatedAt = "2000-01-01T00:00:00Z"
			return g
		},
		"unrunnable condition": func() *Goal {
			return NewGoal("s", "g", []Condition{{Type: ConditionMechanical, Cmd: "no-such-binary", ExpectExit: 0}})
		},
		"satisfied condition": func() *Goal {
			return NewGoal("s", "g", []Condition{{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0}})
		},
	}
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{
		"false":          {exit: 1, out: "fail"},
		"true":           {exit: 0, out: ""},
		"no-such-binary": {exit: 127, out: "command not found"},
	}}

	t.Run("evaluator returns early for cancelled on every writer path", func(t *testing.T) {
		for name, mk := range writerFixtures {
			g := mk()
			g.Status = StatusCancelled
			turns := g.TurnsUsed
			v, block := (&Eval{Runner: runner}).Evaluate(context.Background(), g)
			if block {
				t.Errorf("%s: cancelled goal blocked", name)
			}
			if g.Status != StatusCancelled {
				t.Errorf("%s: status overwritten to %q", name, g.Status)
			}
			if g.TurnsUsed != turns || len(g.Progress) != 0 {
				t.Errorf("%s: cancelled goal was evaluated (turns %d→%d, progress %d)", name, turns, g.TurnsUsed, len(g.Progress))
			}
			if v.Diagnostic != "" || v.Verdict != nil || v.Decision != "" {
				t.Errorf("%s: cancelled goal produced output %+v", name, v)
			}
		}
	})

	t.Run("an unrecognised status yields a diagnostic, no block, never satisfied", func(t *testing.T) {
		for _, bad := range []Status{"paused", ""} {
			for name, mk := range writerFixtures {
				g := mk()
				g.Status = bad
				v, block := (&Eval{Runner: runner}).Evaluate(context.Background(), g)
				if block {
					t.Errorf("status %q / %s: silent block", bad, name)
				}
				if g.Status == StatusSatisfied {
					t.Errorf("status %q / %s: mapped to satisfied", bad, name)
				}
				if g.Status != bad {
					t.Errorf("status %q / %s: status rewritten to %q", bad, name, g.Status)
				}
				if !strings.Contains(v.Diagnostic, "unrecognised goal status") || !strings.Contains(v.Diagnostic, `"`+string(bad)+`"`) {
					t.Errorf("status %q / %s: diagnostic %q does not name the status", bad, name, v.Diagnostic)
				}
			}
		}
	})

	t.Run("dashboard renders cancelled verbatim", func(t *testing.T) {
		g := NewGoal("s", "g", nil)
		g.Status = StatusCancelled
		if m := buildDashboardModel(g, nil); m.Status != "cancelled" {
			t.Fatalf("dashboard status = %q, want cancelled", m.Status)
		}
	})
}
