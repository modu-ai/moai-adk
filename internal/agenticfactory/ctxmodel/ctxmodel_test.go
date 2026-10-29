package ctxmodel

import (
	"errors"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	sha1 = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	sha2 = "b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2"
	hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func item() Item {
	return Item{Section: SectionSpec, Title: "SPEC", Summary: "spec body", Path: ".moai/specs/SPEC-X/spec.md",
		Commit: sha1, TargetHead: sha1, SHA256: hash, Owner: "manager-spec", Revision: "r1"}
}

func TestLinkStatesAdvanceOneStepAndOnlyRemoteVerifiedIsAvailable(t *testing.T) {
	l := &Ledger{Issue: 1}
	if err := l.Add(item()); err != nil {
		t.Fatal(err)
	}
	if l.Items[0].State != StatePrepared || l.Items[0].Available() {
		t.Fatalf("new item: %+v", l.Items[0])
	}
	if err := l.Advance(0, StatePushed); !errors.Is(err, ErrStateJump) {
		t.Errorf("skipping steps: %v", err)
	}
	for _, s := range []LinkState{StateMasked, StateCommitted, StatePushed} {
		if err := l.Advance(0, s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
		if l.Items[0].Available() {
			t.Errorf("%s must not be available", s)
		}
	}
	if err := l.Advance(0, StateRemoteVerified); !errors.Is(err, ErrNeedsLookup) {
		t.Errorf("remote-verified by assertion: %v", err)
	}
}

type fakeVerifier struct {
	sha string
	ok  bool
	err error
	n   int
}

func (f *fakeVerifier) Lookup(commit, path string) (string, bool, error) {
	f.n++
	return f.sha, f.ok, f.err
}

func TestVerifyRemoteNeedsALookupThatConfirmsHash(t *testing.T) {
	l := &Ledger{Issue: 1}
	_ = l.Add(item())
	for _, s := range []LinkState{StateMasked, StateCommitted, StatePushed} {
		_ = l.Advance(0, s)
	}
	if err := l.VerifyRemote(0, &fakeVerifier{sha: sha2, ok: true}); !errors.Is(err, ErrHashMismatch) || l.Items[0].Available() {
		t.Errorf("mismatch: %v %+v", err, l.Items[0])
	}
	if err := l.VerifyRemote(0, &fakeVerifier{ok: false}); !errors.Is(err, ErrNotFoundRemote) {
		t.Errorf("missing: %v", err)
	}
	boom := errors.New("net")
	if err := l.VerifyRemote(0, &fakeVerifier{err: boom}); !errors.Is(err, boom) || l.Items[0].Available() {
		t.Errorf("lookup error: %v", err)
	}
	if err := l.VerifyRemote(0, &fakeVerifier{sha: hash, ok: true}); err != nil || !l.Items[0].Available() {
		t.Fatalf("confirmed: %v %+v", err, l.Items[0])
	}
}

func TestVerifyRemoteRefusesAnItemThatIsNotPushed(t *testing.T) {
	l := &Ledger{Issue: 1}
	_ = l.Add(item())
	v := &fakeVerifier{sha: hash, ok: true}
	if err := l.VerifyRemote(0, v); !errors.Is(err, ErrStateJump) || v.n != 0 {
		t.Errorf("unpushed: %v lookups=%d", err, v.n)
	}
}

func TestAddRejectsUnpinnedOrUntrackedLinks(t *testing.T) {
	bad := map[string]func(*Item){
		"branch ref":    func(i *Item) { i.Commit = "main" },
		"short sha":     func(i *Item) { i.Commit = "a1a1a1a" },
		"reports dir":   func(i *Item) { i.Path = ".moai/reports/t1/verdict.md" },
		"absolute path": func(i *Item) { i.Path = "/Users/x/.moai/specs/a.md" },
		"dotdot":        func(i *Item) { i.Path = ".moai/specs/../../etc/passwd" },
		"bad hash":      func(i *Item) { i.SHA256 = "zz" },
		"no owner":      func(i *Item) { i.Owner = "" },
		"bad section":   func(i *Item) { i.Section = "misc" },
	}
	for name, mut := range bad {
		l := &Ledger{Issue: 1}
		it := item()
		mut(&it)
		if err := l.Add(it); !errors.Is(err, ErrLinkBlocked) {
			t.Errorf("%s: %v", name, err)
		}
		if len(l.Items) != 0 {
			t.Errorf("%s was stored", name)
		}
	}
}

func TestPendingItemCarriesNoLink(t *testing.T) {
	l := &Ledger{Issue: 1}
	if err := l.Add(Item{Section: SectionAudit, Title: "audit", Pending: true, Owner: "plan-auditor"}); err != nil {
		t.Fatal(err)
	}
	out, err := l.Render()
	if err != nil || !strings.Contains(out, "PENDING") || strings.Contains(out, "](") {
		t.Errorf("pending render: %q %v", out, err)
	}
	if l.Items[0].Available() {
		t.Error("a pending item is never available")
	}
}

func TestRefreshMarksStaleAndKeepsTheOldItem(t *testing.T) {
	l := &Ledger{Issue: 1}
	_ = l.Add(item())
	n := l.Refresh(map[string]Current{".moai/specs/SPEC-X/spec.md": {Head: sha2, SHA256: hash}})
	if n != 1 || !l.Items[0].Stale || len(l.Items) != 1 {
		t.Errorf("head moved: n=%d %+v", n, l.Items)
	}
	l2 := &Ledger{Issue: 1}
	_ = l2.Add(item())
	if n := l2.Refresh(map[string]Current{".moai/specs/SPEC-X/spec.md": {Head: sha1, SHA256: hash}}); n != 0 || l2.Items[0].Stale {
		t.Errorf("unchanged: n=%d", n)
	}
	l3 := &Ledger{Issue: 1}
	_ = l3.Add(item())
	if n := l3.Refresh(map[string]Current{".moai/specs/SPEC-X/spec.md": {Head: sha1, SHA256: strings.Repeat("f", 64)}}); n != 1 || !l3.Items[0].Stale {
		t.Errorf("hash moved: n=%d", n)
	}
}

func TestRenderRefusesWhenMaskingWouldChangeIt(t *testing.T) {
	for name, s := range map[string]string{
		"token":  "ghp_abcdefghijklmnopqrstuvwxyz0123456789",
		"email":  "someone@example.com",
		"home":   "/Users/goos/secret/file",
		"rawlog": strings.Repeat("line\n", 40),
	} {
		l := &Ledger{Issue: 1}
		it := item()
		it.Summary = s
		_ = l.Add(it)
		if out, err := l.Render(); !errors.Is(err, ErrUnmasked) || out != "" {
			t.Errorf("%s: out=%q err=%v", name, out, err)
		}
	}
}

func TestRestrictedKeepsOnlySummaryAndHash(t *testing.T) {
	l := &Ledger{Issue: 1}
	it := item()
	it.Restricted = true
	_ = l.Add(it)
	for _, s := range []LinkState{StateMasked, StateCommitted, StatePushed} {
		_ = l.Advance(0, s)
	}
	if err := l.VerifyRemote(0, &fakeVerifier{sha: hash, ok: true}); err != nil {
		t.Fatal(err)
	}
	if l.Items[0].Available() {
		t.Error("a restricted item is never an available link")
	}
	out, err := l.Render()
	if err != nil || strings.Contains(out, "](") || strings.Contains(out, it.Path) || !strings.Contains(out, hash[:12]) || !strings.Contains(out, "RESTRICTED") {
		t.Errorf("restricted render: %q %v", out, err)
	}
}

func TestRenderSectionsAndAvailableLinkShape(t *testing.T) {
	l := &Ledger{Issue: 7}
	_ = l.Add(item())
	for _, s := range []LinkState{StateMasked, StateCommitted, StatePushed} {
		_ = l.Advance(0, s)
	}
	_ = l.VerifyRemote(0, &fakeVerifier{sha: hash, ok: true})
	out, err := l.Render()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"SPEC", "AVAILABLE", ".moai/specs/SPEC-X/spec.md", sha1[:12]} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
	if strings.Contains(out, "main") {
		t.Errorf("render names a branch: %q", out)
	}
	if again, _ := l.Render(); again != out {
		t.Error("render is not deterministic")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	l := &Ledger{Issue: 3}
	_ = l.Add(item())
	b, err := l.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(b)
	if err != nil || got.Issue != 3 || len(got.Items) != 1 || got.Items[0].SHA256 != hash {
		t.Errorf("round trip: %+v %v", got, err)
	}
	if _, err := Unmarshal([]byte("{bad")); err == nil {
		t.Error("corrupt data accepted")
	}
}

// GitVerifier is judged on a real local repository: a path that exists at the
// commit with a matching hash, a path absent at the commit, and a bad commit.
func TestGitVerifierOnLocalFixture(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) string {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.invalid", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	p := filepath.Join(dir, ".moai", "specs", "SPEC-X")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "spec.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".moai/specs/SPEC-X/spec.md")
	run("commit", "-q", "-m", "x")
	head := run("rev-parse", "HEAD")
	v := GitVerifier{Dir: dir}
	got, ok, err := v.Lookup(head, ".moai/specs/SPEC-X/spec.md")
	want := "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"
	if err != nil || !ok || got != want {
		t.Errorf("present: %q %v %v", got, ok, err)
	}
	if _, ok, err := v.Lookup(head, ".moai/specs/SPEC-X/missing.md"); err != nil || ok {
		t.Errorf("absent path: ok=%v err=%v", ok, err)
	}
	if _, ok, err := v.Lookup(strings.Repeat("0", 40), ".moai/specs/SPEC-X/spec.md"); err != nil || ok {
		t.Errorf("unknown commit must read as not found, got ok=%v err=%v", ok, err)
	}
}

func TestUpdatePersistsThroughTheContextDocument(t *testing.T) {
	s, err := claim.OpenWithLimit(t.TempDir(), 3)
	if err != nil {
		t.Fatal(err)
	}
	l, err := Update(s, 5, func(l *Ledger) error { return l.Add(item()) })
	if err != nil || len(l.Items) != 1 || l.Issue != 5 {
		t.Fatalf("first update: %+v %v", l, err)
	}
	l, err = Update(s, 5, func(l *Ledger) error { return l.Advance(0, StateMasked) })
	if err != nil || l.Items[0].State != StateMasked {
		t.Fatalf("second update must see the stored item: %+v %v", l, err)
	}
	boom := errors.New("refused")
	if _, err := Update(s, 5, func(l *Ledger) error { l.Items = nil; return boom }); !errors.Is(err, boom) {
		t.Errorf("fn error: %v", err)
	}
	got, err := Load(s, 5)
	if err != nil || len(got.Items) != 1 || got.Items[0].State != StateMasked {
		t.Errorf("an aborted update changed the stored ledger: %+v %v", got, err)
	}
	if empty, err := Load(s, 99); err != nil || empty.Issue != 99 || len(empty.Items) != 0 {
		t.Errorf("missing ledger reads as empty: %+v %v", empty, err)
	}
	d, _ := s.GetDoc("context", "issue-5")
	if d.Rev != 2 {
		t.Errorf("each successful update is one revision, got %d", d.Rev)
	}
}
