package cli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// candidateFixture builds a hermetic three-repo layout for the candidate
// verb: a bare origin, a primary clone carrying the integration branch
// (develop), and a card worktree on a WT- branch one commit ahead. The
// primary clone dir is the project root (config + candidate record store).
func candidateFixture(t *testing.T) (root, cardWT string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	primary := filepath.Join(base, "primary")
	cardWT = filepath.Join(base, "wt-card")

	run := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// Per-child isolation (card t1478 M2 repair): repo-scoping vars are
		// scrubbed (an inherited GIT_DIR would redirect the verification
		// reads at another repository) AND the global config is dropped (a
		// caller's commit.gpgsign=true with no signing program killed the
		// seed commits before any assertion ran). Identity is repo-local
		// (set just below), so a config-free child is enough.
		cmd.Env = append(gitenv.Scrub(os.Environ()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("", "init", "-q", "--bare", origin)
	run("", "init", "-q", primary)
	run(primary, "config", "user.email", "t@example.com")
	run(primary, "config", "user.name", "t")
	run(primary, "checkout", "-q", "-b", "develop")
	run(primary, "commit", "-q", "--allow-empty", "-m", "seed")
	run(primary, "remote", "add", "origin", origin)
	run(primary, "push", "-q", "-u", "origin", "develop")
	run(primary, "worktree", "add", "-q", "-b", "WT-card", cardWT)
	run(cardWT, "commit", "-q", "--allow-empty", "-m", "card work")
	return primary, cardWT
}

// countingGitSeam records every invocation and refuses nothing; tests use
// the count to prove a refusal happened before any git — and therefore any
// remote — call.
func countingGitSeam(calls *int) func(dir string, args ...string) (string, error) {
	return func(dir string, args ...string) (string, error) {
		*calls++
		runner := factorylane.ExecGitRunner{Dir: dir}
		return runner.Git(args...)
	}
}

func TestIntegrationCandidate(t *testing.T) {
	t.Parallel()

	t.Run("gate off refuses naming the key before any git call", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "false", "true")
		calls := 0
		_, err := runIntegrationCandidate(integrationCandidateInput{
			Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
		}, integrationCandidateSeams{Git: countingGitSeam(&calls)})
		if err == nil {
			t.Fatal("want refusal with the key off")
		}
		if !strings.Contains(err.Error(), "workflow.candidate_ci.enabled") {
			t.Errorf("refusal %q: want it to name workflow.candidate_ci.enabled", err)
		}
		if calls != 0 {
			t.Errorf("git calls: got %d, want 0 (the gate refuses before any git, so no remote can be reached)", calls)
		}
	})

	t.Run("invalid card id refuses before any git call", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		calls := 0
		_, err := runIntegrationCandidate(integrationCandidateInput{
			Root: root, CardID: "../escape", CardWorktree: cardWT, IntegrationBranch: "develop",
		}, integrationCandidateSeams{Git: countingGitSeam(&calls)})
		if err == nil {
			t.Fatal("want refusal for an invalid card id")
		}
		if calls != 0 {
			t.Errorf("git calls: got %d, want 0 (id validation precedes every git and remote call)", calls)
		}
	})

	t.Run("happy path pushes a two-parent candidate and writes a pending record", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		rec, err := runIntegrationCandidate(integrationCandidateInput{
			Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
		}, integrationCandidateSeams{Now: func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }})
		if err != nil {
			t.Fatalf("runIntegrationCandidate: %v", err)
		}
		if rec.Verdict != factory.CandidateVerdictPending {
			t.Errorf("verdict: got %q, want %q", rec.Verdict, factory.CandidateVerdictPending)
		}
		if rec.CandidateBranch != "ci/t9001" {
			t.Errorf("candidate branch: got %q, want ci/t9001", rec.CandidateBranch)
		}
		if rec.IntegrationBranch != "develop" {
			t.Errorf("integration branch: got %q, want develop", rec.IntegrationBranch)
		}

		// The commit graph is the witness (AC-CCI-002-1): exactly two
		// parent lines, the set {tip, pinned}.
		cat := candidateGit(t, cardWT, "cat-file", "-p", rec.CandidateSHA)
		var parents []string
		for _, line := range strings.Split(cat, "\n") {
			if strings.HasPrefix(line, "parent ") {
				parents = append(parents, strings.TrimPrefix(line, "parent "))
			}
		}
		if len(parents) != 2 {
			t.Fatalf("parent lines: got %d (%v), want exactly 2", len(parents), parents)
		}
		tip := candidateGit(t, cardWT, "rev-parse", "refs/heads/develop")
		pinned := candidateGit(t, cardWT, "rev-parse", "refs/heads/WT-card")
		if !sameStringSet(parents, []string{tip, pinned}) {
			t.Errorf("parents: got %v, want {%s, %s}", parents, tip, pinned)
		}

		// The pushed branch tip IS the candidate (the record/branch
		// identity clause) — read from the ORIGIN, not the local ref.
		remoteTip := candidateGit(t, cardWT, "ls-remote", "origin", "refs/heads/ci/t9001")
		if !strings.HasPrefix(remoteTip, rec.CandidateSHA) {
			t.Errorf("origin ci/t9001: got %q, want it to name the candidate %s", remoteTip, rec.CandidateSHA)
		}

		// The record reads back from the store.
		stored, err := factory.ReadCandidateRecord(root, "t9001", rec.PinnedSHA)
		if err != nil {
			t.Fatalf("ReadCandidateRecord: %v", err)
		}
		if stored.CandidateSHA != rec.CandidateSHA {
			t.Errorf("stored candidate sha: got %s, want %s", stored.CandidateSHA, rec.CandidateSHA)
		}
		if stored.PushedAt != "2026-10-09T12:00:00Z" {
			t.Errorf("pushed_at: got %s", stored.PushedAt)
		}
	})

	t.Run("conflict refuses naming the card and writes no branch or record", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		// add/add conflict on shared.txt: the card branch adds it in the
		// card worktree; develop adds it in the PRIMARY (where develop is
		// checked out — a worktree cannot check out another's branch).
		writeAndCommit(t, cardWT, "shared.txt", "card line\n", "card touches shared.txt")
		writeAndCommit(t, root, "shared.txt", "develop line\n", "develop touches shared.txt")

		_, err := runIntegrationCandidate(integrationCandidateInput{
			Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
		}, integrationCandidateSeams{})
		if err == nil {
			t.Fatal("want a conflict refusal")
		}
		if !strings.Contains(err.Error(), "t9001") {
			t.Errorf("refusal %q: want it to name the card", err)
		}
		if !strings.Contains(err.Error(), "conflict") {
			t.Errorf("refusal %q: want it to name the conflict", err)
		}
		remote := candidateGit(t, cardWT, "ls-remote", "origin", "refs/heads/ci/t9001")
		if remote != "" {
			t.Errorf("origin ci/t9001: got %q, want no candidate branch after a conflict", remote)
		}
	})

	t.Run("re-candidate replaces the branch tip in place", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		in := integrationCandidateInput{Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop"}
		first, err := runIntegrationCandidate(in, integrationCandidateSeams{})
		if err != nil {
			t.Fatalf("first candidate: %v", err)
		}
		writeAndCommit(t, cardWT, "more.txt", "more\n", "card work 2")
		second, err := runIntegrationCandidate(in, integrationCandidateSeams{})
		if err != nil {
			t.Fatalf("re-candidate: %v", err)
		}
		if second.CandidateSHA == first.CandidateSHA {
			t.Fatal("re-candidate produced the same commit")
		}
		remoteTip := candidateGit(t, cardWT, "ls-remote", "origin", "refs/heads/ci/t9001")
		if !strings.HasPrefix(remoteTip, second.CandidateSHA) {
			t.Errorf("origin ci/t9001 after re-candidate: got %q, want the new candidate %s (replace-in-place)", remoteTip, second.CandidateSHA)
		}
		// The prior record stays readable but superseded (keyed by its own
		// pinned SHA).
		if _, err := factory.ReadCandidateRecord(root, "t9001", first.PinnedSHA); err != nil {
			t.Errorf("prior record after re-candidate: %v", err)
		}
	})

	t.Run("card SHA equal to the integration tip refuses before any push or record", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		// Park the card branch AT the develop tip: commit-tree dedupes
		// identical parents, so a candidate from here carries ONE parent —
		// not the two-parent merge shape AC-CCI-002-1 witnesses.
		tip := candidateGit(t, root, "rev-parse", "refs/heads/develop")
		candidateGit(t, cardWT, "update-ref", "refs/heads/WT-card", tip)
		rec, err := runIntegrationCandidate(integrationCandidateInput{
			Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
		}, integrationCandidateSeams{})
		if err == nil {
			t.Fatalf("want a refusal, got candidate %s (a one-parent commit would have been pushed)", rec.CandidateSHA)
		}
		if !strings.Contains(err.Error(), tip[:12]) {
			t.Errorf("refusal %q: want it to name the shared SHA %s", err, tip[:12])
		}
		remote := candidateGit(t, cardWT, "ls-remote", "origin", "refs/heads/ci/t9001")
		if remote != "" {
			t.Errorf("origin ci/t9001: got %q, want no candidate branch after the tip-equality refusal", remote)
		}
		if _, err := factory.ReadCandidateRecord(root, "t9001", tip); !errors.Is(err, factory.ErrCandidateRecordAbsent) {
			t.Errorf("record after tip-equality refusal: %v, want absent", err)
		}
	})

	t.Run("push failure names the stage and leaves the prior verdict untouched", func(t *testing.T) {
		t.Parallel()
		root, cardWT := candidateFixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		in := integrationCandidateInput{Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop"}
		first, err := runIntegrationCandidate(in, integrationCandidateSeams{})
		if err != nil {
			t.Fatalf("first candidate: %v", err)
		}
		// A previously GREEN verdict stays exactly what it was.
		green := first
		green.Verdict = factory.CandidateVerdictGreen
		if err := factory.WriteCandidateRecord(root, green); err != nil {
			t.Fatalf("seed green verdict: %v", err)
		}
		writeAndCommit(t, cardWT, "again.txt", "x\n", "card work 3")
		_, err = runIntegrationCandidate(in, integrationCandidateSeams{
			Git: func(dir string, args ...string) (string, error) {
				if args[0] == "push" {
					return "", errors.New("git push: remote refused (test double)")
				}
				runner := factorylane.ExecGitRunner{Dir: dir}
				return runner.Git(args...)
			},
		})
		if err == nil {
			t.Fatal("want the push failure to escape")
		}
		if !strings.Contains(err.Error(), "push") {
			t.Errorf("failure %q: want it to name the stage reached", err)
		}
		stored, err := factory.ReadCandidateRecord(root, "t9001", green.PinnedSHA)
		if err != nil {
			t.Fatalf("read prior record after failed push: %v", err)
		}
		if stored.Verdict != factory.CandidateVerdictGreen {
			t.Errorf("prior verdict after failed push: got %q, want %q (a failed push never mutates a verdict)", stored.Verdict, factory.CandidateVerdictGreen)
		}
	})
}

