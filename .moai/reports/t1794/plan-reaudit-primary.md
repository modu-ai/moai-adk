auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: e04ebd5ca6bc441b380c1072d5a1dff8431459ba
convergence_overall: fail
required_backend: codex inconclusive rcpt-5fe3332abb3abdd31722331a rcpt-e1a8901f72435ba0fd4dcbd0
plan_artifact_hash: GAP
overall_score: 0.81
must_pass_failed: 0
blocking_count: 1
scope: full
fix_scope: .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/spec.md#R-1 .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/plan.md#D-3 .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/design.md#Limits .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/acceptance.md#edge-cases
defect_class: design

# SPEC Review Report: SPEC-MEMORY-FOLD-RENAME-RACE-001

Iteration: re-audit (primary checkout, card t1794 receipt-clearing request)
Verdict: FAIL
Overall Score: 0.81
Plan Artifact Hash: GAP (no audit_cache tool in this session; not computed)
Reasoning context ignored per M1 Context Isolation.

Subject: the plan artifacts as they stand on local `main` (`e04ebd5ca6bc441b380c1072d5a1dff8431459ba`), extracted with `git show main:<path>`. The primary checkout's HEAD is detached at `478285ab4c1` (863 commits behind `main`) and does NOT contain `.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/` (`git ls-tree HEAD <dir> | wc -l` → 0).

Why FAIL: (1) the required codex backend returned no verdict twice (unmet gate — not a reviewed defect), and (2) one blocking finding (D1): the SPEC's statements about the lock-file open contradict the merged code on `main`.

## Must-Pass Results

- [PASS] MP-1 REQ numbering: spec.md:L60-67 define REQ-MRR-001..008, sequential, no gaps or duplicates.
- [PASS] MP-2 GEARS (requirement layer only): L60 "The fold apply path shall hold…" (Ubiquitous); L62 "When another holder keeps…, the requesting fold shall refuse…" (Event-driven); L65 "While the card-close fold step waits on or holds…, the step shall…" (State-driven); L67 "When the store lock refuses…, the fold shall leave…" (Event-driven). ACs are Given/When/Then in acceptance.md — not graded here.
- [PASS] MP-3 frontmatter: spec.md:L2-13 carry all 12 canonical fields (`version: "0.1.5"` quoted, `status: completed`, `priority: P1`, `lifecycle: spec-anchored`, `tags` string). `moai spec lint <copy>/spec.md` → `0 error(s), 2 warning(s)`, exit 0.
- [N/A] MP-4: Go-internal, single-language scope.
- [PASS] MP-5 D7: references SPEC-MEMORY-FOLD-BUDGET-001 and SPEC-MEMORY-INDEX-FOLD-001, both `status: completed` on main. No retired/superseded reference.
- [PASS] MP-6 D8: `grep -n syscall` over spec/plan/acceptance/design → no match.
- [PASS] MP-7: `grep -n 'NEEDS CLARIFICATION'` over plan.md/spec.md/acceptance.md → no match; research.md absent.
- [N/A] MP-8: acceptance.md:L11 "No criterion in this SPEC holds this class" (release-blocking). AC-MRR-007 command 2 re-executed anyway: `git diff --exit-code 2aab5f797 f30bb9088cde7b12612993026b5e8d2f1244e4f0 -- go.mod go.sum` → exit 0, empty.
- [PASS] MP-9: CN-4 verb → `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 0 exit bindings, 23 ordering candidates`; no CONFLICT. The two milestone-bearing candidates (acceptance.md:L64, L108) order steps inside M5 (RED before GREEN), consistent with plan.md M5 step 1.

## Category Scores

| Dimension | Score | Evidence |
|---|---|---|
| Clarity | 0.75 | REQs precise; but the lock-open premise (spec §4 L91-93) no longer describes the merged mechanism (D1) |
| Completeness | 1.0 | HISTORY L18, background §1, REQs §2, constraints §3, Out of Scope §4 with five `### Out of Scope —` H3s, residual risk §5; acceptance DoD §4 |
| Testability | 0.75 | Commands are single invocations with exit expectations; AC-MRR-008 post-close and AC-MRR-010 unadopted are declared honestly |
| Traceability | 0.75 | acceptance.md §2 maps all 8 REQs; REQ-MRR-001's cross-process clause has only the unadopted AC-MRR-010 (declared as R-2) |

## Defects Found

