# t574 — verdict (lane-3, pre-integration)

Card: t574 · Branch: `WT-temp-roots-ac` · Worktree: `.claude/worktrees/t574` · Toolchain: `go1.26.8 darwin/arm64`
Base: merge `f295fe698` (HEAD^1 = `d060e0d13` origin/develop, HEAD^2 = `a4461479e` local develop);
AC baseline `git merge-base develop HEAD` = `a4461479e`.

## Claim

SPEC-TODO-HOME-TEMP-GUARD-001 had no acceptance criterion requiring the production temp-root set
(`defaultTempRoots()`, `internal/kanban/temp_origin.go:60-62`) to contain `/tmp` and `/var/folders`;
the gap reproduced and is closed by an in-place amendment 0.1.5 that adds a positive-direction
clause to AC-THG-006. AC count stays 8, §D.0 mapping unchanged, no code change. SPEC is back to
`completed`.

Decision trail: operator first chose "new AC"; the lead ruled "clause to AC-THG-006" and then deferred
to the operator; the operator re-confirmed "clause to AC-THG-006". Status handling follows the schema's
`completed → in-progress (amendment)` row (manager-spec) and the sync re-close (manager-docs).

## Evidence (commit chain, all `card t574`)

| Commit | Owner | What |
|---|---|---|
| `7da398808` | lane | reproduction mutant prediction pinned before injection |
| `95ba9deb2` | lane | reproduction: `repro-summary.md`, `mutant-kanban.txt` (8 AC-named kanban tests PASS, only `TestDefaultTempRoots_Membership` FAIL), `mutant-cli.txt` (3 PASS) |
| `ac751bc48` | manager-spec | amendment: spec.md status `in-progress`, version 0.1.5, `amendment_of`, `## Amendments` (prior_completed_sha `029ab039f`, verified `-status: in-progress` / `+status: completed`); AC-THG-006 membership clause + M-574 RED control; progress.md block under §E.1 |
| `559fd8de6` | lane | plan-audit verdict exported: `plan-audit.md` PASS 0.94, 0 must-pass failures, D1-D7 optional |
| `d937cd68d` | manager-spec | D1 (claim narrowed to tests that ran), D3 (matrix row per clause), D4 (expected drift note), D5 (Linux TMPDIR qualifier), D6 (anchored selector, RUN=2) |
| `0f30477af` | manager-develop | run-phase mutant prediction pinned (ancestor of HEAD: `merge-base --is-ancestor` exit 0) |
| `4eff56551` | manager-develop | run re-measure (D2 closed): `run-baseline.txt` 2 RUN / 2 PASS exit 0; `run-mutant-source.diff` + `run-mutant.txt` Membership FAIL, ComponentBoundary PASS, exit 1; `run-mutant-revert.txt` empty diff; `run-kanban-pkg.txt` `ok internal/kanban 158.9s` exit 0 |
| `cdee109cd` | manager-docs | sync re-close: spec.md `status: completed`, progress.md block after §E.4 |

Lane re-verification on `cdee109cd` (this run):
- `git diff --stat d937cd68d HEAD -- internal cmd pkg` → empty (no code change).
- spec.md frontmatter: `version: "0.1.5"`, `status: completed`, `updated: 2026-09-10`, `amendment_of: SPEC-TODO-HOME-TEMP-GUARD-001`.
- era.go tokens `sync_commit_sha|mx_commit_sha` in progress.md: 3 at `HEAD~1`, 3 at HEAD.
- Tree-built binary (`go build ./cmd/moai`, exit 0): `spec audit --json --filter-spec SPEC-TODO-HOME-TEMP-GUARD-001`
  → `"drift_findings": []`, `"modern_era_clean": 1` (`sync-audit-json.txt`); `spec lint` → `✓ No findings`, exit 0
  (`sync-lint.txt`).
- CHANGELOG.md:412 (existing 0.1.4 close entry) says "8 acceptance criteria" — still true; no
  AC-THG-006 wording claim to invalidate. No CHANGELOG/README/docs-site edit (no user-visible change).

