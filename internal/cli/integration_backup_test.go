package cli

// integration_backup_test.go — REQ-TBS-001..004 / AC-TBS-003, AC-TBS-004,
// AC-TBS-005, AC-TBS-007, AC-TBS-009 (SPEC-TRUNK-BACKUP-SYNC-001 M2).
//
// Every test runs the verb against a bare origin created under t.TempDir()
// (REQ-TBS-006): no test names the development repository's remotes or any
// network URL, and the captured-command assertions check that every push's
// remote operand resolves under the test's temporary root. The per-child git
// environment carries the same isolation the candidate fixture established
// (repo-scoping scrub + GIT_CONFIG_GLOBAL/GIT_CONFIG_NOSYSTEM dropped,
// identity repo-local), so a developer's global gitconfig never influences a
// fixture and a fixture never writes outside its temp root.

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// backupPinnedRefspecPattern matches the PINNED refspec form the verb pushes
// (sync-audit round 2, R2-F1/R2-F2): the source side is the full SHA the verb
// resolved at invocation start, the target side the constant sync ref.
var backupPinnedRefspecPattern = regexp.MustCompile(`^[0-9a-f]{40}:` + regexp.QuoteMeta(backupTargetRef) + `$`)

// backupGit runs one git command under the fixture's isolation envelope.
func backupGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Per-child isolation (the candidateFixture envelope, card t1478):
	// repo-scoping vars scrubbed AND global/system config dropped (a
	// caller's commit.gpgsign=true with no signing program kills seed
	// commits before any assertion runs). Identity is repo-local.
	cmd.Env = append(gitenv.Scrub(os.Environ()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// backupGitAllowFail is backupGit for commands whose non-zero exit is a
// meaningful answer (git show-ref on an empty repository exits 1).
func backupGitAllowFail(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(gitenv.Scrub(os.Environ()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// backupFixture builds the hermetic layout: a bare origin plus a clone whose
// main branch carries one seed commit.
func backupFixture(t *testing.T) (base, origin, clone string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	base = t.TempDir()
	origin = filepath.Join(base, "origin.git")
	clone = filepath.Join(base, "clone")
	backupGit(t, base, "init", "-q", "--bare", origin)
	backupGit(t, base, "init", "-q", clone)
	backupGit(t, clone, "config", "user.email", "t@example.com")
	backupGit(t, clone, "config", "user.name", "t")
	backupGit(t, clone, "checkout", "-q", "-b", "main")
	backupGit(t, clone, "commit", "-q", "--allow-empty", "-m", "seed")
	backupGit(t, clone, "remote", "add", "origin", origin)
	return base, origin, clone
}

// backupGitCall is one captured (dir, argv) pair.
type backupGitCall struct {
	Dir  string
	Args []string
}

// backupRecorder wraps the scrubbed runner with a capture record, so the
// real-git tests can also assert on the exact argv REQ-TBS-002 pins.
type backupRecorder struct {
	calls []backupGitCall
}

func (r *backupRecorder) seam() integrationBackupSeams {
	return integrationBackupSeams{Git: func(dir string, args ...string) (string, error) {
		r.calls = append(r.calls, backupGitCall{Dir: dir, Args: append([]string(nil), args...)})
		return candidateScrubbedGit(dir, args...)
	}}
}

// backupRemoteRefs reads the bare origin's ref set (heads and tags), mapping
// full ref names to SHAs. An empty repository reads as an empty map.
func backupRemoteRefs(origin string) map[string]string {
	refs := map[string]string{}
	out, _ := backupGitAllowFail(origin, "show-ref")
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			refs[fields[1]] = fields[0]
		}
	}
	return refs
}

// assertBackupCapture enforces the invariants that ride EVERY verb run
// (REQ-TBS-002 + REQ-TBS-006): each push is exactly the constant flags and
// remote plus the PINNED refspec `<sha>:refs/heads/sync/main` (sync-audit
// round 2 — the transmitted commit is the printed one), the force flag never
// appears on any other command, and every push's remote operand resolves
// under the test's temporary root.
func assertBackupCapture(t *testing.T, base string, rec *backupRecorder) {
	t.Helper()
	constantPushPrefix := []string{"push", "--force", "--no-follow-tags", "--no-recurse-submodules", "origin"}
	for _, call := range rec.calls {
		if len(call.Args) > 0 && call.Args[0] == "push" {
			if len(call.Args) != len(constantPushPrefix)+1 || !reflect.DeepEqual(call.Args[:len(constantPushPrefix)], constantPushPrefix) {
				t.Errorf("captured push argv %v, want exactly %v <sha>:refs/heads/sync/main (REQ-TBS-002: the force flag, --no-follow-tags, and --no-recurse-submodules are constants)", call.Args, constantPushPrefix)
				continue
			}
			if !backupPinnedRefspecPattern.MatchString(call.Args[len(call.Args)-1]) {
				t.Errorf("push refspec %q does not match the pinned form <sha>:%s (REQ-TBS-001: the refspec pins the resolved SHA — never the bare branch refspec)", call.Args[len(call.Args)-1], backupTargetRef)
			}
			url := backupGit(t, call.Dir, "remote", "get-url", "origin")
			if !strings.HasPrefix(url, base) {
				t.Errorf("push remote operand resolves to %q, outside the test temp root %q (REQ-TBS-006: no real remote)", url, base)
			}
			continue
		}
		for _, a := range call.Args {
			if a == "--force" || a == "-f" {
				t.Errorf("force flag rides a non-push command %v (REQ-TBS-002: --force only ever accompanies the pinned refspec)", call.Args)
			}
		}
	}
}

// TestIntegrationBackupFirstPush — AC-TBS-003: against a temp remote with no
// sync/main, the verb creates refs/heads/sync/main at local main's commit X,
// changes no other remote ref, and its output names X and the target ref.
func TestIntegrationBackupFirstPush(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	want := backupGit(t, clone, "rev-parse", "refs/heads/main")

	rec := &backupRecorder{}
	var out strings.Builder
	if err := runIntegrationBackup(clone, rec.seam(), &out); err != nil {
		t.Fatalf("verb: %v", err)
	}
	refs := backupRemoteRefs(origin)
	if len(refs) != 1 || refs["refs/heads/sync/main"] != want {
		t.Errorf("remote refs after first push = %v, want exactly {refs/heads/sync/main: %s} (no other remote ref changes)", refs, want)
	}
	if !strings.Contains(out.String(), want) || !strings.Contains(out.String(), "refs/heads/sync/main") {
		t.Errorf("output %q: want it to name the pushed SHA %s and the target ref", out.String(), want)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupRegistered — the verb joins the integration group's
// command list (the integration.go registration site).
func TestIntegrationBackupRegistered(t *testing.T) {
	t.Parallel()

	for _, sub := range newIntegrationCmd().Commands() {
		if sub.Name() == "backup" {
			if sub.Flags().NFlag() != 0 {
				t.Errorf("backup carries %d flags, want 0 (REQ-TBS-001: the verb takes no flags)", sub.Flags().NFlag())
			}
			return
		}
	}
	t.Fatal("integration group does not register a backup subcommand")
}

// TestIntegrationBackupCmdRunE — the cobra path itself: executed with the
// working directory inside the temp clone, the command pushes the clone's
// main to its temp origin and prints SHA + target ref; an extra argument is
// refused (REQ-TBS-001: the verb takes no arguments). t.Chdir forbids
// t.Parallel.
func TestIntegrationBackupCmdRunE(t *testing.T) {
	_, origin, clone := backupFixture(t)
	want := backupGit(t, clone, "rev-parse", "refs/heads/main")

	t.Chdir(clone)
	cmd := newIntegrationBackupCmd()
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute backup from the clone: %v", err)
	}
	if !strings.Contains(out.String(), want) || !strings.Contains(out.String(), "refs/heads/sync/main") {
		t.Errorf("command output %q: want the pushed SHA %s and the target ref", out.String(), want)
	}
	if refs := backupRemoteRefs(origin); refs["refs/heads/sync/main"] != want {
		t.Errorf("remote refs = %v, want refs/heads/sync/main at %s", refs, want)
	}

	refused := newIntegrationBackupCmd()
	refused.SetArgs([]string{"unexpected"})
	if err := refused.Execute(); err == nil {
		t.Error("want an argument refusal: the verb takes no arguments (REQ-TBS-001)")
	}
}

// TestIntegrationBackupNonFFForce — AC-TBS-004 (force scoping): with the
// remote's sync/main at a commit that is not an ancestor of local main (the
// rewound-main shape, built by amending the seed), the push succeeds and
// updates the ref.
func TestIntegrationBackupNonFFForce(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	rec := &backupRecorder{}
	if err := runIntegrationBackup(clone, rec.seam(), &strings.Builder{}); err != nil {
		t.Fatalf("first verb run: %v", err)
	}
	oldTip := backupRemoteRefs(origin)["refs/heads/sync/main"]

	// Rewind: amend the root commit so local main moves to an unrelated
	// SHA — the remote's sync/main tip is no longer an ancestor of it.
	// --allow-empty because the seed commit itself is empty.
	backupGit(t, clone, "commit", "-q", "--allow-empty", "--amend", "-m", "rewritten")
	newTip := backupGit(t, clone, "rev-parse", "refs/heads/main")
	if newTip == oldTip {
		t.Fatal("fixture failed to produce a non-FF shape")
	}

	rec2 := &backupRecorder{}
	if err := runIntegrationBackup(clone, rec2.seam(), &strings.Builder{}); err != nil {
		t.Fatalf("non-FF push must succeed through the constant force flag: %v", err)
	}
	if got := backupRemoteRefs(origin)["refs/heads/sync/main"]; got != newTip {
		t.Errorf("remote sync/main = %s after rewound push, want %s", got, newTip)
	}
	assertBackupCapture(t, base, rec)
	assertBackupCapture(t, base, rec2)
}

// TestIntegrationBackupMissingMainRefusal — AC-TBS-005: when refs/heads/main
// does not resolve, the verb exits non-zero with guidance naming the missing
// ref and the constant target, and no push command is executed.
func TestIntegrationBackupMissingMainRefusal(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	trunk := filepath.Join(base, "trunk-only")
	backupGit(t, base, "init", "-q", "--bare", origin)
	backupGit(t, base, "init", "-q", trunk)
	backupGit(t, trunk, "config", "user.email", "t@example.com")
	backupGit(t, trunk, "config", "user.name", "t")
	backupGit(t, trunk, "checkout", "-q", "-b", "trunk")
	backupGit(t, trunk, "commit", "-q", "--allow-empty", "-m", "seed")
	backupGit(t, trunk, "remote", "add", "origin", origin)

	rec := &backupRecorder{}
	err := runIntegrationBackup(trunk, rec.seam(), &strings.Builder{})
	if err == nil {
		t.Fatal("want a non-zero-exit refusal when refs/heads/main does not resolve")
	}
	for _, want := range []string{"refs/heads/main", "refs/heads/sync/main"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q: want it to name %q", err.Error(), want)
		}
	}
	for _, call := range rec.calls {
		if len(call.Args) > 0 && call.Args[0] == "push" {
			t.Errorf("captured %v — no push may be executed on the refusal path", call.Args)
		}
	}
	if refs := backupRemoteRefs(origin); len(refs) != 0 {
		t.Errorf("remote refs = %v, want none (the refusal precedes any push)", refs)
	}
}

// TestIntegrationBackupPushFailureBothStreams — AC-TBS-009: a pre-receive
// hook rejecting the update makes the verb exit non-zero, print BOTH captured
// streams of the failed push, and issue no git command after it.
func TestIntegrationBackupPushFailureBothStreams(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("pre-receive fixture is a POSIX shell script; the both-streams capture itself is platform-neutral")
	}
	base, origin, clone := backupFixture(t)
	hook := filepath.Join(origin, "hooks", "pre-receive")
	script := "#!/bin/sh\necho prereceive-stdout-marker\necho prereceive-stderr-marker >&2\nexit 1\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := &backupRecorder{}
	err := runIntegrationBackup(clone, rec.seam(), &strings.Builder{})
	if err == nil {
		t.Fatal("want a non-zero exit when the remote rejects the update")
	}
	// The hook writes to both of ITS streams; git forwards them onto the
	// push command's captured streams, so asserting both markers in the
	// verb's error proves both captured streams were printed.
	for _, marker := range []string{"prereceive-stdout-marker", "prereceive-stderr-marker"} {
		if !strings.Contains(err.Error(), marker) {
			t.Errorf("failure report %q: want it to carry %q (both captured streams are printed)", err.Error(), marker)
		}
	}
	if !strings.Contains(err.Error(), "stdout:") || !strings.Contains(err.Error(), "stderr:") {
		t.Errorf("failure report %q: want it to label both captured stream sections", err.Error())
	}
	if len(rec.calls) == 0 {
		t.Fatal("no captured commands")
	}
	last := rec.calls[len(rec.calls)-1]
	if len(last.Args) == 0 || last.Args[0] != "push" {
		t.Errorf("last captured command %v — the failed push must be the final git command (no command after the failure)", last.Args)
	}
	if refs := backupRemoteRefs(origin); len(refs) != 0 {
		t.Errorf("remote refs = %v, want none (the update was rejected)", refs)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupFollowTagsNoLeak — AC-TBS-004 (no tag leak): with the
// clone's gitconfig setting push.followTags=true and an annotated v* tag
// reachable from main, the push still updates ONLY refs/heads/sync/main —
// the --no-follow-tags constant holds.
func TestIntegrationBackupFollowTagsNoLeak(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	backupGit(t, clone, "config", "push.followTags", "true")
	backupGit(t, clone, "tag", "-a", "v9.9.9", "-m", "annotated")
	want := backupGit(t, clone, "rev-parse", "refs/heads/main")

	rec := &backupRecorder{}
	if err := runIntegrationBackup(clone, rec.seam(), &strings.Builder{}); err != nil {
		t.Fatalf("verb: %v", err)
	}
	refs := backupRemoteRefs(origin)
	if len(refs) != 1 || refs["refs/heads/sync/main"] != want {
		t.Errorf("remote refs = %v, want exactly {refs/heads/sync/main: %s} — no tag ref may ride the push despite push.followTags=true (the --no-follow-tags constant holds)", refs, want)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupRecursionHazard — S1 (sync-audit round 1, P1): under a
// caller's push.recurseSubmodules=only the argv without the constant
// --no-recurse-submodules exits 0 as a silent no-op false success — the verb
// prints success while the remote gains NO refs/heads/sync/main. The constant
// holds even where the gitconfig sets push.recurseSubmodules (amended
// REQ-TBS-001/002), so the push must land sync/main regardless.
func TestIntegrationBackupRecursionHazard(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	backupGit(t, clone, "config", "push.recurseSubmodules", "only")

	rec := &backupRecorder{}
	var out strings.Builder
	if err := runIntegrationBackup(clone, rec.seam(), &out); err != nil {
		t.Fatalf("verb: %v", err)
	}
	want := backupGit(t, clone, "rev-parse", "refs/heads/main")
	refs := backupRemoteRefs(origin)
	if refs["refs/heads/sync/main"] != want {
		t.Errorf("remote refs = %v — push.recurseSubmodules=only swallowed the push as a silent no-op (exit 0, success printed: %q); want refs/heads/sync/main at %s", refs, out.String(), want)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupReportsPushedSHA — S3 (sync-audit round 1) as amended
// by round 2 (R2-F1/R2-F2): the verb PINS the SHA at the source — it pushes
// the pinned refspec `<sha>:refs/heads/sync/main` — so when main advances
// between the verb's resolution and its push, the backup carries (and the
// success line names) the PINNED commit, never the post-advance tip, and no
// post-push remote read can rename the report. The fixture advances main the
// moment rev-parse answers (seam-level), so the pinned and advanced SHAs
// differ and the printed one is falsifiable.
func TestIntegrationBackupReportsPushedSHA(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	rec := &backupRecorder{}
	pinned := ""
	seam := integrationBackupSeams{Git: func(dir string, args ...string) (string, error) {
		rec.calls = append(rec.calls, backupGitCall{Dir: dir, Args: append([]string(nil), args...)})
		if len(args) > 0 && args[0] == "rev-parse" {
			pinned = backupGit(t, clone, "rev-parse", "refs/heads/main")
			// main advances AFTER the verb pinned it, BEFORE the push.
			backupGit(t, clone, "commit", "-q", "--allow-empty", "-m", "advanced mid-flight")
			return pinned, nil
		}
		return candidateScrubbedGit(dir, args...)
	}}

	var out strings.Builder
	if err := runIntegrationBackup(clone, seam, &out); err != nil {
		t.Fatalf("verb: %v", err)
	}
	advanced := backupGit(t, clone, "rev-parse", "refs/heads/main")
	if pinned == advanced {
		t.Fatal("fixture failed: main did not advance between pinning and push")
	}
	if remote := backupRemoteRefs(origin)["refs/heads/sync/main"]; remote != pinned {
		t.Errorf("remote sync/main = %s, want the PINNED tip %s — the pinned refspec transmits the pinned commit even though main advanced to %s mid-flight", remote, pinned, advanced)
	}
	if !strings.Contains(out.String(), pinned) {
		t.Errorf("success line %q: want the PINNED transmitted SHA %s (REQ-TBS-001: the printed SHA is the pinned transmitted SHA)", out.String(), pinned)
	}
	if strings.Contains(out.String(), advanced) {
		t.Errorf("success line %q: names the post-advance tip %s, which this run never transmitted", out.String(), advanced)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupScrubbedEnv — AC-TBS-007 (the t1478 shape): with the
// process environment carrying inherited GIT_DIR/GIT_WORK_TREE pointing at a
// foreign temporary repository, the verb still pushes the CALLER's repository
// — repository-location variables were removed before exec.
func TestIntegrationBackupScrubbedEnv(t *testing.T) {
	// t.Setenv mutates the process environment: no t.Parallel here.
	_, origin, clone := backupFixture(t)
	foreignBase := t.TempDir()
	foreign := filepath.Join(foreignBase, "foreign")
	backupGit(t, foreignBase, "init", "-q", foreign)
	backupGit(t, foreign, "config", "user.email", "t@example.com")
	backupGit(t, foreign, "config", "user.name", "t")
	backupGit(t, foreign, "checkout", "-q", "-b", "main")
	backupGit(t, foreign, "commit", "-q", "--allow-empty", "-m", "foreign seed")

	t.Setenv("GIT_DIR", filepath.Join(foreign, ".git"))
	t.Setenv("GIT_WORK_TREE", foreign)

	want := backupGit(t, clone, "rev-parse", "refs/heads/main")
	if err := runIntegrationBackup(clone, integrationBackupSeams{}, &strings.Builder{}); err != nil {
		t.Fatalf("verb under inherited GIT_DIR: %v", err)
	}
	refs := backupRemoteRefs(origin)
	if refs["refs/heads/sync/main"] != want {
		t.Errorf("clone's origin refs = %v, want refs/heads/sync/main at %s — the push must target the caller's repository", refs, want)
	}
	foreignRefs := backupRemoteRefs(foreign)
	if _, ok := foreignRefs["refs/heads/sync/main"]; ok {
		t.Errorf("foreign repository gained refs/heads/sync/main (%v) — the scrub failed and the push landed in the wrong repository", foreignRefs)
	}
}

// TestIntegrationBackupPushURLDivergence — R2-F1 (sync-audit round 2, P2):
// with `remote.origin.pushurl` differing from the fetch url, the pushed-to
// repository and the queried one diverge — so a success line derived from a
// POST-PUSH remote read names a SHA that was never pushed (ls-remote queries
// the fetch url). The fixture gives origin a separate pushurl bare repo and
// pre-loads the fetch repo's sync/main with an older copy of main; the verb
// must report the SHA THIS RUN pinned (== the local main tip it resolved),
// not whatever the fetch repo's sync/main holds.
func TestIntegrationBackupPushURLDivergence(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	fetchRepo := filepath.Join(base, "fetch.git")
	pushRepo := filepath.Join(base, "push.git")
	clone := filepath.Join(base, "clone")
	backupGit(t, base, "init", "-q", "--bare", fetchRepo)
	backupGit(t, base, "init", "-q", "--bare", pushRepo)
	backupGit(t, base, "init", "-q", clone)
	backupGit(t, clone, "config", "user.email", "t@example.com")
	backupGit(t, clone, "config", "user.name", "t")
	backupGit(t, clone, "checkout", "-q", "-b", "main")
	backupGit(t, clone, "commit", "-q", "--allow-empty", "-m", "seed")
	backupGit(t, clone, "remote", "add", "origin", fetchRepo)
	// The fetch repo's sync/main holds an older copy of main.
	backupGit(t, clone, "push", "-q", fetchRepo, "refs/heads/main:refs/heads/sync/main")
	oldFetchSHA := backupGit(t, clone, "rev-parse", "refs/heads/main")
	// main advances past the fetch repo's copy, and the pushurl diverges.
	backupGit(t, clone, "commit", "-q", "--allow-empty", "-m", "second")
	backupGit(t, clone, "remote", "set-url", "--push", "origin", pushRepo)
	pinned := backupGit(t, clone, "rev-parse", "refs/heads/main")
	if pinned == oldFetchSHA {
		t.Fatal("fixture failed: the fetch repo's copy did not diverge from main")
	}

	rec := &backupRecorder{}
	var out strings.Builder
	if err := runIntegrationBackup(clone, rec.seam(), &out); err != nil {
		t.Fatalf("verb: %v", err)
	}
	if got := backupRemoteRefs(pushRepo)["refs/heads/sync/main"]; got != pinned {
		t.Errorf("pushurl repo's sync/main = %s, want the pinned tip %s", got, pinned)
	}
	if got := backupRemoteRefs(fetchRepo)["refs/heads/sync/main"]; got != oldFetchSHA {
		t.Errorf("fetch repo's sync/main = %s, want the untouched older copy %s (the verb pushes via the pushurl)", got, oldFetchSHA)
	}
	if !strings.Contains(out.String(), pinned) {
		t.Errorf("success line %q: want THIS run's pinned SHA %s — under a divergent pushurl the report must name the transmitted commit, never a post-push read of the fetch url", out.String(), pinned)
	}
	if strings.Contains(out.String(), oldFetchSHA) {
		t.Errorf("success line %q: names the fetch repo's stale sync/main SHA %s, which this run never pushed", out.String(), oldFetchSHA)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupCompetitorPostPush — R2-F2 (sync-audit round 2, P2):
// a competing writer force-pushes sync/main immediately AFTER this verb's
// push (seam-level: the recording seam performs the competitor push the
// moment the verb's push returns). The verb's success line must still name
// THIS run's pinned SHA — reporting someone else's push is a false report.
func TestIntegrationBackupCompetitorPostPush(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	rec := &backupRecorder{}
	pinned, competitor := "", ""
	seam := integrationBackupSeams{Git: func(dir string, args ...string) (string, error) {
		rec.calls = append(rec.calls, backupGitCall{Dir: dir, Args: append([]string(nil), args...)})
		if len(args) > 0 && args[0] == "rev-parse" {
			pinned = backupGit(t, clone, "rev-parse", "refs/heads/main")
		}
		outStr, err := candidateScrubbedGit(dir, args...)
		if len(args) > 0 && args[0] == "push" && err == nil {
			// The competitor lands strictly after this verb's push returned.
			backupGit(t, clone, "commit", "-q", "--allow-empty", "-m", "competitor commit")
			competitor = backupGit(t, clone, "rev-parse", "refs/heads/main")
			backupGit(t, clone, "push", "-q", "--force", origin, "refs/heads/main:refs/heads/sync/main")
		}
		return outStr, err
	}}

	var out strings.Builder
	if err := runIntegrationBackup(clone, seam, &out); err != nil {
		t.Fatalf("verb: %v", err)
	}
	if competitor == "" || competitor == pinned {
		t.Fatal("fixture failed: the competitor did not produce a different SHA")
	}
	if remote := backupRemoteRefs(origin)["refs/heads/sync/main"]; remote != competitor {
		t.Fatalf("fixture guard: remote sync/main = %s, want the competitor's %s", remote, competitor)
	}
	if !strings.Contains(out.String(), pinned) {
		t.Errorf("success line %q: want THIS run's pinned SHA %s — a competitor that overwrites sync/main after the push must not rename the verb's report", out.String(), pinned)
	}
	if strings.Contains(out.String(), competitor) {
		t.Errorf("success line %q: names the competitor's SHA %s, which this verb never pushed", out.String(), competitor)
	}
	assertBackupCapture(t, base, rec)
}

// TestIntegrationBackupPinnedRefspecArgv — R2 regression (sync-audit round
// 2): the push argv reaches git with the exact pinned refspec form
// `<sha>:refs/heads/sync/main`, where <sha> is the local main tip resolved
// before the verb ran — never the bare branch refspec.
func TestIntegrationBackupPinnedRefspecArgv(t *testing.T) {
	t.Parallel()

	base, origin, clone := backupFixture(t)
	pinned := backupGit(t, clone, "rev-parse", "refs/heads/main")
	rec := &backupRecorder{}
	if err := runIntegrationBackup(clone, rec.seam(), &strings.Builder{}); err != nil {
		t.Fatalf("verb: %v", err)
	}
	var pushRefspecs []string
	for _, call := range rec.calls {
		if len(call.Args) > 0 && call.Args[0] == "push" {
			pushRefspecs = append(pushRefspecs, call.Args[len(call.Args)-1])
		}
	}
	if len(pushRefspecs) != 1 {
		t.Fatalf("captured %d push commands, want exactly 1", len(pushRefspecs))
	}
	if pushRefspecs[0] != pinned+":"+backupTargetRef {
		t.Errorf("push refspec %q, want the exact pinned form %q", pushRefspecs[0], pinned+":"+backupTargetRef)
	}
	if refs := backupRemoteRefs(origin); refs["refs/heads/sync/main"] != pinned {
		t.Errorf("remote refs = %v, want refs/heads/sync/main at the pinned %s", refs, pinned)
	}
	assertBackupCapture(t, base, rec)
}
