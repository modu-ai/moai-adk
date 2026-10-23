// integration_backup.go — `moai integration backup`
// (SPEC-TRUNK-BACKUP-SYNC-001 REQ-TBS-001..004): push local main to the
// remote ref no workflow reacts to, so landed work exists on a second ref at
// origin between release batches.
//
// The verb is deliberately minimal — no flags, no config keys, no scheduler
// (spec §E: any proposal that grows it re-opens the simplicity ladder). It
// copies the candidate verb's runner shape (scrubbed env, GitExitError,
// force-push of an explicit refspec; sibling exceptions, not a shared
// mechanism — SPEC-CANDIDATE-CI-001 REQ-CCI-016 is the precedent) and differs
// in the one way that matters: its push is the pinned refspec
// `<sha>:refs/heads/sync/main` — flags, remote, and target ref are source
// constants, and the transmitted commit is the printed one.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/spf13/cobra"
)

// @MX:NOTE: [AUTO] The backup push is this constant argv, nothing else.
// @MX:SPEC: SPEC-TRUNK-BACKUP-SYNC-001
// REQ-TBS-002: the target-side ref name, the force flag, --no-follow-tags,
// and --no-recurse-submodules are source constants — never computed from
// flags, configuration, or the environment. --force is scoped by
// construction to exactly this refspec; --no-follow-tags holds even under a
// caller's push.followTags=true, so no annotated tag reachable from main
// rides the push (a tag ref would update a second remote ref and can fire
// release.yml); --no-recurse-submodules holds even under a caller's
// push.recurseSubmodules — under `only` the argv without it exits 0 as a
// silent no-op while the remote gains no refs/heads/sync/main, and under
// `on-demand` the push reaches submodule remotes, a second remote surface
// (sync-audit round 1, S1). The refspec's source side is the PINNED SHA
// resolved once by the rev-parse below (sync-audit round 2, R2-F1/R2-F2):
// <sha>:refs/heads/sync/main transmits exactly the commit the success line
// prints — identical under any remote.origin.pushurl divergence, and no
// post-push remote read exists to rename the report. Provisional while
// operator decisions Q1-Q3 are open: a verdict selecting another surface,
// ref name, or force semantics is a one-commit amendment of these
// constants, the tests, and the runbook together.
const (
	backupTargetRef = "refs/heads/sync/main"
	backupSourceRef = "refs/heads/main"
	backupRemote    = "origin"
)

// integrationBackupSeams carries the verb's test seams. nil Git is the real
// scrubbed runner.
type integrationBackupSeams struct {
	Git func(dir string, args ...string) (string, error)
}

func (s *integrationBackupSeams) gitRunner() func(dir string, args ...string) (string, error) {
	if s.Git != nil {
		return s.Git
	}
	// candidateScrubbedGit (card t1478 M2 repair): repository-location
	// variables are removed before every exec, so an inherited GIT_DIR /
	// GIT_WORK_TREE cannot redirect the push to a foreign repository's
	// origin; a non-zero exit is a *factorylane.GitExitError carrying both
	// streams. The verb runs with the caller's working directory, so any
	// worktree of the repository resolves the same shared refs.
	return candidateScrubbedGit
}

// runIntegrationBackup pushes refs/heads/main to refs/heads/sync/main on
// origin. Order is deliberate (REQ-TBS-003): the source ref must resolve
// BEFORE any push attempt — a refusal names the missing ref and the constant
// target and exits non-zero having issued no push. On a push failure
// (REQ-TBS-004) both captured streams are printed and no further git command
// runs. On success the pinned SHA — the commit the refspec transmitted —
// and the target ref are printed.
func runIntegrationBackup(dir string, seams integrationBackupSeams, out io.Writer) error {
	git := seams.gitRunner()

	sha, err := git(dir, "rev-parse", "--verify", backupSourceRef)
	if err != nil {
		return fmt.Errorf("integration backup: %s does not resolve in %s — nothing to back up; the verb refuses before any push attempt (the constant target is %s on %s)", backupSourceRef, dir, backupTargetRef, backupRemote)
	}
	sha = strings.TrimSpace(sha)

	// R2-F1/R2-F2 (sync-audit round 2): the refspec PINS the resolved SHA —
	// <sha>:refs/heads/sync/main — so the transmitted commit IS the printed
	// one: identical under any remote.origin.pushurl divergence from the
	// fetch url (R2-F1), and no post-push remote read exists to observe a
	// competitor (R2-F2; the round-1 S3 ls-remote lookup and its
	// silent-downgrade path — R2-F3 — are deleted with it). If main advances
	// between this rev-parse and the push, this backup carries the pinned
	// commit; accuracy of the report is the requirement (REQ-TBS-001), and
	// the next run catches up the rest.
	if _, err := git(dir, "push", "--force", "--no-follow-tags", "--no-recurse-submodules", backupRemote, sha+":"+backupTargetRef); err != nil {
		var exitErr *factorylane.GitExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("integration backup: push to %s failed (git exit %d) — no further git command was issued\npush stdout:\n%s\npush stderr:\n%s", backupTargetRef, exitErr.ExitCode, exitErr.Stdout, exitErr.Stderr)
		}
		return fmt.Errorf("integration backup: push to %s failed: %v — no further git command was issued", backupTargetRef, err)
	}

	_, _ = fmt.Fprintf(out, "integration backup: pushed %s to %s on %s\n", sha, backupTargetRef, backupRemote)
	return nil
}

func newIntegrationBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Push local main to the non-reacting sync/main ref at origin (a redundancy copy between release batches)",
		Long: `Push local main to the non-reacting sync/main ref at origin.

Between release batches, local main is the only copy of landed work. This
verb mirrors it onto refs/heads/sync/main — a ref no workflow under
.github/workflows reacts to (the exclusion is regression-guarded by the Go
test suite) — so a second copy exists at origin. The push is the constant
argv ` + "`push --force --no-follow-tags --no-recurse-submodules origin <sha>:refs/heads/sync/main`" + `, where the
refspec pins the local main SHA resolved at invocation start — the printed
SHA is exactly the commit transmitted. Force is scoped by construction to
exactly that refspec, no tag rides the push, and no submodule remote is
reached.

sync/main is a redundancy copy only: it never lands, is never merged into
any branch, and never substitutes for the release/main-batch-YYYYMMDD
pull-request path. See .moai/docs/trunk-backup-runbook.md.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("integration backup: resolve the working directory: %w", err)
			}
			return runIntegrationBackup(cwd, integrationBackupSeams{}, cmd.OutOrStdout())
		},
	}
	return cmd
}
