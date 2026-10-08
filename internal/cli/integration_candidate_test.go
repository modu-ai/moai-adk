package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
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
// stdout, failing the test on error.
func candidateGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
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
