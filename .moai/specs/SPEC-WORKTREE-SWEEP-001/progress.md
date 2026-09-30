# progress.md — SPEC-WORKTREE-SWEEP-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-30
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
tier: M
cycle_type: tdd
note: Authored at plan-phase creation (card t1369). Doctrine budget constraint recorded — worktree-integration.md is already over the 40,000-char instruction-file budget; REQ-WS-014 and AC-WS-014 bound the addition (≤ +1,200 chars, no new headings).
worktree_integration_baseline_chars: 41194   # `wc -c` at run start, 2026-09-30, this tree — the instrument AC-WS-014 is judged by (re-measure at M5, never re-pin)

## §E.2 Run-phase Evidence

Cycle: tdd. Branch WT-worktree-sweep. Run commits: M1 `382a0d825` → M2 `e3428c98f` → M3 `7a0c1ee07` → M4 `aad231a08` → M5 `62197bb38` → M6 `30f36cbd5`. RED evidence captured before each milestone GREEN (verbatim outputs preserved in /tmp/t1369/red_m1.txt … red_m4.txt during the run; the deciding RED lines are transcribed below).

### Milestones

- **M1** — verdict record (`sweepVerdict`), flag surface (`--base` default `origin/develop`, `--yes`, `--json`), dry-run skeleton, cwd-probe platform split, root.go registration. RED: compile failure on undefined sweep symbols (`undefined: sweepProcessCWDs …` exit 1) at `878bf8ca3`.
- **M2** — landing predicate: fetch-failed → `cause=fetch-failed`, ancestry exit 1 → `cause=not-landed`, any other exit → `cause=landed-check-failed`. RED (at `382a0d825`): `sweep_test.go:467: reason must carry cause=fetch-failed, got "cause=not-evaluated"`; `landed must read "undetermined", got not-checked`. GREEN caught a real defect: the success path read as exit −1 (`merge-base --is-ancestor … exited -1`), fixed by treating a nil command error as exit 0.
- **M3** — predicate composition: dirty / ignored (shared decision + hoist-aware filter) / process-cwd. RED (at `e3428c98f`): five cells read `cause=not-evaluated` / `cause=not-landed` with predicates `not-checked`.
- **M4** — tier routing + apply path. RED (at `7a0c1ee07`): every `--yes` tree survived (`the disposable tree must be gone, stat error: <nil>`). GREEN caught two fixture defects (macOS `/var` vs `/private/var` protecting the fixture main checkout; git collapsing a fully-ignored `.moai/` into one entry) and one test defect (`--json` correctly outranks `--yes`; the never-dispose cell split into two invocations).
- **M5** — doctrine additions (see the deltas below).
- **M6** — REFACTOR + scope verification: lsof parser extracted (`parseLsofCWDs`), edge-path pins, coverage hardening.

### AC Matrix (all 14 PASS)

Command (one invocation, env-scrubbed with `unset MOAI_KANBAN* MOAI_FACTORY_WORKER &&`), run against this tree at HEAD `30f36cbd5`, full verbatim output at `.moai/state/verify/t1369/ac14_final.txt` — exit 0, 14 × `--- PASS`:

`go test -cover -v -run 'TestSweepLandedBranchDisposes|TestSweepNotLandedPreserves|TestSweepAncestryUnanswerable|TestSweepDirtyPreserves|TestSweepUnpushedPreserves|TestSweepLockedAndAnchored|TestSweepProcessCWDPredicate|TestSweepDryRunDefaultAndYes|TestSweepL1HoistThenRemove|TestSweepL2DonePath|TestSweepNeverDispose|TestSweepBaseFlag|TestSweepVerdictRecord|TestSweepRemovalFailure' ./internal/cli/worktree/`

