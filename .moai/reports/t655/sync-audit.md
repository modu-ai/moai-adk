# Sync-Phase Audit Report: SPEC-WORKTREE-KEY-WIRING-001 (card t655)

Auditor: sync-auditor (independent, lane-4 worktree, branch `WT-worktree-keys-wiring`, base `1d150a27d`)
Audit date: 2026-09-12
Audited tree: HEAD `195489beb` (8 commits ahead of base, clean)
Cross-model opinion: not invoked — `.moai/config/` carries no `audit_model` key (measured: `grep -rn 'audit_model' .moai/config/sections/` → 0 hits), so no cross-backend fan-out is warranted by config.

## Overall Verdict: PASS

| Dimension | Score | Verdict | Weight |
|-----------|-------|---------|--------|
| Functionality | 10/10 | PASS (must-pass) | 40% |
| Security | 9/10 | PASS (must-pass) | 25% |
| Craft | 9/10 | PASS | 20% |
| Consistency | 10/10 | PASS | 15% |

Harmonic mean = 4 / (1/10 + 1/9 + 1/9 + 1/10) = **9.47**. Must-pass firewall (Functionality + Security) both pass independently. No blocking findings; four optional findings recorded below.

## Verification Evidence (all commands run by this auditor, this run, this tree, HEAD `195489beb`)

**E1. AC re-verification — scoped selector sweep (not a trust of §E.2).**
`go test ./internal/cli/ -run 'TestAutoMerge|TestWorktreeAdvisoryTruthful' -count=1 -v` → **16 test functions, 30 subtests, all PASS**, `ok github.com/modu-ai/moai-adk/internal/cli 3.461s`. This re-executes every cli-side AC judge (AC-WKW-001..010, 013, 014) plus AC-WKW-011, with no empty sweep (swept count counted from the log: 16 top-level `=== RUN` test lines).
Vacuity review of the assertions (read, not assumed):
- AC-WKW-002 `TestAutoMergeHappyPath` — real-git fixture; asserts develop HEAD is a two-parent merge commit (`rev-list --parents -n 1 HEAD` → 3 fields) whose second parent equals the session branch tip, window released, notice names both branches + the merge commit. Cannot pass vacuously.
- AC-WKW-006 `TestAutoMergeZeroPush` — seam-log enumeration: exactly one `merge` carrying `--no-ff` at the develop worktree; banned-substring sweep (push/fetch/pull/remote/origin) over the full log; acquire→release ceremony order; released at end.
- AC-WKW-007 `TestAutoMergeConflict` — two parts: seam abort path (abort once at target, window released, conflict notice) AND a real-git conflicted merge asserting `rev-parse -q --verify MERGE_HEAD` fails after the abort and the session worktree file is byte-identical.
- AC-WKW-013 `TestAutoMergeNoticePrefixDistinct` — constant-level distinctness against BOTH removal prefixes, per-line prefix assertion over three failure paths' output.
- AC-WKW-014 `TestAutoMergeToggleIndependence` — 4-combination truth table; merge-before-remove ordering asserted in the both-on case.
- §E.2's observed-red records (dirty-check bug, EC-13 distinctness violation, by-value seam mutation) corroborate that these assertions have fired red on known inputs (verification-completeness §1.1).

**E2. AC-WKW-012 judges re-executed.**
`go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1` → `ok github.com/modu-ai/moai-adk/internal/config 1.767s` (sweep non-empty; the real umbrella test exists).
`grep -c 'AutoMerge has no production reader' internal/config/types.go` → `0` (pattern absent — `! grep -q` form exits 0, criterion met).

