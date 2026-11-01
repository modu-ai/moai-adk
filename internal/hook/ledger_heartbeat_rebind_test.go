package hook

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// RebindLedgerHeartbeats follows a ledger file that moved (move-repository --from): the bindings that name the old path
// name the new one, nothing else is touched, and the seam renews through the rewritten binding.

func writeBindingFile(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := ledgerHeartbeatDir(root)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRebindLedgerHeartbeats_RewritesOnlyTheBindingsThatNameTheOldPath(t *testing.T) {
	root := t.TempDir()
	for _, s := range []string{"s1", "s2"} {
		if err := BindLedgerHeartbeat(root, s, LedgerBinding{DB: "/old/af/ledger.db", CardID: "c" + s, Epoch: 3}); err != nil {
			t.Fatal(err)
		}
	}
	// the same path spelled with a redundant element is the same path
	if err := BindLedgerHeartbeat(root, "s5", LedgerBinding{DB: "/old/af/../af/ledger.db", CardID: "c5", Epoch: 1}); err != nil {
		t.Fatal(err)
	}
	keep := map[string]string{
		"s3.json":   `{"db":"/elsewhere/ledger.db","card_id":"c3","epoch":4}`,
		"s4.json":   `{"db":"/old/af/ledger.db-other","card_id":"c4","epoch":1}`,
		"s1.stamp":  "1",
		"bad.json":  "{",
		"zero.json": `{"db":"","card_id":"","epoch":0}`,
	}
	for n, body := range keep {
		writeBindingFile(t, root, n, body)
	}
	n, err := RebindLedgerHeartbeats(root, "/old/af/ledger.db", "/new/af/ledger.db")
	if err != nil || n != 3 {
		t.Fatalf("Rebind = %d, %v; want 3 rewritten", n, err)
	}
	for s, card := range map[string]string{"s1": "cs1", "s2": "cs2", "s5": "c5"} {
		raw, err := os.ReadFile(filepath.Join(ledgerHeartbeatDir(root), s+".json"))
		if err != nil || !strings.Contains(string(raw), `"db":"/new/af/ledger.db"`) || !strings.Contains(string(raw), `"card_id":"`+card+`"`) {
			t.Errorf("%s = %q %v, want the new path with the card it had", s, raw, err)
		}
	}
	for name, want := range keep {
		got, err := os.ReadFile(filepath.Join(ledgerHeartbeatDir(root), name))
		if err != nil || string(got) != want {
			t.Errorf("%s was touched: %q %v", name, got, err)
		}
	}
	// no temp file is left beside the bindings
	entries, _ := os.ReadDir(ledgerHeartbeatDir(root))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("a temp file was left: %s", e.Name())
		}
	}
	if runtime.GOOS != "windows" {
		if fi, err := os.Stat(filepath.Join(ledgerHeartbeatDir(root), "s1.json")); err != nil || fi.Mode().Perm() != 0o600 {
			t.Errorf("mode of a rewritten binding: %v %v, want 0600", fi, err)
		}
	}
}

func TestRebindLedgerHeartbeats_NoBindingDirectoryIsNothingToDo(t *testing.T) {
	n, err := RebindLedgerHeartbeats(t.TempDir(), "/old", "/new")
	if n != 0 || err != nil {
		t.Fatalf("Rebind = %d, %v; want 0, nil", n, err)
	}
}

// A binding that cannot be rewritten is an error; the others are still rewritten and counted.
func TestRebindLedgerHeartbeats_AFailureIsReportedAndDoesNotStopTheOthers(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a directory the caller cannot write to")
	}
	root := t.TempDir()
	if err := BindLedgerHeartbeat(root, "s1", LedgerBinding{DB: "/old", CardID: "c1", Epoch: 1}); err != nil {
		t.Fatal(err)
	}
	dir := ledgerHeartbeatDir(root)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	n, err := RebindLedgerHeartbeats(root, "/old", "/new")
	if err == nil || n != 0 {
		t.Fatalf("Rebind = %d, %v; want 0 rewritten and an error", n, err)
	}
	raw, rerr := os.ReadFile(filepath.Join(dir, "s1.json"))
	if rerr != nil || !strings.Contains(string(raw), `"db":"/old"`) {
		t.Errorf("a binding that could not be rewritten was changed: %q %v", raw, rerr)
	}
}

// The seam renews through the rewritten binding: the beat goes to the new path.
func TestRebindLedgerHeartbeats_TheSeamBeatsAtTheNewPath(t *testing.T) {
	root, calls, _, _ := ledgerBeatFixture(t)
	if err := BindLedgerHeartbeat(root, "s1", LedgerBinding{DB: "/old/ledger.db", CardID: "c1", Epoch: 7}); err != nil {
		t.Fatal(err)
	}
	if n, err := RebindLedgerHeartbeats(root, "/old/ledger.db", "/new/ledger.db"); err != nil || n != 1 {
		t.Fatalf("Rebind = %d, %v", n, err)
	}
	LedgerHeartbeatSeam(&HookInput{SessionID: "s1", CWD: root})
	if len(*calls) != 1 || (*calls)[0] != (beatCall{"/new/ledger.db", "c1", 7}) {
		t.Fatalf("beats = %+v, want one at the new path with the card and epoch of the binding", *calls)
	}
}
