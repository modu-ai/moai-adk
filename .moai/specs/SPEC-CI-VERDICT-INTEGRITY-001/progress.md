# progress.md — SPEC-CI-VERDICT-INTEGRITY-001

Card: t1534 · Branch: WT-ci-verdict-integrity · Plan-phase tree: a158b4b5f

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready-with-recorded-debt (operator disposition, 2026-10-06)
- plan_complete_at: 2026-10-06 (operator AskUserQuestion, lane-24 terminal)
- admission_record: 6 audit rounds (0.62 → 0.76 → 0.78 → 0.79 → 0.78 → 0.70); final 0.70 < 0.80
  routed to the operator per the pre-agreed rule. Operator disposition: **debt closure +
  conditional run entry** — the plan body (13 REQ, 14 AC, M1–M3, keep-set structure) has had
  ZERO findings since iteration-2; the 7-item residual lives entirely in the repro/evidence
  infrastructure.
- run_entry_conditions (STEP-0, mandatory): the hold-record fix list
  (`.moai/reports/t1534/hold-record.md`) must be landed BEFORE M1 flip evidence is measured —
  D1 held ✓ (amendment-6, iter-5 re-verified); remaining: P2-L stub log-write/mktemp exit
  semantics, P2-M real-merge-body observation with recorded stub (fixture SHA + served time),
  P2-O single-run mid-run clock progression + re-evaluation hold case, P2-N 25-case matrix +
  acceptance.md P3 wording, P2-Q boundary-limited extractor across E2/E4/E7, P2-S residual
  `${{ }}` = harness error, P2-T runner-options parity (`bash -e`, no pipefail). M1 flip
  evidence measured with unfixed probes is INVALID (P2-L/M/O/P/Q each demonstrably produce
  false verdicts — gate-measured).
- checkpoint: M2 keep-set apply package delivery requires the STEP-0 fixes landed AND any new
  evidence-infra finding from M1 measurement closed or surfaced; unresolved infra findings at
  that point → hold re-enters (operator re-judged checkpoint, not lane discretion).
- decision record: decided_by=lane-24 (orchestrating lane) relaying operator AskUserQuestion
  (terminal, 2026-10-06) evidence_refs=.moai/reports/t1534/{plan-audit.md,hold-record.md,
  plan-audit-iter1..5-*} ladder_path=operator keep-set direct (autonomy policy keep-set
  channel)

## §E.2 Run-phase Evidence

### M1 — five gate mis-judgment repairs (flip evidence; measured post-M1 on WT-ci-verdict-integrity)

| AC | RED (pre-M1, pinned a158b4b5f) | M1 flip observation (verbatim) |
|----|-----|--------------------------------|
| AC-CI-001/002 | L-E2 (cancelled read as PASSED, exit 0) | E2 re-run: `::error::Release PR multi-OS verification concluded 'cancelled' …` exit 1; exclusion A (detect=skipped) → `exclusion=non-release-pr` exit 0; exclusion B (docs-only, go_code=false) → `exclusion=docs-only-release-pr` exit 0 — three shapes measured |
| AC-CI-004 | L-E4 (lookup failure → should_merge=true) | E4 re-run: `::warning::required-checks lookup failed (gh exit 1) - withholding auto-merge` + `--- GITHUB_OUTPUT --- should_merge=false` — the repaired JSON step (name/state/bucket + explicit rc handling, exit 8 = wait) withholds |
| AC-CI-006/007 | L-E7 (25-case matrix) + L-E8 | E7 re-run post-M1 (measured): control exit 0; 5 `failure` variants exit 1 (kept — regression guard); `install-script-parity-failure` + all 18 non-success variants exit non-zero (flipped); compatibility probes are hard checks (measured parity-failure visibility via the needs-list repair) |
| AC-CI-005 | L-E26 (merge served post-crossing, exit 0) | E26 re-run on the repaired body: head re-query observed → served clock read 1200 (past deadline) → merge WITHHELD → **exit 2 FLIP OBSERVED** (213dd949e added the post-head-query re-evaluation — the pre-fix tree measured exit 0 RED through the same probe); GATE-4 control PROBE_DEADLINE=1300 observes the legal in-deadline merge |
| AC-CI-011/012 | E14 (grep 'branch-protection.json.gtmpl' ci.yml = 0, exit 1) | AC-CI-011: validator Dimension D publishes-verification (measured exit 0, all 9 contexts publishable). AC-CI-012: ci.yml go_code filter gained `.github/branch-protection.json.gtmpl` + correspondence guard `TestBranchProtectionDetectFilterCoversParityInput` PASS (internal/template/branch_protection_detect_filter_test.go — re-derives if the parity test stops reading the covered input) |

