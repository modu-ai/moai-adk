package cli

// Regression tests for the repair after the audit of db0a69698a (card issue-1948), through the real `moai af` verbs,
// a real ledger and real git. Values a write hook refuses are assembled at run time (afD1Words, afD1Value).

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// afRepairLand runs `af land` with the gate on the standard queued card and returns the verb's JSON line.
func afRepairLand(t *testing.T, db, repo, gate string) (outcome, reason, stdout string) {
	t.Helper()
	out, _, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	if err != nil {
		t.Fatalf("the verb failed: %v (stdout %.200q)", err, out)
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil {
		t.Fatalf("stdout %.200q is not the verb's JSON line", out)
	}
	return res.Outcome, res.Reason, out
}

// afGateIDPattern is the short id the reason and the landed note carry in place of the gate text: "gate-" and 12 hex digits.
var afGateIDPattern = regexp.MustCompile(`gate-[0-9a-f]{12}`)

// N29: the gate TEXT used to be quoted in the failure reason, the stdout line, the note the next enqueue copies the
// reason into, and the landed note, with core.Redact as the only protection; a credential written into the command
// line stayed in 6 of 9 shapes. The reason, the stdout, the memo and the landed note now carry a short id of the gate
// text; the text itself goes to the failure log only.
func TestAFLand_TheGateTextIsNotInTheReasonTheStdoutOrTheLedgerButTheLogHoldsIt(t *testing.T) {
	afSkipWithoutSh(t)
	// every command is replaced by a built-in that does nothing, so no shape starts a client or reaches a network
	const tail = ": the-tail-of-the-gate-text; exit 1"
	rows := []struct{ name, gate string }{
		{"assign", `@PW@=` + afD1Value + ` true`},
		{"pass_assign", `pass=` + afD1Value + ` true`},
		{"mysql_p", `: mysql -uroot -p` + afD1Value + ` db`},
		{"flag_space", `: mysqldump --@PW@ ` + afD1Value + ` db`},
		{"token_flag", `: gh --@TK@ ` + afD1Value + ` api`},
		{"url_no_dot", `: psql postgres@UC@` + afD1Value + `@db:5432/x`},
		{"curl_u", `: curl -u user:` + afD1Value + ` https://example.test`},
		{"glpat", `CI_@TK@=@GL@` + afD1Value + ` true`},
		{"plain_env", `SECRET_FOO=` + afD1Value + ` true`},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			db, repo, cand, st, epoch := afQueuedCard(t)
			gate := afD1Words.Replace(r.gate) + "\n" + tail
			out, errOut, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
			if err != nil {
				t.Fatalf("the verb failed: %v", err)
			}
			if _, _, eerr := afRunBoth("enqueue", "t1", "--db", db, "--epoch", strconv.FormatInt(epoch, 10), "--candidate", cand, "--repo", repo); eerr != nil {
				t.Fatalf("enqueue after the failure: %v", eerr)
			}
			if strings.Contains(out+errOut, afD1Value) {
				t.Errorf("the value is in the stdout or stderr: %.300q", out)
			}
			if strings.Contains(string(afLedgerBytes(t, db)), afD1Value) {
				t.Error("the value is in the ledger file")
			}
			notes, nerr := st.Notes(context.Background(), 0)
			if nerr != nil {
				t.Fatal(nerr)
			}
			for _, n := range notes {
				if strings.Contains(n.Text, afD1Value) {
					t.Errorf("the value is in a note: %.200q", n.Text)
				}
			}
			var res struct{ Outcome, Reason string }
			if jerr := json.Unmarshal([]byte(out), &res); jerr != nil || res.Outcome != "failed" {
				t.Fatalf("stdout %.200q: want outcome failed", out)
			}
			if !afGateIDPattern.MatchString(res.Reason) {
				t.Errorf("the reason carries no gate id: %.300q", res.Reason)
			}
			m := afD1LogPath.FindStringSubmatch(res.Reason)
			if m == nil {
				t.Fatalf("the reason names no log file: %.300q", res.Reason)
			}
			if log, rerr := os.ReadFile(m[1]); rerr != nil || !strings.Contains(string(log), "the-tail-of-the-gate-text") {
				t.Errorf("the log %s does not hold the gate text: %.300q %v", m[1], log, rerr)
			}
		})
	}
}

