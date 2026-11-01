package cli

// The heartbeat repair (follow-up of the final code audit of a36ea8c368): `moai af claim` records the ledger file's path in
// the session's heartbeat binding (<project>/.moai/state/ledger-heartbeat/<session>.json), and the hooks of that session
// renew the lease through it. `moai af move-repository --from` moves the default-location ledger file to the key of the
// moved project directory; a binding that still named the old path made every renewal fail silently (the file is not
// there any more), and the lease ran out under a live session. After the file move and the recorded move the verb
// rewrites every binding of the project that names the old path so that it names the new one. The tests go through the
// real verbs and the real heartbeat path of the hooks (hook.LedgerHeartbeatSeam), and begin with TestAFHeartbeatFrom_ so
// that one anchored run picks them; with AF_TEST_BIN set the verbs are that binary's.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/hook"
)

func afHBBindingDir(project string) string {
	return filepath.Join(project, ".moai", "state", "ledger-heartbeat")
}

func afHBReadBinding(t *testing.T, project, session string) hook.LedgerBinding {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(afHBBindingDir(project), session+".json"))
	if err != nil {
		t.Fatalf("the binding of %s: %v", session, err)
	}
	var b hook.LedgerBinding
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("the binding of %s is not JSON: %v (%s)", session, err, raw)
	}
	return b
}

// afHBLease is the lease of card t1 in the ledger at path.
func afHBLease(t *testing.T, path string) ledger.Lease {
	t.Helper()
	st, err := ledger.Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	leases, err := st.Leases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range leases {
		if l.CardID == "t1" {
			return l
		}
	}
	t.Fatalf("no lease for t1 in %s: %+v", path, leases)
	return ledger.Lease{}
}

// afHBRenews runs the heartbeat path the hooks run for the session and reports whether the lease in the ledger at path
// was renewed by it.
func afHBRenews(t *testing.T, project, session, path string) bool {
	t.Helper()
	before := afHBLease(t, path).HeartbeatAt
	time.Sleep(60 * time.Millisecond) // the lease clock has millisecond resolution
	hook.LedgerHeartbeatSeam(&hook.HookInput{SessionID: session, CWD: project})
	return afHBLease(t, path).HeartbeatAt.After(before)
}

// afHBMove runs the move with --from and returns the printed JSON as a map.
func afHBMove(t *testing.T, old string) map[string]any {
	t.Helper()
	out, errOut, err := afRunBoth("move-repository", "--reason", "moved", "--from", old)
	if err != nil {
		t.Fatalf("move-repository --from: %v %.300s %.300s", err, out, errOut)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("stdout %.300q is not the verb's JSON line: %v", out, err)
	}
	return res
}

// The failure the repair is for: a claimed card, the project directory renamed, the ledger moved with --from; the hooks of
// the session must renew the lease at the new place.
func TestAFHeartbeatFrom_AfterTheMoveTheHooksRenewTheLeaseAtTheNewPlace(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	if got := afHBReadBinding(t, d.moved, "s1"); got.DB != d.oldLedger || got.CardID != "t1" {
		t.Fatalf("setup: the claim bound %+v, want the old ledger path %s", got, d.oldLedger)
	}
	res := afHBMove(t, d.old)
	if n, _ := res["heartbeat_bindings_rewritten"].(float64); n != 1 {
		t.Errorf("the verb printed %+v, want heartbeat_bindings_rewritten 1", res)
	}
	if w, ok := res["warning_heartbeat_bindings"]; ok {
		t.Errorf("the verb warned although every binding was rewritten: %v", w)
	}
	if got := afHBReadBinding(t, d.moved, "s1"); got.DB != d.newLedger || got.CardID != "t1" || got.Epoch != 1 {
		t.Errorf("the binding is %+v after the move, want the new ledger path %s with the card and epoch it had", got, d.newLedger)
	}
	if runtime.GOOS != "windows" {
		if fi, err := os.Stat(filepath.Join(afHBBindingDir(d.moved), "s1.json")); err != nil || fi.Mode().Perm() != 0o600 {
			t.Errorf("the rewritten binding: %v %v, want mode 0600", fi, err)
		}
	}
	if !afHBRenews(t, d.moved, "s1", d.newLedger) {
		t.Error("the hooks of the session did not renew the lease at the new place")
	}
	if afIDExists(d.oldLedger) {
		t.Errorf("a renewal left a ledger at the old place %s", d.oldLedger)
	}
}