## Baseline-attribution

Worktree `.claude/worktrees/t574`, go1.26.8, measurements in this run against the commits named above;
lint/audit by a binary built from this tree and invoked by path (not the installed `moai`).

## Gaps

- Linux / Windows cells not measured (on Linux with TMPDIR unset the mutant drops `/var/folders`
  instead of `/tmp`) — CI.
- No sync-auditor pass: the change is acceptance-criteria text only; plan-audit covered the text and
  run re-measure covered the judging test. Lead decides whether a sync audit is required.
- The M-574 run-phase control exercised only the two judging tests; the wider AC-named set under the
  mutant is the reproduction-phase evidence (`mutant-kanban.txt`), not re-run.
- `internal/cli` not run in the run phase (no code change; reproduction-phase `mutant-cli.txt` only).

## Residual-risk

- Membership clause pins count + literals + presence of `os.TempDir()`; a mutant that keeps 3 members
  but changes the `os.TempDir()` slot is not caught by definition (already recorded in the SPEC).
- AC-THG-006's boundary half still reads "(M1 RED 관측 예정)" although `.moai/reports/t536/m1-mutant-red.txt:22`
  recorded that RED — stale wording outside this card's clause scope, candidate follow-up.
- plan-audit D7: `Authored-By-Agent:` trailers on agent commits sit above the mandatory final `🗿 MoAI`
  line, so `git interpret-trailers` may not parse them; ownership attribution may report Unmeasured.
- Out-of-scope observation (plan-audit): progress.md §E.4 `sync_commit_sha` line carries an end-of-line
  comment the parser reads as part of the value (pre-existing since `586f26f2a`).
- This verdict commit lands after the sync re-close commit; it touches only `.moai/reports/t574/`, not
  SPEC artifacts.

## Integration window re-measure (lead-assigned)

Claim: the card absorbed onto local develop `987eb7e40` (which now carries t657, merged outside dispatch)
still passes its AC-named tests and SPEC checks; no absorb delta reaches kanban except
`internal/core/git`, which the whole-package run covers.

Evidence:
- `window-acquire.txt`: window acquired by lane-3 on `WT-temp-roots-ac`.
- `absorb-merge.txt`: `git merge --no-ff 987eb7e40` exit 0, no conflict. Absorb commit
  `2d2e7d8170827ae71879272049be8cfa71920cd2`, parents `cf49681de…` (card tip) + `987eb7e40…` (develop).
- Delta judgment (`absorb-changed-files.txt` 275 files vs `absorb-kanban-deps.txt`, the 12 module-internal
  packages of `go list -deps -test ./internal/kanban/`, plus go.mod/go.sum and this SPEC dir):
  `absorb-delta-hits.txt` = `internal/core/git/manager.go`, `internal/core/git/status_optional_locks_test.go`.
  `internal/web` changed 18 files but is not a kanban dependency (`web-in-deps=0`). No `go:embed` in
  `internal/kanban` (grep exit 1). Matcher control: a planted `internal/kanban/temp_origin.go` entry is
  matched (count 1).
- `absorb-kanban-ac-tests.txt`: 9 AC-named kanban tests (incl. `TestDefaultTempRoots_Membership`) —
  9 RUN / 9 PASS / 0 FAIL-SKIP, exit 0.
- `absorb-cli-ac-tests.txt`: 3 `TestTempOriginGuidance_*` PASS, exit 0.
- `absorb-kanban-pkg.txt`: `go test -count=1 ./internal/kanban/` → `ok … 197.3s`, exit 0 (covers the
  changed `internal/core/git` dependency).
- Binary rebuilt from the absorbed tree: `absorb-spec-lint.txt` `✓ No findings` exit 0;
  `absorb-spec-audit.txt` `"drift_findings": []`, `"modern_era_clean": 1`.

Baseline-attribution: absorb commit `2d2e7d817` tree, go1.26.8, this run.

Gaps: `internal/core/git` own package tests not run here (owned by the card that changed it); CI matrix not yet run.