| AC | Status | Test (all in internal/cli/worktree/sweep_test.go) | Actual output |
|----|--------|---------------------------------------------------|---------------|
| AC-WS-001 | PASS | TestSweepLandedBranchDisposes | `--- PASS: TestSweepLandedBranchDisposes (5.56s)` — tree gone, branch resolves, `Removed 1 worktree(s). Branches were left intact.` |
| AC-WS-002 | PASS | TestSweepNotLandedPreserves | `--- PASS: TestSweepNotLandedPreserves` — `cause=not-landed`, tree survives |
| AC-WS-003 | PASS | TestSweepAncestryUnanswerable (4 subtests) | `--- PASS: TestSweepAncestryUnanswerable (4.50s)` — fetch-failed / ancestry-exit-2 / unresolvable-base preserve with distinct causes, none `cause=not-landed` |
| AC-WS-004 | PASS | TestSweepDirtyPreserves (+ignored-content subtest) | `--- PASS: TestSweepDirtyPreserves (9.90s)` — dirty=yes, file intact |
| AC-WS-005 | PASS | TestSweepUnpushedPreserves | `--- PASS: TestSweepUnpushedPreserves (8.03s)` — remote-unreachable commit, `cause=not-landed` |
| AC-WS-006 | PASS | TestSweepLockedAndAnchored (3 subtests) | `--- PASS: TestSweepLockedAndAnchored` — lock/registry sources named; unreadable source → anchored=undetermined + exit 2 |
| AC-WS-007 | PASS | TestSweepProcessCWDPredicate (2 runtime subtests) + GOOS=windows build (cell c) | `--- PASS: TestSweepProcessCWDPredicate` — occupied/undetermined, never a negative |
| AC-WS-008 | PASS | TestSweepDryRunDefaultAndYes (3 subtests) | `--- PASS: TestSweepDryRunDefaultAndYes (14.21s)` — bare removes nothing + preview; `--yes` removes exactly the disposable; `--json` parity |
| AC-WS-009 | PASS | TestSweepL1HoistThenRemove (2 subtests) | `--- PASS: TestSweepL1HoistThenRemove (9.08s)` — hoist-before-remove order asserted; hoist failure → `cause=hoist-failed` |
| AC-WS-010 | PASS | TestSweepL2DonePath | `--- PASS: TestSweepL2DonePath (8.37s)` — done core called once with (false,false,true), branch survives |
| AC-WS-011 | PASS | TestSweepNeverDispose | `--- PASS: TestSweepNeverDispose (6.01s)` — main/own tree absent; on-base and locked preserve; healthy candidate disposes |
| AC-WS-012 | PASS | TestSweepBaseFlag | `--- PASS: TestSweepBaseFlag (8.44s)` — bare `cause=not-landed` vs origin/develop; `--base origin/main` → landed |
| AC-WS-013 | PASS | TestSweepVerdictRecord (+TestSweep_JSONEmitsEveryNonProtectedTree) | `--- PASS: TestSweepVerdictRecord (3.75s)` — all 11 fields, no not-checked/undetermined on a fully-evaluated record |
| AC-WS-014 | PASS | TestSweepRemovalFailure + doctrine checks below | `--- PASS: TestSweepRemovalFailure (6.75s)` — non-blocking notice, remaining tree processed, exit 0 |

AC-WS-014 doctrine checks (commands run at M5, this tree @`62197bb38`):

- `wc -c .claude/rules/moai/workflow/worktree-integration.md` → `41930`; delta vs the §E.1 run-start baseline `41194` = **+736** (bound ≤ +1,200) — PASS.
- `grep -c '^#' worktree-integration.md` → `53` before (origin/develop measure) and `53` after — **no new headings** — PASS.
- `grep -c 'worktree sweep' <file>` → 1 hit in `worktree-integration.md`, 1 hit in `kanban-dispatch.md` — PASS; the kanban-dispatch addition names `worktree-integration.md` (companion-pointer reference) — PASS.

### §E item results

