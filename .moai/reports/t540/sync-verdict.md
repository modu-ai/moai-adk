# t540 — sync-phase verdict (SPEC-CODEX-SKILL-PATH-SLASH-001)

Card t540 · branch `WT-codex-path-escape` · measured at `e5df637bc` · lane-1.

Terminal state: **`implemented`** — lead ruling of 2026-09-08. NOT `completed`.

## Claim

1. Sync-phase artifacts are complete for the delivered scope (M1 seam + M2 publisher):
   CHANGELOG entry emitted, `progress.md` §E.4 populated, `spec.md` frontmatter advanced
   `in-progress → implemented`.
2. No docs-site surface exists for this change; none was invented.
3. AC-CSPS-001 remains **OPEN and unmeasured**. It is never recorded as satisfied, waived,
   or not-applicable.
4. M0 is measurable in principle on a Windows CI runner; two of its three axes are now
   measured and the third is not.
5. A process defect is recorded below (t562 absorption). It is recorded as a defect only —
   it is NOT part of the ruling's justification.

## Evidence

Re-measured this run, in this worktree, at `HEAD = e5df637bc`:

```
go test ./internal/cli/ -run 'TestConfigPath|TestUpsertCodexSkillDisable' -timeout 1800s
  → ok  github.com/modu-ai/moai-adk/internal/cli  0.882s   (all named tests PASS)
go vet ./internal/cli/...                → rc=0
go build ./...                           → rc=0
GOOS=windows GOARCH=amd64 go build ./... → rc=0
golangci-lint run ./internal/cli/...     → rc=0, "0 issues."
```

Exit codes were read directly, not through a pipe: an earlier reading in this session took
`$?` after `| tail`, which reports `tail`'s status rather than the command's. The figures
above are from the re-measurement.

Landing state:

```
git merge-base --is-ancestor HEAD origin/develop   → NOT-LANDED
git rev-parse --short origin/develop               → 91d25bc61
```

## Baseline-attribution

Every figure above was produced in this run, in
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t540`, at `HEAD = e5df637bc` on branch
`WT-codex-path-escape`. Nothing is carried over from §E.2's run-phase table.

## Defect — the M0 gate was bypassed by absorption (recorded, not acted on)

`plan.md §H` states "M0 is the gate; nothing lands before it". M0 was never run. Yet:

```
git branch -a --contains HEAD
  * WT-codex-path-escape
  + WT-codex-read-inverse        ← t562
```

t562 (`SPEC-CODEX-SKILL-PATH-READBACK-001`) has absorbed this branch's M1+M2 commits. If
t562 lands on `develop`, M1+M2 land with it — without M0 having been measured. The gate the
SPEC wrote for itself is, in practice, already routed around.

This is recorded as a defect and is deliberately **excluded from the ruling's rationale**.
Using "the gate has already been bypassed" to justify relaxing the gate would derive the rule
from its violation. The ruling rests instead on `acceptance.md:10` (`Blocking: YES — nothing
lands before it`) being unmeasured, which is a fact about the gate rather than about the breach.

Disposition: reported to the lead. Not reverted by this lane — t562's merge window is a
separate decision, held for separate reasons.

## M0 feasibility — measured on request, scope-limited

The lead's correction stands and is recorded: AC-CSPS-001 is **not** structurally
unmeasurable. Its own Given clause admits "a Windows host (physical, VM, or a Windows CI
runner)", and its Gap paragraph scopes the limitation to this worktree ("not satisfiable
**here** … must be executed **elsewhere**"). The earlier lane phrasing that generalized this
to the whole gate was wrong and is withdrawn.

| Axis | State | Evidence |
|---|---|---|
| Windows codex-cli distribution | **exists** | `npm view @openai/codex optionalDependencies --json` → `@openai/codex-win32-x64`, `@openai/codex-win32-arm64`, both `0.153.4`; local `codex --version` → `codex-cli 0.153.4` (same version) |
| Windows CI runner in this repository | **exists** | `ci.yml:376`, `release-pr-multi-os.yml:91`, `test-install.yml:100` matrices; `test-install.yml:146`/`:196` `runs-on: windows-latest` |
| codex-cli installs and runs on `windows-latest` | **UNVERIFIED** | not attempted |
| `codex debug prompt-input` runs there without interactive auth | **UNVERIFIED** | not attempted |

Two measured axes are not a gate half-closed. This report does not claim M0 is achievable,
feasible overall, or ready to run.

Line citations above are moving coordinates and are pinned to `e5df637bc`; re-measure before
relying on them.

## Gaps

- AC-CSPS-001: not attempted, not simulated. The gate is open.
- AC-CSPS-004 arms A / A' / C and AC-CSPS-005: delegated to t562, not t540 debt. Their
  closure was not verified here — `SPEC-CODEX-SKILL-PATH-READBACK-001` does not exist in this
  tree, so t562's stated scope was read from the dispatch rather than from its SPEC.
- Per-AC PASS statuses were re-measured in aggregate (one scoped selector, `ok`), not as eight
  independent per-criterion selectors.
- The unconditional-`ReplaceAll` mutant was not re-injected this run; that discrimination
  claim is carried from §E.2.
- `sync_commit_sha` holds `pending-backfill-t540`. A commit cannot cite its own hash; the
  backfill is owed and nothing schedules it.

## Residual-risk

- `implemented` with an open blocking gate is a legitimate but unusual state; drift detection
  treats a non-`completed` V3R6 SPEC as live, so an audit will surface this SPEC carrying a
  sync section with a placeholder SHA. That is the honest state, not a defect.
- The CHANGELOG leads with the user-visible Windows consequence and carries the open gate as a
  sub-bullet. A reader skimming only the lead sentence could take away "Windows is fixed" when
  no Windows runtime has ever executed this code.
- Two sessions in this card (the lane and manager-docs) each had their shell CWD silently
  land in a different tree than intended. Both were caught by a file-not-found; a command that
  happens to succeed in the wrong tree would have produced a plausible wrong measurement
  instead. Every figure in this report was taken with an absolute path or `git -C`.
