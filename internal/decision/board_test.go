package decision

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func validRuling(scope string) Record {
	return Record{
		Scope:        scope,
		Kind:         KindRuling,
		DecidedBy:    "claude+leader",
		EvidenceRefs: ".moai/reports/t1/plan-audit-iter3.md",
		LadderPath:   "②",
		Body:         "one delta round; then hold+split",
	}
}

func TestAppend_ValidRecordRoundTripsEveryField(t *testing.T) {
	board := filepath.Join(t.TempDir(), "decisions", "board.jsonl")
	now := time.Date(2026, 10, 3, 9, 15, 0, 0, time.UTC)
	first, err := Append(board, validRuling("card:t1"), AppendOptions{Now: now})
	if err != nil {
		t.Fatalf("append ruling: %v", err)
	}
	waitFile := filepath.Join(t.TempDir(), "progress.md")
	if err := os.WriteFile(waitFile, []byte("wait record: id=w-t1-1 waiting_on=leader reason=x recheck=y\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := validRuling("card:t1")
	second.Supersedes = first.ID
	second.Resolves = "w-t1-1"
	got, err := Append(board, second, AppendOptions{Now: now.Add(time.Second), WaitFile: waitFile})
	if err != nil {
		t.Fatalf("append superseding record: %v", err)
	}
	res, err := Read(board, ReadOptions{Scope: "card:t1", All: true})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if res.Status != StatusOK || len(res.Records) != 2 {
		t.Fatalf("read status=%s records=%d, want ok/2", res.Status, len(res.Records))
	}
	var back Record
	for _, r := range res.Records {
		if r.ID == got.ID {
			back = r
		}
	}
	if back.Scope != "card:t1" || back.Kind != KindRuling || back.DecidedBy != "claude+leader" ||
		back.EvidenceRefs != second.EvidenceRefs || back.LadderPath != "②" || back.Body != second.Body ||
		back.Supersedes != first.ID || back.Resolves != "w-t1-1" || back.Created != "2026-10-03T09:15:01Z" {
		t.Fatalf("round-trip lost a field: %+v", back)
	}
	if !strings.HasPrefix(back.ID, "d-20261003T091501Z-") {
		t.Fatalf("record id %q does not carry the UTC creation stamp", back.ID)
	}
}

func TestAppend_RefusesInvalidRecordsAndAppendsNothing(t *testing.T) {
	board := filepath.Join(t.TempDir(), "board.jsonl")
	if _, err := Append(board, validRuling("standing-free"), AppendOptions{}); err == nil {
		t.Fatalf("scope outside card:<id>|standing accepted")
	}
	cases := map[string]func(*Record){
		"missing decided_by":    func(r *Record) { r.DecidedBy = "" },
		"missing evidence_refs": func(r *Record) { r.EvidenceRefs = "" },
		"missing ladder_path":   func(r *Record) { r.LadderPath = "" },
		"missing body":          func(r *Record) { r.Body = "" },
		"kind outside enum":     func(r *Record) { r.Kind = "verdict" },
		"standing without predicate": func(r *Record) {
			r.Scope = ScopeStanding
			r.Kind = KindStandingRule
		},
		"release-scope without cards": func(r *Record) {
			r.Scope = ScopeStanding
			r.Kind = KindReleaseScope
			r.Release = "v3.2.0"
		},
		"release-scope without release": func(r *Record) {
			r.Scope = ScopeStanding
			r.Kind = KindReleaseScope
			r.Cards = []string{"t1"}
		},
		"release-scope on a card scope": func(r *Record) {
			r.Kind = KindReleaseScope
			r.Release = "v3.2.0"
			r.Cards = []string{"t1"}
		},
		"unknown supersedes":           func(r *Record) { r.Supersedes = "d-missing" },
		"resolves without a wait file": func(r *Record) { r.Resolves = "w-t1-1" },
		"empty card id":                func(r *Record) { r.Scope = "card:" },
	}
	for name, mutate := range cases {
		r := validRuling("card:t1")
		mutate(&r)
		if _, err := Append(board, r, AppendOptions{}); err == nil {
			t.Errorf("%s: accepted, want refusal", name)
		}
	}
	if _, err := os.Stat(board); !os.IsNotExist(err) {
		data, _ := os.ReadFile(board)
		t.Fatalf("refused records left the board non-empty: %q", data)
	}
}

func TestAppend_RefusesResolvesNotPresentInWaitFile(t *testing.T) {
	board := filepath.Join(t.TempDir(), "board.jsonl")
	waitFile := filepath.Join(t.TempDir(), "progress.md")
	if err := os.WriteFile(waitFile, []byte("wait record: id=w-other waiting_on=leader\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := validRuling("card:t1")
	r.Resolves = "w-t1-1"
	if _, err := Append(board, r, AppendOptions{WaitFile: waitFile}); err == nil {
		t.Fatalf("resolves naming a wait id absent from the wait file was accepted")
	}
}

func TestAppend_StandingAndReleaseScopeRecordsAccepted(t *testing.T) {
	board := filepath.Join(t.TempDir(), "board.jsonl")
	standing := validRuling(ScopeStanding)
	standing.Kind = KindStandingRule
	standing.Predicate = "plan verdict admitted, hash unchanged"
	if _, err := Append(board, standing, AppendOptions{}); err != nil {
		t.Fatalf("standing rule: %v", err)
	}
	scope := validRuling(ScopeStanding)
	scope.Kind = KindReleaseScope
	scope.Release = "v3.2.0"
	scope.Cards = []string{"t1480", "t1481"}
	if _, err := Append(board, scope, AppendOptions{}); err != nil {
		t.Fatalf("release-scope: %v", err)
	}
}

func TestRead_ReturnsCardAndStandingRecordsAndHidesSuperseded(t *testing.T) {
	board := filepath.Join(t.TempDir(), "board.jsonl")
	a, err := Append(board, validRuling("card:x"), AppendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(board, validRuling("card:y"), AppendOptions{}); err != nil {
		t.Fatal(err)
	}
	standing := validRuling(ScopeStanding)
	standing.Kind = KindStandingRule
	standing.Predicate = "always"
	if _, err := Append(board, standing, AppendOptions{}); err != nil {
		t.Fatal(err)
	}
	newer := validRuling("card:x")
	newer.Supersedes = a.ID
	b, err := Append(board, newer, AppendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := Read(board, ReadOptions{Scope: "card:x"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusOK {
		t.Fatalf("status %s, want ok", res.Status)
	}
	ids := map[string]bool{}
	for _, r := range res.Records {
		ids[r.ID] = true
		if r.Scope == "card:y" {
			t.Fatalf("card:y record leaked into card:x read")
		}
	}
	if ids[a.ID] || !ids[b.ID] || len(res.Records) != 2 {
		t.Fatalf("default read = %v, want the superseding record and the standing record only", ids)
	}
	all, err := Read(board, ReadOptions{Scope: "card:x", All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Records) != 3 {
		t.Fatalf("--all read returned %d records, want 3 (superseded included)", len(all.Records))
	}
}

func TestRead_ReportsAbsentEmptyAndUnparseableStatuses(t *testing.T) {
	dir := t.TempDir()
	absent, err := Read(filepath.Join(dir, "none.jsonl"), ReadOptions{})
	if err != nil || absent.Status != StatusAbsent {
		t.Fatalf("absent board: status=%s err=%v, want absent", absent.Status, err)
	}
	empty := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if res, err := Read(empty, ReadOptions{}); err != nil || res.Status != StatusEmpty {
		t.Fatalf("empty board: status=%s err=%v, want empty", res.Status, err)
	}
	broken := filepath.Join(dir, "broken.jsonl")
	if err := os.WriteFile(broken, []byte("{not json\n\nalso not\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := Read(broken, ReadOptions{})
	if err != nil {
		t.Fatalf("broken board: %v", err)
	}
	if res.Status != StatusOK || res.Unparseable != 2 || len(res.Records) != 0 {
		t.Fatalf("broken board: status=%s unparseable=%d records=%d, want ok/2/0", res.Status, res.Unparseable, len(res.Records))
	}
	if !strings.Contains(res.StatusLine(), "unparseable=2") || !strings.Contains(res.StatusLine(), "board=ok") {
		t.Fatalf("status line %q", res.StatusLine())
	}
}

func TestRecord_LineRendersTheDecisionRecordForm(t *testing.T) {
	r := validRuling("card:t1")
	line := r.Line()
	want := "decision record: decided_by=claude+leader evidence_refs=.moai/reports/t1/plan-audit-iter3.md ladder_path=②"
	if !strings.HasPrefix(line, want) {
		t.Fatalf("Line() = %q, want prefix %q", line, want)
	}
}

func TestBoardPath_LinkedWorktreeSharesThePrimaryBoard(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	primary := t.TempDir()
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		env := []string{"GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_NOSYSTEM=1"}
		for _, kv := range os.Environ() {
			if !strings.HasPrefix(kv, "GIT_") {
				env = append(env, kv)
			}
		}
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git(primary, "init", "-q", "-b", "main")
	if err := os.MkdirAll(filepath.Join(primary, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(primary, "f"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(primary, "add", "f")
	git(primary, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "init")
	linked := filepath.Join(t.TempDir(), "wt")
	git(primary, "worktree", "add", "-q", linked)
	a, err := BoardPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BoardPath(linked)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("primary board %q != linked-worktree board %q", a, b)
	}
	rec, err := Append(a, validRuling("card:t1"), AppendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := Read(b, ReadOptions{Scope: "card:t1"})
	if err != nil || len(res.Records) != 1 || res.Records[0].ID != rec.ID {
		t.Fatalf("record written from the primary not read from the linked worktree: %+v err=%v", res, err)
	}
}

func TestBoardPath_LivesUnderTheProjectHomeStateNotTheTree(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	p, err := BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(p, home) {
		t.Fatalf("board path %q is not under MOAI_HOME %q", p, home)
	}
	if strings.HasPrefix(p, root) {
		t.Fatalf("board path %q lies inside the working tree %q", p, root)
	}
	if filepath.Base(p) != "board.jsonl" || filepath.Base(filepath.Dir(p)) != "decisions" {
		t.Fatalf("board path %q, want .../decisions/board.jsonl", p)
	}
}