STEP-0 (pre-M1 gate verification): LG-1 CONFIRMED (correct implementation scored DEFECT exit 0 — late_calls counted the head re-query; fixed: late flag judged only on `gh pr merge` lines), LG-2 CONFIRMED (`CROSSED` marker line corrupted cat-based integer comparisons; fixed: clock file numeric-only), LG-3 CONFIRMED (same-indent next-step boundary missed; fixed: dedent test `<=` across all four extractors; verified LG-1 now exit 2 and LG-3 now exit 9 on regenerated probes). Incident recorded: the first LG-1/LG-3 fixture runs invoked the REAL `gh` binary (HERE-detached /tmp probe copy missed the recording stub) — `gh pr merge 1` attempted against modu-ai/moai-adk and REFUSED (PR 1 CLOSED, never merged, no state change; verified read-only).

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete (STEP-0 + M1 + M2 + M3 all landed)
- run_complete_at: 2026-10-06 (lane-24 direct completion after the run
  delegate was stopped output-silent at 2h27m; M1 landed by the delegate as
  d74707a9e before the stop; M2/M3 completed lane-direct per the
  consecutive-stall rule — zombie delegate TaskStopped, work checkpointed on
  disk, zero loss)
- run_commits: d74707a9e (M1 five gate repairs + STEP-0) → 845e9a20c (M2 SSoT
  correction + validator publishability) → 6a8107f41 (M3 ci-watch
  supported-field repair + harness transition)
- verification: ci-watch test harness 12 pass / 0 fail (incl. field-contract
  regression + missing-required-pending); validator exit 0 — all 9 main
  contexts publishable incl. per-combination Build x5 and Analyze (Go) (go);
  E26 RED exit 0 (merge served 1200 past deadline 1140 through the real
  merge body) + GATE-4 control PROBE_DEADLINE=1300 observes the legal
  in-deadline merge; sh -n clean on all touched scripts; blank-line variant
  repro exits 0 identically (GATE-2 closed)
- M2 keep-set package: PENDING — the branch-protection apply payload,
  pre-apply live-diff command, and post-apply GET command are packaged at
  `.moai/worktrees/t1534/.moai/reports/t1534/` for the operator
  (AskUserQuestion delivery per the keep-set channel; delivery due next)
- debt conditions status: D1 held through iter-5/6 (fixture byte-conformance,
  not re-raised); the 7-item repro-infra list landed via amendments 5-7 and
  is VERIFIED by the STEP-0/E26 measurements above

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete (single sync commit: CHANGELOG [Unreleased] Fixed entry +
  spec.md frontmatter `in-progress → implemented → completed` + this §E.4;
  `sync_commit_sha` backfilled in the follow-up commit)
- sync_complete_at: 2026-10-06
- sync_commit_sha: eaf2e34f4
- b12_self_test_a: pre-emission grep `grep -c 'SPEC-CI-VERDICT-INTEGRITY-001'
  CHANGELOG.md` → 0 (exit 1) — no duplicate entry from a parallel sync
- b12_self_test_b: AC count match — the counter on acceptance.md (tier M AC
  source) reports live=14, excluded=0, ambiguous=0; the CHANGELOG entry
  references the same 14 (AC-CI-001..014)
- b12_self_test_c: file path verification — every path named in the entry
  verified present via `ls` at write time (11 paths, 0 misses)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` → `### Fixed` —
  first entry (reverse-chronological house order)
- frontmatter_status_transitions: in-progress → implemented → completed on the
  single sync commit (manager-docs, per the Status Transition Ownership Matrix;
  `status:` + `updated:` only, no body change — `updated:` already 2026-10-06)
- canary_compliance_check: n/a — this SPEC defines no forward-looking policy
  gated by its own sync tests
- ac_ci_012_resolution: Path A (implement-flip, coordinator disposition) —
  amendment `f6fbc0d9e` added the ci.yml filter entry + the correspondence
  guard `TestBranchProtectionDetectFilterCoversParityInput`; flip re-measured
  by manager-docs (E14 probe grep 0→1, guard test PASS `ok 0.406s`)
- ac_ci_009_evidence_path_note: acceptance.md names
  `.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` as the keep-set
  delivery surface; the actual package lives at
  `.moai/reports/t1534/keep-set-package.md` (card evidence path is
  authoritative) — resolved by this note, no manager-spec wording touch
- codemaps: omitted — `ci-watch` has zero representation in
  `.moai/project/codemaps/` (0 hits incl. modules.md); `scripts/` shell
  tooling is not a codemap fold unit. Recorded as omission; scope not expanded
- mx_tag_check: zero `@MX` tags present on the touched surfaces (4 workflows +
  2 scripts + the new guard test) — none stale, none owed
- sync_delegate: manager-docs (card t1534, lane-24 orchestration)
- gate_round_4 (post-CI-no-publish hold, lane-direct): 4 repairs, all probe-verified
  before commit `ef8ebd702` — (1) run.sh GH/POLL init moved before the PR
  base-branch block (gate caught $GH used-when-unset); (2) run.sh worse()
  aggregates cancel as FAIL in every entry order ([pass,cancel]→fail,
  [cancel,pass]→fail, [pass,pass]→pass observed); (3) validator
  matrix.exclude subtraction — repro `run-matrix-exclude.sh` RED exit 0
  (excluded combo counted publishable) → GREEN exit 1 naming the phantom,
  positive controls intact (phantom=1 / phantom-control=0 / malformed=1 /
  repo-root=0); (4) AC-CI-012 guard scoped to the go_code filter block —
  mutant probe (entry relocated to top level) FAILs, restored tree passes.
  Pushed a3febc2dd..ef8ebd702; CI publication still platform-blocked
  (04:17Z+ no-publish gap — leader watching).
