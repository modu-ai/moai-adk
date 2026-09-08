# t540 — absorb-merge verification (integration window, lane-1, 2026-09-08)

## Claim

`develop` was absorbed into `WT-codex-path-escape`; the one conflict was resolved preserving both
sides; the merged tree builds, vets, and passes the card-affected tests.

## Evidence

**Window and tip, re-measured inside the window** (not carried from the dispatch):

```
$ git rev-parse --short refs/heads/develop   → 3ac58b5a1
$ git rev-parse --short origin/develop       → 3ac58b5a1
$ git rev-list --count --left-right origin/develop...HEAD
221     10
```

The dispatch stated develop was 57 commits ahead. Measured here it is **221**. The action is
unchanged; the figure reported onward is the measured one.

**Conflict — one file, one hunk:**

```
$ git merge --no-edit develop
Auto-merging CHANGELOG.md
CONFLICT (content): Merge conflict in CHANGELOG.md
```

Resolved by preserving both sides: 5 lines from ours (t540's own `### Fixed` entry) followed by
14 lines from theirs. Neither side was treated as stale — they are independent cards.

**Post-resolution token census, each side against controls** (`/usr/bin/grep -c`, absolute path —
the shell `grep` here is a ugrep wrapper that skips silently):

| token | merged | ours `0caa6097e` | theirs `3ac58b5a1` |
|---|---|---|---|
| `SPEC-CODEX-SKILL-PATH-SLASH-001` (ours) | 1 | 1 | 0 |
| `SPEC-TODO-HOME-TEMP-GUARD-001` | 1 | 0 | 1 |
| `SPEC-AC-COLLECTOR-ANCHOR-001` | 1 | 0 | 1 |
| `SPEC-DOCTOR-STAT-SEAM-001` | 1 | 0 | 1 |
| `SPEC-SEAM-GREENFIELD-001` | 2 | 0 | 2 |
| `SPEC-STATE-ANCHOR-VALIDATE-001` | 1 | 0 | 1 |
| `SPEC-WEB-WRITE-SAFETY-001` | 2 | 0 | 2 |
| **control** `SPEC-STATUS-DRYRUN-001` | 2 | 2 | 2 |
| **control** `SPEC-CODEX-ENABLED-FATAL-001` | 2 | 2 | 2 |
| **control** `^### Fixed$` | 9 | 9 | 9 |

Every side-specific token survives at its own side's count; the three controls are identical across
all three blobs, so the census is not reading a file where everything happens to be equal. Line
arithmetic agrees exactly: merged 1340 = theirs 1335 + ours' 5-line block.

**Merged-tree re-measurement** (merge commit `5bf686c76`):

```
$ go build ./...                  → rc 0
$ go vet ./internal/cli/...       → rc 0
$ go test ./internal/cli -run 'Codex|CodexConfig|SkillsDisable|Slash|Separator' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  123.908s
$ ... -v | grep -cE '^=== RUN +Test'   → 960
```

The 960 count is recorded because a selector that matches nothing also prints `ok`. It matched.

## Baseline-attribution

Measured in this run, in `.claude/worktrees/t540` on `WT-codex-path-escape`, in the merged tree at
`5bf686c76`, inside the integration window held by lane-1. Parent blobs read with `git show` at the
two SHAs named above.

## Gaps

- Verification was scoped to `internal/cli` — the only package this branch changes
  (`codex_config_path.go`, `codex_skills_disable.go` and their tests). The full suite was not run
  locally by design; CI on the develop push is the full-suite verdict.
- Windows was not exercised. AC-CSPS-001 stays open and unmeasured (see `progress.md` § Gate
  re-scoping).
- The token census covers `CHANGELOG.md`, the only conflicted file. Files auto-merged without
  conflict were not censused.

## Residual-risk

An auto-merged file could carry a semantic clash that produced no conflict and no compile error —
the census would not see it, and a package-scoped test run would only catch it inside
`internal/cli`. The develop-push CI is what covers that.
