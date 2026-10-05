# progress.md — SPEC-MOAI-HYGIENE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-MOAI-HYGIENE-001 v0.3.0 authored 2026-10-05 by manager-spec (card t1518, plan phase, worktree `WT-audit-log-gc` @ develop `6643c7bba`).
- Revision history: v0.1.0 (plan-audit iter 1/3 audited `8039ea714`, FAIL 0.73, D1–D14) → v0.2.0 (D1–D14 remediation: MP-8 evidence ledger, REQ consolidation 19 → 16, registry-absent classification, mtime prohibition, re-stat-under-lock, symlink refusal, mode precedence; iter 2/3 audited `58c01b4e1`, FAIL 0.81, D15–D22 blocking + D23–D25 optional) → v0.3.0 (D15 rotation apply-mode-only — one decision in REQ-HYG-004, no exemption clauses; D16 staging artifacts as rotator-owned registry entries + under-lock crash recovery; D17 Windows rotation gated on a verified LockFileEx sidecar lock, else skip+retry; D18 lock class excluded from GC scope — empty-file dating impossible + two-live-locks window, upstream parked at Q7; D19 symlink refusal scoped strictly below the resolved root + swap-after-check residual named; D20 transcript-absent ⇒ unmeasured + `HygieneHeartbeatStaleWindow`; D21 per-class shape/dating/grouping table with M3 field-pinning REDs; D22 runtime `testing.Testing()` root guard + escape-pattern greps + L-C3 control; D23 rotation-count bound; D24 `would-probe` + hash exclusion; D25 reclaim limitation + Q8).
- Tier M artifact set complete: spec.md, plan.md, acceptance.md, progress.md, decision-index.md.
- Evidence basis: 2026-10-04 read-only hygiene audit (`hygiene.md` + `hygiene.json`, scratchpad) + plan-phase code-owner mapping (spec.md §G) + plan-audit iter-1 verdict (`.moai/reports/t1518/plan-audit.md`).
- SPEC ID pre-write check: `PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`, executed in Bash); catalogue dedup confirmed (1,048 existing SPECs, no MOAI-HYGIENE entry).
- MP-8 evidence ledger (acceptance.md §A.1, L-001..L-016 + L-C1/L-C2/L-C3): every cell carries command + verbatim stdout + exit code + per-entry tree SHA pin — L-001..L-011, L-013..L-016 and the L-C1/L-C2 controls at `8039ea714c988f4264cd8785405dbb1c410370b6`; **L-012 re-captured at v0.3.0 under the renamed test at `58c01b4e15f483f25c261ced08e92d7647ea7815`**; 14 RB go-test cells observed RED (12× `FAIL ./internal/hygiene [setup failed]` exit 1; 2× `ok … [no tests to run]` exit 0 — empty-sweep reds), 2 RG grep cells observed structurally red (exit 2, targets absent) with seeded positive controls proving the patterns non-vacuous (L-C1: 5 hits; L-C2: 1 hit — both exit 0); **L-C3 (escape-form control) is an authored obligation awaiting its M1 observation by design**.
- Spec lint after revision (measured 2026-10-05, this tree's build): `go run ./cmd/moai spec lint SPEC-MOAI-HYGIENE-001` → `✓ No findings — all SPEC documents are valid` (16 REQ / 16 AC, all maps lines resolved).
- plan-phase scope held: read-only on `internal/` — no implementation files touched.

## §E.2 Run-phase Evidence

Run phase executed by manager-develop (card t1518) on branch `t1518-run` in the agent's isolated worktree, linear on the card HEAD `bde9be69c` (the spawn type's worktree isolation guard refuses card-tree git — recorded in §E.3; the lane lands the branch with one `--ff-only` merge). Commits, all carrying `card t1518`:

| Commit | Content |
|---|---|
| `e7c564b38` | STEP 0 — plan-audit iter-3 defect closure (D26–D35), spec v0.4.0, status draft → in-progress |
| `3ba96f561` | M1 — rotator core + sink registry + completeness guard (v0.4.1: registry extended with the 11 writers the guard verified) |
| `26593db6f` | M2 — three-signal fail-closed liveness evaluator |
| `b410c82cf` | M3 — GC sweep, modes, audit granularity, symlink defense |
| `fa4c202f0` | M4 — named thresholds, workflow.hygiene config, `moai clean` extension |
| `7a9957d09` | M5 — SessionStart wiring, best-effort, never blocks launch |
| (this commit) | M6 — closure: lint fixes, MX seam annotations, §E.2/§E.3 |

