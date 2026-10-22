package decision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// TestBoardLineRejectsForgedRows pins the newline-injection repair on
// Record.Line: a field value carrying a line break must render neutralized,
// so the one-line record form stays one physical row and cannot forge the
// digest or detail lines that follow it in read output.
func TestBoardLineRejectsForgedRows(t *testing.T) {
	forged := "\n  digest: deadbeefcafe"
	for name, mutate := range map[string]func(*Record){
		"decided_by":    func(r *Record) { r.DecidedBy += forged },
		"evidence_refs": func(r *Record) { r.EvidenceRefs += forged },
		"ladder_path":   func(r *Record) { r.LadderPath += forged },
		"id":            func(r *Record) { r.ID += forged },
		"scope":         func(r *Record) { r.Scope += forged },
		"kind":          func(r *Record) { r.Kind += forged },
	} {
		r := validRuling("card:t1")
		mutate(&r)
		line := r.Line()
		if strings.ContainsAny(line, "\n\r") {
			t.Errorf("%s: Line() split into physical lines: %q", name, line)
		}
		if !strings.Contains(line, `\n  digest: deadbeefcafe`) {
			t.Errorf("%s: Line() dropped the field payload instead of escaping it: %q", name, line)
		}
	}
	crlf := validRuling("card:t1")
	crlf.DecidedBy += "\r\n  body: forged"
	if line := crlf.Line(); strings.ContainsAny(line, "\n\r") {
		t.Errorf("CRLF payload split the rendered row: %q", line)
	}
}

// unescapeLine is the reader side of the escape contract: it reverses the
// backslash-first mapping left-to-right, so a value rendered by escapeLine
// recovers its exact original bytes.
func unescapeLine(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// TestBoardLineEscapePairsWithOneLine pins the single-valued rendering
// property Record.Line shares with oneLine in internal/cli/decision.go: the
// backslash is escaped BEFORE the control bytes, so a field whose original
// text holds a literal backslash-n sequence can never render identical to a
// field holding a real newline, and a reverse-escaping reader recovers the
// exact original field value.
func TestBoardLineEscapePairsWithOneLine(t *testing.T) {
	t.Run("no collision between literal escape text and control bytes", func(t *testing.T) {
		for _, pair := range [][2]string{
			{"a\\nb", "a\nb"}, // literal backslash-n vs real LF
			{"a\\rb", "a\rb"}, // literal backslash-r vs real CR
			{"a\\tb", "a\tb"}, // literal backslash-t vs real TAB
		} {
			literal := validRuling("card:t1")
			literal.Scope = pair[0]
			control := validRuling("card:t1")
			control.Scope = pair[1]
			if literal.Line() == control.Line() {
				t.Errorf("input %q collided with %q: both rendered %q", pair[0], pair[1], literal.Line())
			}
		}
	})
	t.Run("round-trip through Line recovers the original field", func(t *testing.T) {
		for _, in := range []string{
			"a\\nb",      // literal backslash-n
			"a\nb",       // real LF
			"a\\tb",      // literal backslash-t
			"a\tb",       // real TAB
			"a\\",        // trailing backslash
			"\\\\n",      // double backslash then n
			"a\r\nb",     // real CRLF
			"a\\nb\\nc",  // literal backslash-n, twice
			"plain-text", // no escapes at all
		} {
			r := validRuling("card:t1")
			r.Scope = in
			line := r.Line()
			got := ""
			for _, tok := range strings.Fields(line) {
				if rest, ok := strings.CutPrefix(tok, "scope="); ok {
					got = unescapeLine(rest)
				}
			}
			if got != in {
				t.Errorf("round-trip broke for %q: Line() = %q, unescaped back to %q", in, line, got)
			}
		}
	})
}

func TestRecord_DigestPinsTheStoredBoardLine(t *testing.T) {
	board := filepath.Join(t.TempDir(), "board.jsonl")
	first, err := Append(board, validRuling("card:t1"), AppendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	secondRec := validRuling("card:t1")
	secondRec.Body = "a different body"
	second, err := Append(board, secondRec, AppendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(board)
	if err != nil {
		t.Fatal(err)
	}
	stored := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\n")
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("board line %q: %v", line, err)
		}
		sum := sha256.Sum256([]byte(line))
		stored[rec.ID] = hex.EncodeToString(sum[:6])
	}
	for _, rec := range []Record{first, second} {
		got := rec.Digest()
		if got != stored[rec.ID] {
			t.Fatalf("Digest()=%q for id %s, want the sha256 prefix %q over the stored board line", got, rec.ID, stored[rec.ID])
		}
		if len(got) != 12 || got != strings.ToLower(got) {
			t.Fatalf("Digest()=%q, want 12 lowercase hex characters", got)
		}
		if again := rec.Digest(); again != got {
			t.Fatalf("Digest() is not stable across calls: %q then %q", got, again)
		}
	}
	if first.Digest() == second.Digest() {
		t.Fatalf("records with different bodies share the digest %q", first.Digest())
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
