package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestTodoAuditPRIncompleteIsUnknown(t *testing.T) {
	for _, saturated := range []bool{false, true} {
		t.Run(fmt.Sprint(saturated), func(t *testing.T) {
			_, store := todoFixture(t)
			if _, _, err := store.Add("audit card"); err != nil {
				t.Fatal(err)
			}
			spy := &spyRunner{prJSON: "[]"}
			if saturated {
				prs := make([]kanban.PRRecord, todoPROpenPRLimit)
				for i := range prs {
					prs[i] = kanban.PRRecord{Number: i + 1, Title: fmt.Sprintf("unrelated %d", i), State: "OPEN"}
				}
				raw, _ := json.Marshal(prs)
				spy.prJSON = string(raw)
			} else {
				spy.ghFail = fmt.Errorf("offline")
			}
			installSpy(t, spy)
			out, _, err := runTodo(t, "pr", "t1", "--json")
			if err != nil || !strings.Contains(out, `"outcome":"unknown"`) || !strings.Contains(out, `"pr_lookup"`) {
				t.Fatalf("incomplete result=%s err=%v", out, err)
			}
		})
	}
}

func TestTodoAuditPRMissingSkipsNetwork(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := store.Add("existing card"); err != nil {
		t.Fatal(err)
	}
	spy := installSpy(t, &spyRunner{prJSON: "[]"})
	out, _, err := runTodo(t, "pr", "t999")
	if err == nil || !strings.Contains(err.Error(), "t999") || len(spy.calls) != 0 {
		t.Fatalf("stdout=%q err=%v calls=%v", out, err, spy.calls)
	}
}

func TestTodoAuditMultilineRows(t *testing.T) {
	for _, verb := range []string{"pr", "list", "next", "history"} {
		t.Run(verb, func(t *testing.T) {
			_, store := todoFixture(t)
			original := "card text\nforged\trow\rreturn"
			if _, _, err := store.Add(original); err != nil {
				t.Fatal(err)
			}
			installSpy(t, &spyRunner{prJSON: "[]"})
			args := []string{verb}
			if verb == "history" {
				args = append(args, "t1")
			}
			out, _, err := runTodo(t, args...)
			if err != nil || strings.Count(out, "\n") != 1 || strings.Contains(out, "\r") {
				t.Fatalf("stdout=%q err=%v", out, err)
			}
			wantTabs := map[string]int{"pr": 6, "list": 2, "next": 1, "history": 4}[verb]
			if strings.Count(out, "\t") != wantTabs {
				t.Fatalf("row separators=%q", out)
			}
			rec, err := store.Load()
			if err != nil || rec.Items[0].Text != original {
				t.Fatalf("stored text changed: %v %v", rec, err)
			}
		})
	}
}

func TestTodoAuditNegativeLimitEmpty(t *testing.T) {
	todoFixture(t)
	out, _, err := runTodo(t, "list", "--limit", "-1")
	if err == nil {
		t.Fatalf("negative limit accepted: %q", out)
	}
}

func TestTodoAuditPickDropped(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := store.Add("original card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "obsolete"); err != nil {
		t.Fatal(err)
	}
	before, _ := store.Load()
	out, _, err := runTodo(t, "next", "t1")
	rec, loadErr := store.Load()
	if err == nil || loadErr != nil || rec.Items[0].State != kanban.BacklogStateDropped || rec.Items[0].Text != before.Items[0].Text {
		t.Fatalf("stdout=%q err=%v rec=%+v load=%v", out, err, rec, loadErr)
	}
}

func TestTodoAuditLandedHelpLazy(t *testing.T) {
	todoFixture(t)
	cmd := newTodoLandedCmd()
	if cmd.Long != "" || cmd.Flags().Lookup("ref").Usage != "" {
		t.Fatal("landed help materialized before rendering")
	}
	out := helpOutput(t, "landed")
	// The subject is LAZINESS plus a materialized body, not which default ref
	// the resolver picks: that default depends on ambient project state other
	// parallel tests perturb, so assert the resolution-independent invariant.
	if !strings.Contains(out, "Record what YOU assert") || !strings.Contains(out, "the same one") {
		t.Fatalf("lost help: %s", out)
	}
}

func TestTodoAuditPRUsesQueueRoot(t *testing.T) {
	for _, verb := range []string{"pr", "done"} {
		t.Run(verb, func(t *testing.T) {
			root, store := todoFixture(t)
			if _, _, err := store.Add("project A card"); err != nil {
				t.Fatal(err)
			}
			other := t.TempDir()
			initGitRepo(t, other)
			for _, args := range [][]string{
				{"-C", root, "update-ref", "refs/remotes/origin/main", "HEAD"},
				{"-C", other, "-c", "user.name=Audit", "-c", "user.email=audit@example.invalid", "commit", "--allow-empty", "-qm", "fix(t1): unrelated project B"},
				{"-C", other, "update-ref", "refs/remotes/origin/main", "HEAD"},
			} {
				if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
					t.Fatalf("%s %v", out, err)
				}
			}
			old, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(other); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(old) })
			prior := todoRunCommand
			t.Cleanup(func() { todoRunCommand = prior })
			todoRunCommand = func(name string, args ...string) (string, error) {
				if name == "gh" {
					return "[]", nil
				}
				return prior(name, args...)
			}
			if verb == "pr" {
				out, stderr, err := runTodo(t, "pr", "t1", "--json")
				if err != nil || !strings.Contains(out, `"outcome":"no-link"`) {
					t.Fatalf("stdout=%s stderr=%q err=%v", out, stderr, err)
				}
			} else {
				out, _, err := runTodo(t, "done", "t1", "--require-landed")
				if err == nil {
					t.Fatalf("archived using unrelated project: %s", out)
				}
				rec, loadErr := store.Load()
				if loadErr != nil || len(rec.Items) != 1 {
					t.Fatalf("queue changed: %+v %v", rec, loadErr)
				}
			}
		})
	}
}