- **E2 Cross-platform build** (HEAD `30f36cbd5`): `go build ./...` → exit 0 (`HOST_BUILD_OK`); `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`WIN_BUILD_OK`); `go vet ./internal/cli/worktree/ ./internal/cli/` → exit 0.
- **E3 Coverage**: `go test -coverprofile -run 'TestSweep' ./internal/cli/worktree/` → sweep files `184/198` statements = **92.9%** (threshold ≥85% on new sweep files). Per-function: sweepAncestryOutcome/newSweepCmd/classifySweepVerdicts/sweepCWDOccupied/sweepTierOf/renderSweepJSON/renderSweepReport 100%, sweepEvaluate 90%, sweepIgnoredReason 92.9%, runSweep 83.3%, applySweepVerdicts 77.8% (residual: encode/list-error paths). Package-level figure under the sweep-only run: 39.0% (the package predates this SPEC; not the judged metric).
- **E4 Subagent-boundary grep**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ | grep -v _test.go | grep -v '// '` → 0 matches (exit 1).
- **E5 Lint**: `golangci-lint run --timeout=2m ./internal/cli/worktree/...` → `0 issues.` (baseline at run start was also 0 — no new findings).
- **E6 Commits** (local, branch WT-worktree-sweep; lane does not push — integration is the lane session's): `382a0d825` (M1), `e3428c98f` (M2), `7a0c1ee07` (M3), `aad231a08` (M4), `62197bb38` (M5), `30f36cbd5` (M6).
- **E7 Blockers**: none.
- **E8 RED evidence**: per-milestone verbatim outputs captured pre-GREEN (M1 compile failure; M2/M3/M4 assertion failures quoted above). Honest notes: (a) the lock-source-unreadable cell (inside TestSweepLockedAndAnchored) pins behavior already shipped in the M1 skeleton, so it passed on arrival — its M1-shape was RED'd only indirectly; (b) TestSweepUnpushedPreserves (AC-WS-005) is the AC-named pin of behavior RED'd by TestSweepNotLandedPreserves in M2; (c) the M6 edge-path pins harden error branches whose behaviors were implemented in M2–M4 under their own RED evidence.

### Repair round (post sync-audit FAIL 85.7, one BLOCKING finding)

- **F1 (fixed)** — partial/absent evidence retrieval no longer disposes. The seam `sweepHoistBeforeDisposal` (a thin alias of the done-path routine, which reports only hard failures) was replaced by `sweepHoistEvidence`, which runs the shared hoist routine and returns whether retrieval was COMPLETE: an unresolved project root (retrieval unattempted) and a destination-conflict partial copy (REQ-RLC-006 skips, never overwrites) both return complete=false. The apply loop now runs the full-retrieval guard ABOVE the tier switch, so BOTH tiers honor it — the L2 done core (whose internal hoist reports hard failures only) is never reached on an incomplete retrieval, and its own hoist is then a no-op copy (identical destinations). New cause token `cause=hoist-partial` joins the family (`hoist-failed` semantics unchanged); hoist.go's nil-return contract for its other callers is untouched. RED justification: the failing path is the auditor's quoted trace (hoist.go:82-83 skip → sweep.go:533 Remove, exit 0), reproduced live post-fix by the demo in `.moai/state/verify/t1369/repair_run.txt` context — the four new cells assert the inverse of the pre-fix behavior (destination-conflict, root-unresolved × L1/L2).
- **F2 (NOT fixed — known limitation, recorded per auditor scope)**: the removal-time window still re-reads only the ignored-content predicate; the cwd-probe and anchor predicates are not re-read at removal time — a process moving into a tree between classification and its removal turn is not re-detected. Follow-up card candidate.
- **F3 (absorbed)** — CHANGELOG.md sentence corrected: L1 trees are hoisted then removed directly by the sweep (they are not routed to the session-end path); L2 through the done contract; both tiers refuse disposal on incomplete retrieval.
- **F5 (absorbed)** — the empty-sweep assertion strengthened to the full `Nothing to sweep: 0 worktree(s) evaluated.` sentence.
- Doctrine files untouched (AC-WS-014 delta stays +736); `wc -c` re-checked at repair close: unchanged.
- Repair verification: scoped `go test -v -run 'TestSweepL1HoistThenRemove|TestSweepRemovalTimeIgnoredKeep|TestSweep' ./internal/cli/worktree/` (env-scrubbed single invocation) → exit 0, 30 top-level `--- PASS`, 0 FAIL (`.moai/state/verify/t1369/repair_run.txt`); live-binary demo of the destination-conflict path: verdict `DISPOSE` → hoist skips → `Keeping … cause=hoist-partial` → tree survives, destination file untouched; `go vet` 0, `golangci-lint` 0 issues, `GOOS=windows go build ./...` exit 0; full worktree package re-run after the repair → exit 0.

### Files changed (run-phase)

| File | Change |
|------|--------|
| internal/cli/worktree/sweep.go | NEW (~570 lines) — verb, verdict record, seams, classification, rendering, apply |
| internal/cli/worktree/sweep_cwd_posix.go | NEW — lsof cwd probe (pattern source: internal/cli/update_worktree_processes.go, cited) |
| internal/cli/worktree/sweep_cwd_windows.go | NEW — Windows stub: unanswerable → preserve |
| internal/cli/worktree/sweep_test.go | NEW (~1610 lines) — mock seam bundle + real-git fixture family, 30+ tests |
| internal/cli/worktree/sweep_cwd_windows_test.go | NEW — Windows-side stub test |
| internal/cli/worktree/root.go | registration line + Long verb enumeration (the package verb-registry guard requires both) |
| internal/cli/worktree/root_test.go | expected subcommand count 10 → 11 (guard constant) |
| .claude/rules/moai/workflow/worktree-integration.md | hoist [HARD] clause extended (+736 bytes, no new headings) |
| .claude/rules/moai/workflow/kanban-dispatch.md | one pointer sentence in the integration section (+322 bytes) |

PRESERVE surfaces untouched: clean.go, done.go, remove.go, guard.go bodies, internal/session/, internal/cli/update_worktree_processes.go, other SPEC directories. `make build` not required (no template-touched content; the rules files are not template mirrors — local-only doctrine per the repo's template-namespace contract, unmodified in templates/).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-30
run_commit_sha: 30f36cbd5
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (lane does not fetch/commit outside its own worktree; pre-commit re-reads of HEAD and branch performed before each of the 6 commits)
l44_post_push_fetch: not-run (lane does not push; integration window owned by the lane session)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass
cross_platform_build.windows: pass
total_run_phase_files: 9
m1_to_mN_commit_strategy: one commit per milestone (M1-M6), conventional subjects, card id in every message

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-30
sync_commit_sha: 7bf07cad9
sync_status: complete
b12_self_test_a: pass  # grep -c 'SPEC-WORKTREE-SWEEP-001' CHANGELOG.md → 0 pre-emission (no duplicate entry)
b12_self_test_b: pass  # AC count in acceptance.md = 14 (AC-WS-001..014); CHANGELOG entry cites 14 — match
b12_self_test_c: pass  # every file path in the CHANGELOG entry verified via ls/grep before commit
changelog_entry_position: CHANGELOG.md [Unreleased] → Added, first bullet
frontmatter_status_transitions.in-progress_to_completed: 2026-09-30  # single sync commit (3-phase close merged)
canary_compliance_check:
  docs_site_touched: false  # docs-site (adk.mo.ai.kr) content is oss-docs harness scope — follow-up candidate only
  readme_touched: true      # worktree verb enumeration +`sweep`, ko-canonical, 4-locale same-commit (ko/en/ja/zh)
  readme_parity: pass       # H2 counts 12/12/12/12, code fences 46 each, switcher headers intact, URL blacklist clean
mx_tag_validation: sync sub-step pass  # no @MX:TODO residue in new sweep files (E2-E6 clean; lint 0 issues)
sync_note: run-phase §E.2/§E.3 evidence unchanged; sync is artifact-only (CHANGELOG + README 4-locale verb row + this §E.4 + spec.md status transition).
