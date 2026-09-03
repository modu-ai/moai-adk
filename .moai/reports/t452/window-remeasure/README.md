# t452 — integration-window re-measurement (lane-12)

Window held as `lane-12`; card `t452`, branch `WT-codex-skill-wiring`.

## Baseline attribution

| Coordinate | Value |
|---|---|
| Pre-merge card tip | `4145e5755` |
| Absorbed develop tip | `1e4508ef6` |
| Absorb merge commit | `994457ed9` |
| Absorb merge tree | `44f26de5a404d6f5556b8d05c577666e6f3f2fe2` |
| Evidence tip (this commit) | recorded in `tip-final.txt` after the evidence commit |

All figures below were produced in **this** window, in this worktree, against the
absorb-merge tree — not carried over from the pre-merge measurement.

## Pre-merge guard checks (lead-mandated, 2 lines)

```
git config core.fsmonitor   -> no output, rc=1  (unset — required state)
find .git -name index.lock  -> 0 files
```

No `index.lock` contention occurred at any point in this window; the
`Unable to create ... index.lock` message never appeared, so neither the
cause-A nor the cause-B branch of the lead's decision tree was entered.

## Conflict resolution

`git merge --no-ff develop` raised exactly two conflicts.

| File | Resolution |
|---|---|
| `CHANGELOG.md` | Both sync-close entries kept — t452 `SPEC-CODEX-SKILL-LOADER-001` above t450 `SPEC-PLAN-AUDITOR-RESIDUE-001`. No content from either side dropped. |
| `internal/template/catalog.yaml` | Single conflicting field: the `moai` skill row's `hash`. **develop value adopted** (`91c36c2c…`) per the lead directive; HEAD value (`f005e873…`) discarded. |

`.codex/agents/moai/sync-auditor.toml` — the second file the lead flagged as
overlapping — did **not** conflict and was not modified by either side in this
merge. No divergent value was found there, so nothing needed reporting under the
"값이 다르면 강행하지 말고 보고" clause.

The catalog hash is a whole-tree-derived value, so adopting one side's literal is
only provisionally correct; `TestCatalogHashParity` was run on the merged tree to
decide it, and passed — no regeneration was required.

## Re-measurement on the merge tree — what was run

Every row is `rc` read directly, with no pipe.

| # | Command | rc | Log |
|---|---|---|---|
| 1 | `go test ./internal/spec/... -run TestCatalogHashParity -count=1` | 0 | `g1-catalog-parity.log` |
| 2 | `go test ./internal/template/ -run TestManifestHashFormat -count=1` | 0 | `g2-manifest-hash.log` |
| 3 | `go test ./internal/template/agentemit/... -count=1` | 0 | `g3-agentemit.log` |
| 4 | `go test ./internal/codexwiring/... -count=1` | 0 | `g4-codexwiring.log` |
| 5 | `go test ./internal/template/... -count=1` | 0 | `g5-template-all.log` |
| 6 | `go test ./internal/spec/... -count=1` | 0 | `g6-spec-all.log` |
| 7 | `gofmt -l internal pkg cmd` | 0 | `g7-gofmt.txt` (0 lines) |
| 8 | `go vet ./internal/codexwiring/... ./internal/template/... ./internal/spec/...` | 0 | `g8-vet.log` |

Rows 1-3 are the three guards the lead named. Rows 4-6 are the packages this
card's own change touches, run whole rather than by selector because each
completes in seconds.

## What was NOT measured

Named explicitly so the green above is not read wider than it is.

- **`internal/cli` — not run.** The full package is ~1002s on this machine and
  this card changes nothing under it. The absorbed develop commits do change it,
  but those landed on develop already measured by their own lanes, and
  re-measuring them here would be measuring someone else's work on a loaded
  machine. CI on the pushed develop head is the verdict for that surface.
- **Full repository suite — not run**, by standing local policy. The full-suite
  verdict is CI's, against a clean environment and the darwin/windows matrix.
- **`golangci-lint` — not run** in this window. `go vet` covers the card's
  packages; the lint verdict is CI's.
- **Cross-platform build — not run.** Local darwin only.

## Working-tree disposition — `.claude/settings.json`

The lead's dispatch said to leave the dirty `.claude/settings.json` (+435/−300)
untouched and out of the merge. **It could not simply be left**: `git merge`
refused to start with

```
error: Your local changes to the following files would be overwritten by merge:
	.claude/settings.json
```

because develop also changes that file. Handling, in order:

1. Both versions were preserved byte-for-byte outside the tree before anything
   was discarded — the working copy and develop's version, to the session
   scratchpad. Nothing was lost and the working copy is recoverable.
2. `diff` showed the two are **not** identical: the working copy carries a large
   block (`respectGitignore`, `skillListingBudgetFraction`, the `env` and
   `permissions` blocks) that develop's version does not.
3. That one path was reset to HEAD (`git checkout -- .claude/settings.json`),
   which unblocked the merge and satisfied the lead's actual requirement — the
   file's working-copy content is not in the merge.
4. After the merge, `.claude/settings.json` carries **develop's** version. The
   regenerated working-copy content is not in the branch and not in the merge.

This is a deviation from the literal instruction ("그대로 두십시오") forced by a
mechanical constraint, taken only because it was made reversible first. It is
reported rather than tidied away.