// candidateGit runs one read-only git command in dir and returns trimmed
// stdout, failing the test on error. The child carries the same per-child
// config isolation candidateFixture's runner does.
func candidateGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Same isolation as the fixture's runner: repo-scoping vars scrubbed,
	// global config dropped — a verification read must look at the repo
	// dir names, never at whatever the process environment designates.
	cmd.Env = append(gitenv.Scrub(os.Environ()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s (in %s): %v", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out))
}

// writeAndCommit writes file content in dir and commits it on the checked
// out branch.
func writeAndCommit(t *testing.T, dir, file, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	candidateGit(t, dir, "add", file)
	candidateGit(t, dir, "commit", "-q", "-m", msg)
}

// TestIntegrationCandidateRecandidatePreservesVerdict pins the
// identical-SHA re-candidate contract (card t1478 M2 repair): fixed git
// dates make both invocations construct the SAME candidate commit — same
// tree, same parents, same message, same author/committer dates — so the
// re-push does not move the remote ref and fires no CI event. The only
// verdict this candidate will ever have is the one already recorded, so
// the re-candidate must PRESERVE verdict, run id, and observation time,
// refreshing only PushedAt. Overwriting with pending would strand a green
// candidate permanently unverifiable.
func TestIntegrationCandidateRecandidatePreservesVerdict(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-09T12:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-09T12:00:00Z")
	fixed := func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	in := integrationCandidateInput{Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop"}
	first, err := runIntegrationCandidate(in, integrationCandidateSeams{Now: fixed})
	if err != nil {
		t.Fatalf("first candidate: %v", err)
	}
	// An explicit observation records green + run identity (REQ-CCI-010's
	// shape, seeded here the way the verdict observer writes it).
	observed := first
	observed.Verdict = factory.CandidateVerdictGreen
	observed.RunID = "run-abc123"
	observed.ObservedAt = "2026-10-09T13:00:00Z"
	if err := factory.WriteCandidateRecord(root, observed); err != nil {
		t.Fatalf("seed observed verdict: %v", err)
	}

	second, err := runIntegrationCandidate(in, integrationCandidateSeams{Now: fixed})
	if err != nil {
		t.Fatalf("re-candidate: %v", err)
	}
	if second.CandidateSHA != first.CandidateSHA {
		t.Fatalf("fixture: the git dates did not pin the commit — first %s, second %s (the test measures the identical-SHA path)",
			first.CandidateSHA[:12], second.CandidateSHA[:12])
	}
	stored, err := factory.ReadCandidateRecord(root, "t9001", second.PinnedSHA)
	if err != nil {
		t.Fatalf("read record after re-candidate: %v", err)
	}
	if stored.Verdict != factory.CandidateVerdictGreen {
		t.Errorf("verdict after identical re-candidate: got %q, want %q (the re-push fires no CI event; the recorded verdict is the only one)", stored.Verdict, factory.CandidateVerdictGreen)
	}
	if stored.RunID != "run-abc123" {
		t.Errorf("run id after identical re-candidate: got %q, want run-abc123", stored.RunID)
	}
	if stored.ObservedAt != "2026-10-09T13:00:00Z" {
		t.Errorf("observed_at after identical re-candidate: got %q, want the preserved observation", stored.ObservedAt)
	}
}

// TestIntegrationCandidatePushRecordSerialized pins the per-card
// push+record critical section (card t1478 M2 repair): while one
// candidate invocation holds the card's mutation lock mid-push, a second
// invocation of the same card must not interleave its own record write —
// the A-push → B-push+record → A-record order left remote=B, record=A.
// The second invocation waits the lock budget and refuses with the busy
// sentinel (transient, retry-me), never writing a divergent record.
func TestIntegrationCandidatePushRecordSerialized(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")
	in := integrationCandidateInput{Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop"}

	pushStarted := make(chan struct{})
	releasePush := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := runIntegrationCandidate(in, integrationCandidateSeams{
			Git: func(dir string, args ...string) (string, error) {
				if args[0] == "push" {
					close(pushStarted)
					<-releasePush
				}
				runner := factorylane.ExecGitRunner{Dir: dir}
				return runner.Git(args...)
			},
		})
		firstDone <- err
	}()
	select {
	case <-pushStarted:
	case err := <-firstDone:
		t.Fatalf("first candidate finished before its push ran: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatal("first candidate never reached its push")
	}

	// The second invocation contends on the same card's lock while the
	// first is mid-section; the wait budget expires and it refuses busy.
	_, err := runIntegrationCandidate(in, integrationCandidateSeams{})
	if err == nil {
		close(releasePush)
		t.Fatal("second candidate interleaved with the first's push+record section — the per-card serialization is absent")
	}
	if !errors.Is(err, factory.ErrCandidateMutationBusy) {
		t.Errorf("second candidate error %v: want it to wrap factory.ErrCandidateMutationBusy", err)
	}
	close(releasePush)
	if err := <-firstDone; err != nil {
		t.Fatalf("first candidate after release: %v", err)
	}
	// With the section free, the next invocation proceeds normally.
	if _, err := runIntegrationCandidate(in, integrationCandidateSeams{}); err != nil {
		t.Fatalf("candidate after the section freed: %v", err)
	}
}

// TestIntegrationCandidateCallerTreeGuard pins the foreign-caller refusal
// (card t1478 M2 repair): the verb builds the candidate from the CALLER's
// tree, so running it from a tree other than the card's recorded worktree
// would force-overwrite ci/<card> with a foreign candidate. The guard
// reads the factory card record's WorktreePath and refuses the mismatch,
// naming both trees.
func TestIntegrationCandidateCallerTreeGuard(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")
	// The factory db carries the card record the guard consults.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES('run-cli','active','t','t')`); err != nil {
		t.Fatalf("place run row: %v", err)
	}
	placeCardRow(t, db, "run-cli", "t9001", cardWT)
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}

	t.Run("matching tree admits", func(t *testing.T) {
		if err := candidateCallerTreeGuard(root, "t9001", cardWT, "run-cli"); err != nil {
			t.Errorf("guard(root, t9001, %s): %v", cardWT, err)
		}
	})
	t.Run("foreign tree refuses naming both", func(t *testing.T) {
		other := filepath.Join(filepath.Dir(cardWT), "wt-other")
		candidateGit(t, root, "worktree", "add", "-q", "-b", "WT-other", other)
		err := candidateCallerTreeGuard(root, "t9001", other, "run-cli")
		if err == nil {
			t.Fatal("want a refusal for a caller outside the card's recorded worktree")
		}
		if !strings.Contains(err.Error(), cardWT) {
			t.Errorf("refusal %q: want it to name the recorded worktree %s", err, cardWT)
		}
		if !strings.Contains(err.Error(), other) {
			t.Errorf("refusal %q: want it to name the caller's tree %s", err, other)
		}
	})
	t.Run("unknown card refuses fail-closed", func(t *testing.T) {
		if err := candidateCallerTreeGuard(root, "t9999", cardWT, "run-cli"); err == nil {
			t.Fatal("want a refusal when the card has no factory record to verify against")
		}
	})

	// The verb wires the guard: from the foreign tree, the command refuses
	// before any push.
	other := filepath.Join(filepath.Dir(cardWT), "wt-other")
	t.Chdir(other)
	cmd := newIntegrationCandidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--card", "t9001"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("the candidate verb ran from a foreign tree without refusing")
	}
	remote := candidateGit(t, root, "ls-remote", "origin", "refs/heads/ci/t9001")
	if remote != "" {
		t.Errorf("origin ci/t9001 after the foreign-tree run: got %q, want no candidate branch (the refusal precedes any remote call)", remote)
	}
}