D1. SPEC-vs-merged-code drift — spec.md:L91-93 (§4 symlink OOS), spec.md:L99 (§5 R-1), plan.md:L45 (D-3), design.md:L38, acceptance.md:L132 — every one states the lock open "follows symbolic links and performs no identity check" and that hardening "would need its own SPEC or a reopening of this one". On `main`, `internal/cli/fold_store_lock_unix.go:69` opens with `unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW`, and `fold_store_lock_windows.go:65` refuses a symlink via `os.Lstat`. The flag was absent at the SPEC's pin (`git show f30bb9088:internal/cli/fold_store_lock_unix.go | grep -c O_NOFOLLOW` → 0) and arrived with `f518deac8` (2026-10-11, card t1595 integration merge). No HISTORY/amendment row records this change; R-1's residual risk is therefore stale, and no criterion covers the new refusal. — Severity: major — Class: blocking — Required fix: add a HISTORY amendment row naming `f518deac8` (t1595) as the change that added O_NOFOLLOW (POSIX) and the Lstat pre-check (Windows); restate R-1 as narrowed (POSIX closed by the open flag; Windows Lstat pre-check is check-then-open), and update plan D-3, design.md Limits and acceptance §3 to match — or state explicitly that this SPEC's text is frozen at `f30bb9088` and point to the SPEC/card that owns the hardening.
D2. Line citations pinned to `2aab5f797` (spec.md:L46-47 `:414`, `:649`; plan §C.2) no longer match main (`applyFold` is at `memory_fold.go:461`). The pin is stated explicitly, so this is not wrong — Severity: minor — Class: optional — Required fix: none required; optionally add "(pinned to 2aab5f797)" beside P1/P2.
D3. Lint warnings (acceptance.md:L106, L165): assertion strings `--- PASS: TestFoldStoreLockCrossProcess` / `--- FAIL: TestFoldRenameWindowConcurrentWriter` lack a trailing delimiter (prefix-match vacuity) — Severity: minor — Class: optional.

Merged-code checks that AGREE with the SPEC: all cited tests exist on main (`TestFoldRenameWindowConcurrentWriter`, `…ContentionRefusesCleanly`, `…LockFileInvisibleToTooling`, `…OnDoneContentionAbandonsWithoutWrite`, `…SpanHeldThroughApply`, `…WindowsSemantics`, `TestReviewSequentialAbandonedTempOwnership`, `TestReviewArchiveUpdateDuringEffectiveScan`, `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename` — 1 each); `TestFoldStoreLockCrossProcess` is absent (0), matching AC-MRR-010 "unadopted"; `TestMemoryFold*` = 27 and `TestMemoryFoldOnDone_*` = 13, matching AC-MRR-007; `applyFold` acquires the lock first and defers release (`memory_fold.go:471-475`), matching REQ-MRR-001. The on-done path now holds the lock across its whole transaction through `runUnderFoldLock` (`:823-827`) and `applyFoldLocked` — wider than REQ-MRR-001, not contradicting it.

## Claim

FAIL. No must-pass criterion fails; one blocking drift finding (D1); the required codex gate is unmet.

## Evidence

- `git rev-parse main HEAD` → `e04ebd5ca6bc441b380c1072d5a1dff8431459ba` / `478285ab4c1fc893b020bb464d14cb69c8422523`; `git rev-list --count HEAD..main` → 863.
- `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `mcp__moai__codex_audit(project_root=/Users/goos/MoAI/moai-adk-go, target=baseBranch, mode=adversarial)` attempt 1 → `verdict: fail`, `gate_unmet: workflow.audit.gates.codex is required, but this audit returned no verdict (fail-open inconclusive)`, summary "`codex stdout closed before response to id=1`… HEAD `478285ab4c` … inconclusive", `audit_receipt: rcpt-5fe3332abb3abdd31722331a`.
- Attempt 2 (single retry) → same gate_unmet; summary "internal check `exit_code: 2` … `codex stdout closed before response to id=1`. The designated SPEC path is not in the current tree", `audit_receipt: rcpt-e1a8901f72435ba0fd4dcbd0`.
- `moai verify audit-plan … --result-file .moai/state/audit-plan-result.json` → `convergence_check: {ok: false, unmet: ["codex"], reason: "codex: entry does not answer a required gate (verdict \"inconclusive\", gate \"required\") …"}`.
- `moai spec lint` (copy of main's spec.md) → `0 error(s), 2 warning(s)`, exit 0.
- `git grep -n O_NOFOLLOW main -- internal/cli/fold_store_lock_unix.go` → `:59`, `:67`, `:69` (`unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o644)`).
- `git log main --first-parent -G O_NOFOLLOW -- internal/cli/fold_store_lock_unix.go` → `f518deac8 2026-10-11 Merge WT-dispatch-integrity into main (card t1595, integration merge)`.

## Baseline-attribution

All reads against commit `e04ebd5ca` (local `main`) via `git show`/`git grep main`; the CLI build answering was `69c1ffc1f4` (MCP server banner; codex result `build_commit`). Commands and output above were run in this session.

## Gaps

- Required codex backend: inconclusive on both attempts (transport failure `codex stdout closed before response to id=1`). Unmet gate, not a reviewed defect. Its `baseBranch` target resolved against the primary checkout's detached HEAD `478285ab4`, which lacks this SPEC directory, so even a completed codex review would not have read these artifacts.
- `mcp__moai__audit_multi` was not called although `cross_model_active: true`; the caller directed a single `codex_audit` leg (codex is the only enforced backend). The convergence check was run on a hand-written digest of the codex_audit result and also flags `plan_source` mismatch.
- `plan_artifact_hash` not computed (no audit_cache tool available).
- The Group 4 traceability verb collected `0 REQ definitions` (the `- REQ-MRR-00N — **…**` form is not one of its three definition shapes); traceability was mapped by hand from acceptance.md §2 (8 rows, one per REQ).
- No test was executed (plan-phase re-audit; MP-8 N/A). Merged behavior was read, not run.

## Residual-risk

- D1 was judged by reading code, not by running a symlink probe; the Windows Lstat guard's check-then-open window was not measured.
- R-2 (cross-process clause) and R-3 (card-close span) remain unverified exactly as the SPEC records them.