// N29: the same on a gate that passes: the landed note carries the id, not the text.
func TestAFLand_TheLandedNoteCarriesTheGateIDAndNotTheGateText(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	gate := afD1Words.Replace(`pass=` + afD1Value + ` true`)
	out, errOut, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	if err != nil {
		t.Fatalf("the verb failed: %v", err)
	}
	if strings.Contains(out+errOut, afD1Value) || strings.Contains(string(afLedgerBytes(t, db)), afD1Value) {
		t.Error("the gate text's value is in the stdout or the ledger")
	}
	notes, nerr := st.Notes(context.Background(), 0)
	if nerr != nil {
		t.Fatal(nerr)
	}
	landed := false
	for _, n := range notes {
		if n.Kind == "landed" {
			landed = true
			if !afGateIDPattern.MatchString(n.Text) {
				t.Errorf("the landed note carries no gate id: %.300q", n.Text)
			}
		}
	}
	if !landed {
		t.Errorf("no landed note: %.300q", out)
	}
}

// afRepairACLEntries is the ACL entries `ls -lde` prints under the permission line of path (none when it has no ACL).
func afRepairACLEntries(t *testing.T, path string) []string {
	t.Helper()
	out, err := exec.Command("ls", "-lde", path).Output()
	if err != nil {
		t.Fatalf("ls -lde %s: %v", path, err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")[1:]
}

// N22: the ledger lives below a directory with an inheritable ACL (TMPDIR points inside it, as a project under a
// shared directory would). The log of a failed gate may hold unmasked output, so the file and its directory must end
// without the ACL the 0600 mode does not remove.
func TestAFLand_TheLogFileEndsWithoutAnInheritedACL(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS inheritable ACLs")
	}
	afSkipWithoutSh(t)
	parent, err := os.MkdirTemp("", "moai-af-acl-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(parent) })
	acl := "everyone allow read,execute,list,search,readattr,readextattr,readsecurity,file_inherit,directory_inherit"
	if out, cerr := exec.Command("chmod", "+a", acl, parent).CombinedOutput(); cerr != nil {
		t.Fatalf("chmod +a: %v %s", cerr, out)
	}
	t.Setenv("TMPDIR", parent) // before the first t.TempDir(), which makes its root below TMPDIR
	db, repo, _, _, _ := afQueuedCard(t)
	if got := afRepairACLEntries(t, filepath.Dir(db)); len(got) == 0 {
		t.Fatal("the fixture is wrong: the ledger directory inherited no ACL")
	}
	_, reason, _ := afRepairLand(t, db, repo, "printf 'hello from the gate\\n'; exit 1")
	m := afD1LogPath.FindStringSubmatch(reason)
	if m == nil {
		t.Fatalf("the reason names no log file: %.300q", reason)
	}
	for _, p := range []string{filepath.Dir(m[1]), m[1]} {
		if got := afRepairACLEntries(t, p); len(got) != 0 {
			t.Errorf("%s has an ACL: %q", p, got)
		}
	}
}

// N23: gate-logs is a symbolic link. The verb must not follow it: no log in the target, its mode unchanged, and the
// failure is recorded with a reason that says logging failed.
func TestAFLand_ALogDirectoryThatIsALinkIsRefusedAndItsTargetIsLeftAlone(t *testing.T) {
	afSkipWithoutSh(t)
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on native Windows")
	}
	db, repo, _, st, _ := afQueuedCard(t)
	target := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(filepath.Dir(db), "gate-logs")); err != nil {
		t.Fatal(err)
	}
	outcome, reason, _ := afRepairLand(t, db, repo, "echo hello; exit 1")
	if outcome != "failed" || !strings.Contains(reason, "logging failed") {
		t.Errorf("outcome %q reason %.200q: want a failure that says logging failed", outcome, reason)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("the link's target has mode %v, want it untouched at 0755", fi.Mode().Perm())
	}
	if entries, _ := os.ReadDir(target); len(entries) != 0 {
		t.Errorf("the link's target holds %v, want nothing", entries)
	}
	if rows, err := st.Queue(context.Background()); err != nil || len(rows) != 1 || rows[0].FailReason == "" {
		t.Errorf("the failure was not recorded: %+v %v", rows, err)
	}
}