// TestIntegrationCandidateIgnoresRepoScopingEnv pins the repo-scoping
// env scrub (card t1478 M2 repair, P1): the git environment beats
// cmd.Dir, so an inherited GIT_DIR/GIT_WORK_TREE made the verb build the
// candidate from ANOTHER repository's branch and force-push it to THAT
// repository's origin. Every git child must run with repo-scoping vars
// removed (gitenv.Env) — the candidate always targets the card worktree's
// own origin.
func TestIntegrationCandidateIgnoresRepoScopingEnv(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")

	// A second, unrelated repo with its own origin bare repo — the wrong
	// target the scrubbed env must keep the verb away from.
	base := filepath.Dir(root)
	otherOrigin := filepath.Join(base, "other-origin.git")
	other := filepath.Join(base, "other")
	candidateGit(t, "", "init", "-q", "--bare", otherOrigin)
	candidateGit(t, "", "init", "-q", other)
	candidateGit(t, other, "config", "user.email", "t@example.com")
	candidateGit(t, other, "config", "user.name", "t")
	candidateGit(t, other, "commit", "-q", "--allow-empty", "-m", "seed")
	candidateGit(t, other, "remote", "add", "origin", otherOrigin)
	candidateGit(t, other, "push", "-q", "-u", "origin", "HEAD:refs/heads/develop")

	// The hazard env: repo-scoping vars pointing at the OTHER repo.
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	rec, err := runIntegrationCandidate(integrationCandidateInput{
		Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
	}, integrationCandidateSeams{})
	if err != nil {
		t.Fatalf("candidate under repo-scoping env: %v", err)
	}

	// The card's OWN origin carries the candidate.
	own := candidateGit(t, cardWT, "ls-remote", "origin", "refs/heads/ci/t9001")
	if !strings.HasPrefix(own, rec.CandidateSHA) {
		t.Errorf("own origin ci/t9001: got %q, want the candidate %s", own, rec.CandidateSHA[:12])
	}
	// The OTHER repository's origin carries nothing.
	foreign := candidateGit(t, other, "ls-remote", "origin", "refs/heads/ci/t9001")
	if foreign != "" {
		t.Errorf("foreign origin ci/t9001: got %q — the verb force-pushed into another repository's remote", foreign)
	}
}