### AC → green table (each line: the exact ledger command, re-run at closure)

| AC | Command (verbatim) | Observed result |
|---|---|---|
| AC-HYG-001 (L-001) | `go test ./internal/hygiene/ -run '^TestRotator_RotatesOverThreshold$' -count=1` | `ok … internal/hygiene 0.296s` — exit 0 |
| AC-HYG-002 (L-002) | `go test ./internal/hygiene/ -run '^TestRotator_KeepOneCap$' -count=1` | `ok … 0.129s` — exit 0 (crash-recovery arms green) |
| AC-HYG-003 (L-003) | `go test ./internal/hygiene/ -run '^TestRotator_UnderThresholdAndAbsent$' -count=1` | `ok … 0.117s` — exit 0 |
| AC-HYG-004 (L-004) | `go test ./internal/hygiene/ -run '^TestRotator_ConcurrentSerialize$' -count=1 -race` | `ok … 1.386s` — exit 0 (stale/no-exclusion arms green) |
| AC-HYG-005 (L-005) | `go test ./internal/hygiene/ -run '^TestSinkRegistryCompleteness$' -count=1` | `ok … 0.359s` — exit 0 (real-tree arm green over the v0.4.1 registry) |
| AC-HYG-006 (L-006) | `go test ./internal/hygiene/ -run '^TestLivenessVerdictMatrix$' -count=1` | `ok … 0.108s` — exit 0 (neverDead flips assert unmeasured≠DEAD) |
| AC-HYG-007 (L-007) | `go test ./internal/hygiene/ -run '^TestReportModeByteIdentical$' -count=1` | `ok … 0.229s` — exit 0 (hash-equal; no lockfile; 1+1 summary rows; self-sink `skipped-report-mode`) |
| AC-HYG-008 (L-008) | `go test ./internal/hygiene/ -run '^TestApplyModeDeletionSet$' -count=1` | `ok … 0.124s` — exit 0 (D28 race arm + D29 crash arm + already-gone green) |
| AC-HYG-009 (L-009) | `go test ./internal/hygiene/ -run '^TestAuditRowsComplete$' -count=1` | `ok … 0.123s` — exit 0 |
| AC-HYG-010 (L-010) | `go test ./internal/hygiene/ -run '^TestUnitIndependence$' -count=1` | `ok … 0.117s` — exit 0 |
| AC-HYG-011 (L-011) | `go test ./internal/hygiene/ -run '^TestSymlinkRefusalParentSwap$' -count=1` | `ok … 0.119s` — exit 0 (anchored-action refusal; external sentinel survives) |
| AC-HYG-012 (L-012) | `go test ./internal/hygiene/ -run '^TestLockClassExcluded$' -count=1` | `ok … 0.120s` — exit 0 (both modes; registry structurally lock-free) |
| AC-HYG-013 (L-013) | `go test ./internal/hook/ -run '^TestSessionStartHygieneBestEffort$' -count=1` | `ok … internal/hook 0.684s` — exit 0, run test count ≥ 1 (empty-sweep RED flipped per §1.1) |
| AC-HYG-014 (L-014) | `go test ./internal/cli/ -run '^TestCleanHygieneFlags$' -count=1` | `ok … internal/cli 0.827s` — exit 0, run test count ≥ 1 (7 subtests: dry-run, config-apply-without---apply, apply, 3×D30, rotator-dry) |
| AC-HYG-015 (L-015) | widened threshold grep over `internal/hygiene` | exit **1**, empty output; L-C1 control re-observed: 6 hits, exit 0 |
| AC-HYG-016 (L-016) | full escape-pattern grep over `internal/hygiene` + both wiring test files | exit **1**, empty output; L-C2 1 hit / L-C3 3 hits, exit 0; runtime guard asserted (non-temp root refused, registered temp root accepted) |

### §E obligations

