# t313 — Stale-Card Verdict (2026-09-02, lane-7)

## Claim

Card t313 ("EnterWorktree 분기 기준을 develop 으로 고정 + moai web 기준 브랜치 선택 표면 추가")
requires no implementation: its dedicated SPEC (SPEC-WORKTREE-BASEREF-001, Tier M) is already
completed through the 3-phase close and landed on origin/develop. The queue card is stale.

## Evidence

All commands run 2026-09-02 in worktree `.claude/worktrees/t313`, branch `WT-baseline-branch`
(based on local develop `b7462203a`).

**SPEC completion state** (read from `.moai/specs/SPEC-WORKTREE-BASEREF-001/`):

- `spec.md` frontmatter: `status: completed`, 16 GEARS REQ / 16 AC (Tier M ceiling).
- `progress.md` §E.1: plan-audit iter-2 **PASS-WITH-DEBT 0.92** (`.moai/reports/t313/plan-audit-iter2.md`).
- §E.2: run-phase **16/16 AC PASS** (`.moai/reports/t313/run-verdict.md`).
- §E.4: sync-audit **PASS 0.90** harmonic (`.moai/reports/t313/sync-audit.md`), 3-phase close.

**Landing measurement**:

| Check | Command | Result |
|---|---|---|
| Merge commit on local develop | `git merge-base --is-ancestor 62ff3c2e6 HEAD` | rc=0 |
| Merge commit on remote develop | `git merge-base --is-ancestor 62ff3c2e6 origin/develop` | rc=0 |
| Recorded run SHA is ancestor | `git merge-base --is-ancestor 8c46460ff HEAD` | rc=1 |
| Recorded sync SHA is ancestor | `git merge-base --is-ancestor de2042416 HEAD` | rc=1 |

The merge commit `62ff3c2e6 merge(WT-worktree-baseref): integrate card t313` carries the full
commit chain on develop: `a9c61cf56` (config schema key) → `26cc9ba90` (hook SessionStart
origin/HEAD alignment) → `97aef573d` (git worktree add base operand) → `c59e74232` (doctor
diagnostic) → `80d9e7e5b` (web console free-text field) → `7d46e69c9` (docs) → run/sync evidence
(`92102de1e`, `b0d179de1`, `3fd8b5072`) → merge `62ff3c2e6` → post-merge `f5a834fef` (doctor
golden pass-count restamp, already on develop).

**⚠️ SHA drift**: the SHAs recorded in `progress.md` (`8c46460ff` run-final, `de2042416` sync)
are NOT ancestors of develop — the branch was rebased before the merge. Commit messages are
preserved verbatim (`7d46e69c9` carries the run-final message, `b0d179de1` the sync message).
Auditing by recorded SHA yields a false "not landed"; audit by commit-message grep instead.

**Independent re-measurement in this tree** (b7462203a basis):

| Check | Command | Result |
|---|---|---|
| Feature tests live | `go test ./internal/config -run 'GitStrategy' -count=1` | `ok ... 0.400s` |
| Local key configured | `grep worktree_base_branch .moai/config/sections/git-strategy.yaml` | `worktree_base_branch: develop` |
| Template neutral default | `grep worktree_base_branch internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | `worktree_base_branch: ""` (line 12) |

**Card's last unverified item, closed by observation (2026-09-02)** — the card mandated
measuring whether Claude Code's `fresh` worktree baseline actually reads
`refs/remotes/origin/HEAD`. Measured as a side effect of creating THIS worktree
(`EnterWorktree(t313)`) after the lead's `git remote set-head origin develop`:

- `git rev-parse HEAD` → `f7cabfc296` (new branch sits exactly on this commit)
- `git merge-base HEAD origin/develop` → `f7cabfc296` (= origin/develop tip — HEAD is its direct child)
- `git rev-parse origin/HEAD` → `f7cabfc296`
- `git merge-base HEAD origin/main` → `7ad9f8534` (different — rules out a main baseline)

Conclusion: `fresh` followed `refs/remotes/origin/HEAD` → the set-head fix is effective.
This is behavioral evidence for the gap SPEC-WORKTREE-BASEREF-001 inherited as G2 ("inferred
from behaviour, not read from Claude Code source"); the doctor item
(`internal/cli/doctor_worktree_base.go`, `Worktree Base Branch`) remains the standing mitigation.

## Baseline-attribution

Measured in this run, in this tree (worktree t313, branch `WT-baseline-branch` at develop
`b7462203a` + the one evidence commit below). The completion-state readings cite the SPEC
artifacts as they exist on develop; the re-measurements are this session's own command outputs.

## Gaps

- Claude Code `fresh`'s source code was not read (the observation is behavioral, matching how
  the SPEC itself framed G2).
- The web console POST → `applyGitStrategyKey` path was not driven through a browser (the
  SPEC's own audit gap 4 — unchanged by this verdict; unit round-trip coverage exists).
- Windows/Linux behavior of the alignment step not exercised locally (SPEC audit gap 2 — CI covers).

## Residual-risk

- `refs/remotes/origin/HEAD` is repository-global metadata an external actor can move; the
  SPEC accepted this by design (idempotent realignment + doctor surfacing).
- If the lead merges this evidence branch, the queue card t313 should be closed (done/drop at
  the lead's discretion — queue mutations are lead-owned, untouched by this lane).

## Disposition

Zero code commits from this lane on t313 — the correct outcome is "no work needed", not
"work skipped". Worktree t313 is disposable (no unique content beyond this evidence commit,
which lives on the branch ref and survives worktree removal).