// TestIntegrationCandidateRunSelection pins the run-selection contract
// (card t1478 M2 repair): with TWO active runs, a bare discovery refuses
// ambiguous — but the launcher-selected run (MOAI_KANBAN_ID, the
// autoLaneRunResolveFn precedent) and the explicit --run flag both name
// the run whose card record the tree guard verifies against, so the
// candidate records for the right card.
func TestIntegrationCandidateRunSelection(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	for _, run := range []string{"run-a", "run-b"} {
		if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES(?,'active','t','t')`, run); err != nil {
			t.Fatalf("place run %s: %v", run, err)
		}
	}
	placeCardRowIn(t, db, "run-a", "t9001", cardWT)
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}

	t.Run("launcher-selected run resolves", func(t *testing.T) {
		t.Setenv(config.EnvFactoryRunID, "run-a")
		runID, err := candidateRunSelection(root, "")
		if err != nil {
			t.Fatalf("candidateRunSelection with MOAI_KANBAN_ID=run-a: %v", err)
		}
		if runID != "run-a" {
			t.Errorf("run selection: got %q, want run-a", runID)
		}
		if err := candidateCallerTreeGuard(root, "t9001", cardWT, runID); err != nil {
			t.Errorf("tree guard under the selected run: %v", err)
		}
	})
	t.Run("explicit --run flag resolves", func(t *testing.T) {
		runID, err := candidateRunSelection(root, "run-a")
		if err != nil {
			t.Fatalf("candidateRunSelection(--run run-a): %v", err)
		}
		if runID != "run-a" {
			t.Errorf("run selection: got %q, want run-a", runID)
		}
	})
	t.Run("no selection with two active runs refuses ambiguous", func(t *testing.T) {
		t.Setenv(config.EnvFactoryRunID, "")
		if _, err := candidateRunSelection(root, ""); err == nil {
			t.Fatal("want the ambiguous refusal when neither flag nor env names a run")
		}
	})
}

// placeCardRow inserts one card row with the given run and worktree (a
// fixture placement beside fcPlace, which lives in factory_card_test.go
// and pins its own column set).
func placeCardRow(t *testing.T, db *homestate.FactoryDB, runID, cardID, worktree string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage,lease_holder,lease_expires_at,heartbeat_at,decision_gate,decision_question,decision_resume,merge_sha,worktree_path,evidence_sha,hint_after) VALUES(?,?,'lane-test','in-progress',1,'','2026-10-09T00:00:00Z','','','','','','','','',?,'','')`,
		runID, cardID, worktree); err != nil {
		t.Fatalf("place %s: %v", cardID, err)
	}
}

// placeCardRowIn is the run-pinned alias the two-active-run fixture uses.
func placeCardRowIn(t *testing.T, db *homestate.FactoryDB, runID, cardID, worktree string) {
	t.Helper()
	placeCardRow(t, db, runID, cardID, worktree)
}

// sameStringSet compares two string slices ignoring order.
func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
		if seen[s] < 0 {
			return false
		}
	}
	return true
}