**E3. Regression (DoD #5).** `go test ./internal/cli/ -run 'TestSessionWorktree|TestPRMergeCleanup' -count=1` → `ok ... 2.707s`.

**E4. Cross-platform (DoD #2).** `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

**E5. Quality gates (DoD #3).** `gofmt -l` over the 11 touched Go files → empty. `go vet ./internal/cli/ ./internal/config/` → clean (exit 0, no output). `golangci-lint run ./internal/cli/... ./internal/config/...` → `0 issues.`

**E6. Template neutrality (§3 of the audit scope).**
`grep -rn 'develop_branch' internal/template/templates/` → exactly ONE hit, the workflow.yaml:61 comment naming the key (accurate, no value leaked). `auto_merge` in templates: the workflow.yaml comment naming the consumer + `auto_merge: false` at line 64 (distributed default intact). `git-strategy.yaml.tmpl` carries zero `develop_branch` (no develop-specific value in any shipped template). One pre-existing skill-module line also mentions `auto_merge` — see F3.

**E7. 3-phase close verification.**
- `spec.md` frontmatter `status: completed`, `updated: 2026-09-12` (read directly).
- Sync commit `313d7feb5` (`git show`): exactly 3 files — progress.md §E.4, spec.md frontmatter (2 lines: status only), CHANGELOG (1 line) — with `Authored-By-Agent: manager-docs` trailer and 🗿 MoAI trailer. Ownership matrix satisfied (`implemented → completed` owned by manager-docs).
- `sync_commit_sha: "313d7feb5"` backfilled by `195489beb` — the D3 placeholder→backfill chain is complete and honest (placeholder in the sync commit itself, per the schema's backfill exemption).
- CHANGELOG factual claims sampled and verified: (a) three session launchers wired — init.go:507-512, profile_setup.go:260-264, web.go:111-115, all merge-before-cleanup ✓; (b) advisory "emitted by init / update / update-template-sync / web" — `grep -l emitWorktreeAdvisory` → exactly those 4 call sites ✓; (c) "off by default" — template line 64 + defaults.go `AutoMerge: false` ✓; (d) "seam signatures accept no force parameter at all" — session_worktree_automerge.go:99-108 ✓; (e) "14 acceptance criteria" — acceptance.md carries AC-WKW-001..014 ✓.
- Docs-site no-change decision reproduces: `grep -rn 'auto_merge' docs-site/content/` → 0 hits; `auto-merge` hits only the unrelated `/moai sync` PR auto-merge pages — exactly as progress.md §E.4 records.

## Findings

All findings are non-blocking (verdict unaffected). Reported for coverage; the orchestrator treats them as discretionary per the finding-consumption brake.

- **F1** [minor] [optional] internal/cli/session_worktree_automerge.go:316,332 — the config-supplied `develop_branch` and the session branch are interpolated into git argv without a `--` end-of-options separator. Today this is fail-safe: the session branch comes from `git rev-parse --abbrev-ref HEAD` (git-constrained), `develop` never reaches `git merge` as an argument (the merge target is the worktree's HEAD), and a `-`-leading rev expression like `--all..x` errors into the skip+notice path. Threat domain is the project's own config file (same trust domain the pre-existing manual integration path already uses). — Required fix (if adopted): insert `--` before rev arguments in `gitRevListCountReal` and `gitMergeNoFFReal`.
- **F2** [minor] [optional] internal/cli/session_worktree_automerge_test.go (AC-WKW-006) — the zero-push assertion enumerates the seam log; a hypothetical future `exec.Command("git", "push", ...)` written directly in the engine body (not via a seam) would evade it. Compensating controls today: the real implementations are four whitelisted invocations (code-review surface), and the real-git fixtures fail on any push (no remote configured). — Required fix (if adopted): a static source-scan test asserting `sessionExitAutoMerge` contains no direct `exec.Command` outside the seam definitions (house pattern: `TestNew_NoAskUserQuestion`).
- **F3** [info] [optional] internal/template/templates/.claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md:190 — the pre-existing line "`auto_merge` (bool, default: false) — automatically merge worktree branches" was FALSE before this card (the key was declared-but-unread) and became TRUE the moment this card landed. It is not in REQ-WKW-012's enumerated artifact list, and no edit is required post-state — the claim is now accurate, including "off by default". Recorded because the honesty sweep's completeness here is coincidental (the feature landing repaired the doc), not maintained (no test or artifact pins it). No action required for this card.
- **F4** [info] [optional] internal/cli/session_worktree_automerge.go:341-343 — `gitMergeInProgressReal` conflates "no merge in progress" with "git errored" (both read as not-in-progress), documented in-code as best-effort; the abort-also-fails branch emits a distinct manual-resolution notice, so the unhandled state is loud. Acceptable as designed.

## Gaps (explicitly NOT observed by this audit)

- The full `go test ./internal/cli/` package suite was NOT re-run (slot discipline — a full-package run requires the heavy-test lease; CI on origin/develop owns the package verdict). The §E.2 run-2 figure (`ok ... 2392.715s`, 0 FAIL) is attributed to the run record, not re-measured here. My evidence covers every SPEC-relevant selector plus the DoD #5 regression selectors.
- Coverage percentages (§E.2: engine functions 100%, `gitRevListCountReal` 85.7%) were not re-measured; the scoped `-coverprofile` record in progress.md is the attributed source.
- The run-1 flake record (`TestHookWrapper_LargeStdin_DoesNotExceedTimeout`) was not independently reproduced; the isolation-control claim (158.9ms green) is attributed to §E.2.

## Residual Risk

- A process death between acquire and release leaves a stale integration-window hold — the identical failure mode the manual ceremony has (accepted at plan-audit; human `--force` recovery, never machine displacement). The skip-on-stale rule in REQ-WKW-004 keeps auto-merge from ever worsening it.
- The auto-merge is git-flow-manual-mode-gated and reads the project's own `develop_branch`; a misconfigured value fails safe to skip+notice (verified AC-WKW-004), but a *valid* value naming an unintended branch would be honored — inherent to reusing the configured integration target (the SPEC's binding design decision #2).
- F2's seam-log blind spot (see above) means the zero-push guarantee rests on code inspection + fixtures for any future engine-body git call added without a seam.

## Recommendation

**PASS.** All 14 acceptance criteria re-verified green by this auditor's own runs; the security-critical invariants (never displace a hold, never push, conflict leaves no MERGE_HEAD, OFF = byte-identical) hold mechanically and are asserted by non-vacuous tests including real-git fixtures; the honesty artifacts agree with each other and with the tree; the 3-phase close is complete and correctly owned. The four optional findings are hardening opportunities for follow-up cards, not defects in this one.