func TestTodoAuditAnalyzeAllocationBound(t *testing.T) {
	rec := auditAnalysisRecord(100)
	allocations := testing.AllocsPerRun(1, func() { analyzeQueue(rec) })
	if allocations > 5000 {
		t.Fatalf("100-card analysis allocates %.0f objects; want <=5000", allocations)
	}
}

func auditAnalysisRecord(n int) *kanban.BacklogRecord {
	rec := &kanban.BacklogRecord{}
	for i := 0; i < n; i++ {
		rec.Items = append(rec.Items, kanban.BacklogItem{ID: fmt.Sprintf("t%d", i+1), State: kanban.BacklogStateQueued, Text: fmt.Sprintf("unique%d task%d implement%d fixture%d", i, i, i, i)})
	}
	return rec
}

func BenchmarkTodoAuditAnalyze(b *testing.B) {
	for _, n := range []int{100, 500} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			rec := auditAnalysisRecord(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				analyzeQueue(rec)
			}
		})
	}
}

func TestTodoAuditDropReasonSingleLine(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := store.Add("original"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runTodo(t, "drop", "t1", "reason\nsecond\tcolumn")
	if err != nil || strings.Count(out, "\n") != 1 || strings.Contains(out, "\t") {
		t.Fatalf("stdout=%q err=%v", out, err)
	}
}

func TestTodoAuditFindingNoteSingleLine(t *testing.T) {
	rec := auditAnalysisRecord(2)
	f := kanban.BacklogFinding{SubjectID: "t1", RelatedID: "t2", Note: "reason\nsecond\tcolumn"}
	out := todoFindingLine(rec, "t1", f)
	if strings.Contains(out, "\n") || strings.Count(out, "\t") != 1 {
		t.Fatalf("finding=%q", out)
	}
}

func TestTodoAuditPRDoesNotMigrateLegacyJSON(t *testing.T) {
	_, store := todoFixture(t)
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(), []byte(staleBacklogJSON), 0600); err != nil {
		t.Fatal(err)
	}
	installSpy(t, &spyRunner{prJSON: "[]"})
	out, _, err := runTodo(t, "pr", "t1", "--json")
	if err != nil || !strings.Contains(out, `"card_id":"t1"`) {
		t.Fatalf("legacy read=%s %v", out, err)
	}
	if _, err := os.Stat(store.EnginePath()); !os.IsNotExist(err) {
		t.Fatalf("read migrated legacy JSON: %v", err)
	}
	raw, err := os.ReadFile(store.Path())
	if err != nil || string(raw) != staleBacklogJSON {
		t.Fatalf("legacy changed: %q %v", raw, err)
	}
}

func TestTodoAuditSubprocessScopeIgnoresInheritedRepo(t *testing.T) {
	root, _ := todoFixture(t)
	other := t.TempDir()
	initGitRepo(t, other)
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	out, err := todoRunCommand("git", "rev-parse", "--show-toplevel")
	want, _ := filepath.EvalSymlinks(root)
	got, _ := filepath.EvalSymlinks(strings.TrimSpace(out))
	if err != nil || got != want {
		t.Fatalf("git root=%q want=%q err=%v", out, want, err)
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nprintf '%s|%s' \"$PWD\" \"${GH_REPO-}\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_REPO", "unrelated/repository")
	out, err = todoRunCommand("gh", "pr", "list")
	parts := strings.Split(out, "|")
	if err != nil || len(parts) != 2 {
		t.Fatalf("gh result=%q %v", out, err)
	}
	got, _ = filepath.EvalSymlinks(parts[0])
	if got != want || parts[1] != "" {
		t.Fatalf("gh scope=%q want=%q|", out, want)
	}
}

func TestTodoAuditPreparedAnalysisMatchesComparator(t *testing.T) {
	texts := []string{"CAFÉ ship", "café ship", "fix gate now", " fix  gate now ", "gate now fix", "fix gate now please", "", "   ", "repeat repeat", "repeat", "한국어 카드", "다른 카드"}
	for _, a := range texts {
		for _, b := range texts {
			rec := &kanban.BacklogRecord{Items: []kanban.BacklogItem{{ID: "t1", Text: a, State: kanban.BacklogStateQueued}, {ID: "t2", Text: b, State: kanban.BacklogStateQueued}}}
			score := kanban.TokenSetJaccard(a, b)
			want := ""
			if kanban.NormalizeCardText(a) != "" && kanban.NormalizeCardText(a) == kanban.NormalizeCardText(b) {
				want = kanban.BacklogRelationDuplicateForced
				score = 1
			} else if score >= kanban.BacklogNearDuplicateThreshold && score < 1 {
				want = kanban.BacklogRelationNearDuplicate
			}
			pairs, recorded := analyzeQueue(rec)
			if pairs != 1 || (recorded == 0) != (want == "") {
				t.Fatalf("%q/%q pairs=%d findings=%+v want=%s", a, b, pairs, rec.Findings, want)
			}
			if want != "" && (rec.Findings[0].Relation != want || rec.Findings[0].Score != score) {
				t.Fatalf("%q/%q finding=%+v want=%s %v", a, b, rec.Findings[0], want, score)
			}
		}
	}
}