// Only a binding that names the moved ledger is rewritten: every session of the project that held that ledger (two here),
// and nothing else (another ledger, a path that only starts the same way, the stamp files, a file that is not a binding).
func TestAFHeartbeatFrom_OnlyBindingsThatNameTheMovedLedgerAreRewritten(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	dir := afHBBindingDir(d.moved)
	if err := hook.BindLedgerHeartbeat(d.moved, "s3", hook.LedgerBinding{DB: d.oldLedger, CardID: "t1", Epoch: 1}); err != nil {
		t.Fatal(err)
	}
	other := map[string]string{
		"s2.json":    afHBJSON(t, hook.LedgerBinding{DB: filepath.Join(t.TempDir(), "elsewhere", "ledger.db"), CardID: "c9", Epoch: 4}),
		"s4.json":    afHBJSON(t, hook.LedgerBinding{DB: d.oldLedger + "-other", CardID: "t1", Epoch: 1}),
		"s1.stamp":   "1700000000000000000",
		"junk.json":  "not json at all",
		"notes.txt":  d.oldLedger,
		"empty.json": afHBJSON(t, hook.LedgerBinding{}),
	}
	for name, body := range other {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	res := afHBMove(t, d.old)
	if n, _ := res["heartbeat_bindings_rewritten"].(float64); n != 2 {
		t.Errorf("the verb printed %+v, want heartbeat_bindings_rewritten 2 (s1 and s3)", res)
	}
	if _, ok := res["warning_heartbeat_bindings"]; ok {
		t.Errorf("the verb warned for files that are not bindings of the moved ledger: %v", res["warning_heartbeat_bindings"])
	}
	for _, s := range []string{"s1", "s3"} {
		if got := afHBReadBinding(t, d.moved, s); got.DB != d.newLedger {
			t.Errorf("the binding of %s names %s, want %s", s, got.DB, d.newLedger)
		}
	}
	for name, want := range other {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != want {
			t.Errorf("%s was touched: %q %v, want %q", name, got, err, want)
		}
	}
}

func afHBJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The normal path beside it: a project that never moved keeps renewing through the binding the claim wrote.
func TestAFHeartbeatFrom_AProjectThatNeverMovedKeepsRenewing(t *testing.T) {
	afSkipWithoutSh(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	ledgerPath, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "t1", "--path", "a.txt", "--class", "A"},
		{"claim", "--session", "s1", "--worktree", repo},
	} {
		if out, errOut, err := afRunBoth(args...); err != nil {
			t.Fatalf("moai af %s: %v %s %s", args[0], err, out, errOut)
		}
	}
	if got := afHBReadBinding(t, repo, "s1"); got.DB != ledgerPath {
		t.Fatalf("the claim bound %+v, want %s", got, ledgerPath)
	}
	if !afHBRenews(t, repo, "s1", ledgerPath) {
		t.Error("the hooks of a session in a project that never moved did not renew the lease")
	}
}

// A failure to rewrite does not undo the move that was recorded: it is a named warning in the verb's output.
func TestAFHeartbeatFrom_AFailureToRewriteIsANamedWarningAndTheMoveStands(t *testing.T) {
	afSkipWithoutSh(t)
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a directory the caller cannot write to")
	}
	d := afIDDefaultMoved(t)
	dir := afHBBindingDir(d.moved)
	before, err := os.ReadFile(filepath.Join(dir, "s1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	res := afHBMove(t, d.old) // the verb succeeds: the move is recorded
	w, _ := res["warning_heartbeat_bindings"].(string)
	if w == "" || !strings.Contains(w, "ledger-heartbeat") {
		t.Errorf("the verb printed %+v, want warning_heartbeat_bindings naming the binding directory", res)
	}
	if n, _ := res["heartbeat_bindings_rewritten"].(float64); n != 0 {
		t.Errorf("heartbeat_bindings_rewritten %v, want 0 when nothing could be rewritten", res["heartbeat_bindings_rewritten"])
	}
	if got, err := os.ReadFile(filepath.Join(dir, "s1.json")); err != nil || string(got) != string(before) {
		t.Errorf("a binding that could not be rewritten was changed: %q %v", got, err)
	}
	// the move stands: the ledger is at the new key, bound to the new repository, the move is listed
	st, err := ledger.Open(context.Background(), d.newLedger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if moves, err := st.RepositoryMoves(context.Background()); err != nil || len(moves) != 1 {
		t.Errorf("moves %+v %v, want the one just recorded", moves, err)
	}
	if afIDExists(d.oldLedger) {
		t.Errorf("the ledger is still at the old place %s", d.oldLedger)
	}
}

// A project that has no binding directory (no session claimed through the hooks' state) has nothing to rewrite, and that
// is not a failure.
func TestAFHeartbeatFrom_AProjectWithoutBindingsHasNothingToRewriteAndNothingToWarnAbout(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	if err := os.RemoveAll(afHBBindingDir(d.moved)); err != nil {
		t.Fatal(err)
	}
	res := afHBMove(t, d.old)
	if n, _ := res["heartbeat_bindings_rewritten"].(float64); n != 0 {
		t.Errorf("heartbeat_bindings_rewritten %v, want 0", res["heartbeat_bindings_rewritten"])
	}
	if _, ok := res["heartbeat_bindings_rewritten"]; !ok {
		t.Errorf("the verb printed %+v, want the field (0) so that a reader sees the step ran", res)
	}
	if w, ok := res["warning_heartbeat_bindings"]; ok {
		t.Errorf("a project without bindings is a warning: %v", w)
	}
}