- `go test ./internal/hygiene/... -count=1 -race`: `ok … 3.065s` (final closure run). Package coverage **85.9%** (≥ 85 target); production plumbing additionally driven cross-package by the CLI/hook wiring suites.
- `go vet ./internal/hygiene/`: clean. `golangci-lint run internal/hygiene/...` at the CI version **v2.1.6**: **0 issues** (6 errcheck fixes landed at M6).
- `GOOS=windows go build ./internal/hygiene/ ./internal/cli/`: pass; `GOOS=windows go vet ./internal/hygiene/`: pass (LockFileEx sidecar path compiles).
- Isolation verification (DoD #3): content hash of the real `.moai/logs` + `.moai/state` taken before/after the verification batch — **identical** (`44c8b5cb…` both sides; the own-session runtime trace writer excluded as the named non-verification observer; the full-suite run had flagged exactly that file, nothing else). No `hygiene-audit.jsonl` exists in the real logs dir — no test touched the real tree.
- Mutant-probe record (test-level, per §C): AC-HYG-006 — the matrix's neverDead rows force every OTHER signal negative and assert no DEAD on unmeasured; AC-HYG-008 — the mtime-only seed, goal-triple partial-deletion, D28 fresh-state and D29 crash arms each fail under their mutant; AC-HYG-004 — the stale-decision and held/unverifiable-exclusion arms assert no action; AC-HYG-002 — recovery arms assert the staged chunk is never overwritten/deleted; AC-HYG-012 — both modes assert lock bytes untouched and the registry lock-free.
- Config: `workflow.yaml` template carries the neutral `hygiene:` block; `internal/config/defaults.go` carries the six named constants; the six keys are in the shipped-key triage inventory (class W).

### Real-repo dry-run demonstration (report only — no mutation)

Built from this tree (`go build -o /tmp/t1518-moai ./cmd/moai`), run against the actual primary checkout:

```
$ cd /Users/goos/MoAI/moai-adk-go && /tmp/t1518-moai clean --audit-logs
· resolved project root: /Users/goos/MoAI/moai-adk-go
· hygiene: dry-run (pass --apply on this invocation to mutate)
· [dry-run] summary: skipped-absent=8, skipped-report-mode=1, skipped-under-threshold=11
```
(counts close: 8+1+11 = the 20 registered sinks; the one over-threshold sink is recorded `skipped-report-mode` — all rotation apply-only.)

```
$ /tmp/t1518-moai clean --session-state   (same invocation surface)
· resolved project root: /Users/goos/MoAI/moai-adk-go
· hygiene: dry-run (pass --apply on this invocation to mutate)
· [dry-run] kept: state/agent-stops/<uuid>.json — indeterminate liveness — kept with reason (fail-closed)   (…every candidate)
· [dry-run] kept: state/agent-stops/3786c20e-….json — live session
```
(every candidate kept WITH its reason; one genuinely live session affirmatively detected via a fresh transcript under the real profile roots — the fail-closed liveness machinery works on live data.)

Post-run state of the real logs dir: no `*.1` / `*.staging` artifacts, no deletions; the only new file is `.moai/logs/hygiene-audit.jsonl` — the REQ-HYG-010 named exception — carrying exactly the two summary rows:
`{"ts":"2026-10-05T03:46:37Z","unit":"rotator","mode":"report","outcome":"summary","counts":{"skipped-absent":8,"skipped-report-mode":1,"skipped-under-threshold":11}}` (verbatim; the gc row likewise).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-05T12:47:00Z+09:00
run_commit_sha: 7a9957d09   # last code commit (M5); the M6 evidence commit lands this file
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a — commits landed on the agent-isolated worktree branch t1518-run (linear on card HEAD bde9be69c); no shared-checkout git operations were performed (the worktree isolation guard refuses cross-tree git for this spawn type); the lane lands the branch with one --ff-only merge into WT-audit-log-gc
l44_post_push_fetch: n/a — push is the lane's batch (operator's push-lead discipline)
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6: 0 issues; go vet clean
cross_platform_build:
  windows: pass   # GOOS=windows go build + go vet on internal/hygiene and internal/cli
total_run_phase_files: 24   # 16 new (internal/hygiene 14, wiring tests 2) + 8 modified
m1_to_mN_commit_strategy: one commit per milestone (STEP0/M1..M5/M6), each conventional, card t1518 trailer, pathspec staging, HEAD+branch re-read immediately before each
pre_existing_failures_disclosed:
  - TestStaleRunNoticeLegacyLeaderSpelling / LegacySessionRecord / FactoryLegacyLabel (internal/hook) — fail identically on base bde9be69c (verified in a throwaway clone at base); dependency files untouched by this card; out of scope here, recommended for its own card
worktree_note: >
  manager-develop spawns auto-isolate to their own agent worktree and the
  isolation guard refuses card-tree git (two refusal shapes observed and
  recorded; routing around it is prohibited by doctrine). The work ran on a
  branch created at the card HEAD inside the agent tree; land it with:
  git -C <card-tree> merge --ff-only t1518-run
  (history stays linear on bde9be69c; every commit already carries card t1518).
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

