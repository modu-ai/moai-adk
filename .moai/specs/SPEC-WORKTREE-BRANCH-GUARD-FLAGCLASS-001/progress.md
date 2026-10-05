# progress.md — SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-03
plan_tier: M
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
plan_phase_note: >
  Class B defect card t467. Root cause measured on tree d592b0551
  (plan-phase synthetic probe, output preserved verbatim in acceptance.md
  §D.0): 10 git-branch mutation forms outside the [dDmMcC] class pass the
  matcher; query axis clean in source. Milestones M1 (measurement matrix,
  decides fix class) → M2 (matcher fix) → M3 (tests + condition docs).
  No commits made in plan phase (orchestrator commits after reading).
audit_iterations:
  - "1: FAIL 0.79 (report: .moai/reports/plan-audit/SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001-review-1.md) — fixes D1-D7 applied in spec/plan/acceptance v0.2.0 (D1 extends+REQ-009 bounded doctrine update; D2 matrix M-20..M-23 + M-14 non-target annotation; D3 §G corrected discrimination rule; D4 evidence labels; D5 tier wording; D6 citation; D7 AC-008 negative-path cell); re-audit (iteration 2 of 2) pending"
  - "2: score 0.83 (clears Tier M 0.80; verdict FAIL defect-driven, Tier M ceiling 2/2 reached) — all D1-D7 verified RESOLVED; fixes D8-D14 applied in v0.3.0 (D8 M-24 attached -u spelling + §G rule 2 'beginning with u'; D9 rule-extension route: positional-creation rule replaces false coverage sentence, cells M-25..M-28; D10 t added to short-cluster set; D11 M1 expected-count 23 rows/24 variants with §D.1 named authority; D12 AgentType SSOT contradiction acknowledged + M3 payload capture + doc-reconciliation blocker path; D13 E-3 git-rejected annotation; D14 git-form-liveness vs guard-allow label split); escalation per Retry Loop Contract is the orchestrator's (auditor recommended micro-fix-then-proceed)"
  - "run-gate: FAIL 0.88 (trajectory 0.79→0.83→0.88; score clears, verdict defect-driven) — D8-D14 all verified RESOLVED; fixes G1-G5 applied in v0.4.0 (G1 value-consuming arity rule generalized + pinned set extended --no-contains/--sort/--color/--abbrev/--column + Q-13/Q-14 space-separated-value cells; G5 M-29 -vt cluster + M-30 --create-reflog cells; G2 AC-001 variant count corrected 32→34 at v0.4.0; G4 AC-008 Then-clause split per-axis; G3 §B headline qualified with abbreviation-prefix residual); M-rows 30, Q-rows 14, expected RED 25 rows / 26 variants"
  - "run-gate #2: FAIL 0.89 (trajectory 0.79→0.83→0.88→0.89; G1-G5 all verified RESOLVED, corrected rule passed a full 44-cell trace) — fixes H1-H5 applied in v0.5.0 (H1 -l short list selector + variadic list-action rule + Q-15, -a <name> sibling git-rejected rc 128 fail-closed-safe; H3 cluster rule 'beginning with u' → 'containing any of dDmMcCftu' mid-cluster-u measured; H5 git -C <path> wrapper named residual/follow-up-card; H2 stale not-enumerated clause swept; H4 value-taking vs list-action selector wording de-conflated); M-rows 30, Q-rows 15, expected RED unchanged 25 rows / 26 variants; auditor cycle-dynamics note: H1/H3 are the last shapes git branch's grammar admits at this granularity — next optional-only round would be terminal PASS-WITH-DEBT"
  - "gate #3: FAIL (lead adjudicated STOP: option 1, one bounded final fix round; [HARD] gate #4 unconditional terminal, any finding outside spelling/arity/selector dimensions forces scope reduction) — ONLY measured L1-L3 applied in v0.6.0 (L1 --color/--abbrev/--column reclassified ATTACHED-ONLY optional-value: space-separated token is a creation operand, measured --color colprobe rc 0 created; wrong examples --column always/--abbrev 12 corrected; M-31/M-32 deny cells + Q-17 attached-form allow pin; error attribution: misclassification entered via auditor's own gate-#1 G1 prescription. L2 M-33 -vux mid-cluster-u cell (codex-measured completed mutation) + stale AC-002 cluster set swept to dDmMcCftu in spec §E + acceptance §D.2. L3 variadic rule widened: filter selectors --contains/--no-contains/--merged/--no-merged/--points-at also select list/filter mode, measured --contains HEAD main rc 0 output '* main'; Q-16 cell, closing an over-match regression the v0.5.0 H1 edit introduced); M-rows 33 (37 variants), Q-rows 17, expected RED 28 rows / 29 variants; no new grammar territory opened"
```

## §E.2 Run-phase Evidence

### M1 — synthetic measurement matrix (RED, pre-fix)

- **Command**: `go test ./internal/hook/ -run TestBranchGuardFlagClassMatrix -count=1 -v`
- **Tree SHA**: `38274782c9ab06e6adbd035a7b8a9a118cbe1932` (pre-fix tree, branch
  `WT-branchguard-flagclass`)
- **Exit**: FAIL (RED — the defect)
- **Verbatim output** (summary lines; per-cell FAIL lines below):

```
=== RUN   TestBranchGuardFlagClassMatrix
--- FAIL: TestBranchGuardFlagClassMatrix (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/E-2_git_branch_--set-upstream-to=origin/main_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/E-3a_git_branch_-F_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/E-3b_git_branch_--FORCE_topic_abc1234 (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/E-6_git_branch_-f_x_y_&&_git_status (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-06_git_branch_-f_topic_abc1234 (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-07_git_branch_-f_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-08_git_branch_--force_topic_abc1234 (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-09_git_branch_-df_old (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-10_git_branch_-fm_renamed (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-11_git_branch_-vD_old (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-12_git_branch_-u_origin/main_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-13_git_branch_--set-upstream-to=origin/main_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-14_git_branch_--set-upstream_origin/main_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-15_git_branch_--unset-upstream_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-16_git_branch_-t_topic_origin/main (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-17_git_branch_--track_topic_origin/main (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-18_git_branch_--no-track_topic_origin/main (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-19_git_branch_--edit-description_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-20_git_branch_--delete_old (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-21a_git_branch_--move_renamed (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-21b_git_branch_--copy_copied (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-22_git_branch_--set-upstream-to_origin/main_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-23_git_branch_--track=direct_topic_abc1234 (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-24_git_branch_-umain_topic (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-25_git_branch_-v_vbranch (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-26_git_branch_--_cr2 (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-27_git_branch_-q_qbranch (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-28_git_branch_--no-force_nfbranch (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-29_git_branch_-vt_vtbranch_main (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-30_git_branch_--create-reflog_crbranch (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-31_git_branch_--color_colprobe (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-32_git_branch_--abbrev_12_abbranch (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/M-33_git_branch_-vux_main_x (0.00s)
    --- FAIL: TestBranchGuardFlagClassMatrix/P-01b_git_branch_--force_topic_abc1234 (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.679s
FAIL
```

- **RED count**: 34 FAIL cells = §D.1's predicted 28 pre-fix allow M-rows
  (29 command variants: M-06..M-33, M-21 dual-form) + 5 expected-deny
  duplicate rows from the pair/edge tables (E-2 ≡ M-13, E-3a, E-3b, E-6,
  P-01b ≡ M-08). All 22 query cells (Q-01..Q-17 incl. Q-15a/Q-15b), the
  P-01a/P-02a-c whole-token allow cells, the E-5a/E-5b unknown-flag cells,
  and every pre-fix deny cell (M-01..M-05a/b, E-1, E-4) PASS (green-now).
- **§D.1 cross-read**: ZERO contradictions — every measured cell matches its
  §D.1 expectation (allow rows where §D.1 says the pre-fix guard allows,
  deny rows where §D.1 says it denies). No blocker per plan.md §D.
- **Fix-class decision**: the matrix confirms the under-match defect is
  flag-class-shaped (single-char class + "non-flag token" rule), not
  anchor-shaped — the M2 fix is the token-level classifier (plan.md §G
  corrected rule), replacing the `git branch` regex entry inside the
  `branchStatePatterns` set.

### Pre-flight baseline (Section C, recorded before M1)

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `golangci-lint run --timeout=2m` → 1 pre-existing finding
  (`internal/template/catalog_tree_hash.go:60:14` errcheck — outside this
  card's change surface; BASELINE, not NEW)
- `go test ./internal/hook/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 39.008s` (green baseline)
- B2 pre-scan: `Retired`/`superseded` markers exist only in agent-retirement
  machinery (`subagent_start.go`, `retired_events.go`, `audit_test.go`) — no
  cross-SPEC policy conflict on the branch-guard surface.

### M2 — matcher fix (GREEN)

- Commit `7d6f11687`: `git branch` regex entry replaced by the token-level
  flag-class classifier (`matchGitBranchMutation` +
  `classifyGitBranchTail` + `isGitBranchFlagAbbreviation`), slotted INTO
  `branchStatePatterns` as a predicate-matcher entry (struct gains an
  optional `match func(string) bool` field) so the M6 blanking tests keep
  their deny-origin semantics. All non-`branch` patterns byte-identical.
- **Command**: `go test ./internal/hook/ -run TestBranchGuardFlagClassMatrix -count=1`
- **Verbatim output**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.192s` (exit 0)
- Full package: `ok  github.com/modu-ai/moai-adk/internal/hook  68.269s` (exit 0) — all pre-existing pins green
- `golangci-lint run ./internal/hook/...` → `0 issues.`; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

### M3 — test completion + doctrine alignment

- **M3.1 end-to-end denial** (`TestBranchGuardFlagClass_EndToEndDenial`):
  11 commands through `checkBranchState` in a real primary fixture
  (`-f`, `--force`, `-df`, `-fm`, `-vD`, `-vux`, `-u`,
  `--set-upstream-to=`, `--unset-upstream`, `-t`, `--edit-description`)
  → all deny with `BRANCH_GUARD_VIOLATION: git branch` prefix. PASS.
- **M3.2 query-allowlist regression**
  (`TestBranchGuardFlagClass_QueryAllowlistEndToEnd`): `--merged`,
  `--no-merged`, `--points-at`, `--format`, `--sort committerdate`,
  `--list develop -v`, `--contains HEAD main` (Q-16), plus the N3 run-gate
  debt case `git branch -al develop` → all allow. PASS.
- **M3.3 condition tests** (`TestBranchGuardFlagClass_SubagentNegativePath`,
  D7): subagent-shaped `HookInput` (AgentType zero-valued, env unset) +
  `git branch -f x y` in a primary fixture → deny stands. Labeled in the
  test's doc comment as pinning guard LOGIC only, NOT the payload shape.
  REQ-WBG-F-008's two axes are encoded as documented conditions in the same
  comment block.
- **M3.4 D12 contested AgentType axis — capture IMPRACTICABLE (documented,
  axis left contested, no resolution fabricated)**:
  1. Nested `claude -p` probe (temp settings wiring a PreToolUse tee hook in
     `/tmp/t467-capture`, run with `--setting-sources none`): REFUSED twice
     by the runtime worktree-session guard with the verbatim reason
     "this command runs claude … in a plain command, so what it runs cannot
     be shown not to be git" (tried with and without a `cd /tmp` prefix).
     Nested claude is the only route to a real tool-spawned PreToolUse
     payload from this session (this agent has no Task/Agent tool).
  2. The hook's own logging (`internal/hook/trace`, `trace-*.jsonl`, 275
     PreToolUse rows incl. this session's subagent-context rows) records NO
     `agent_type`/`agent_id` field — structurally inconclusive for the axis.
  3. Per D12 and the coordinator's resume instruction: no capture
     CONTRADICTED the guard's reading, so no doc-reconciliation blocker
     fires; the axis REMAINS CONTESTED (`branch_guard.go:30-33` reading vs
     `hooks-system.md:114`). A future capture from a non-worktree session
     (or a lead-side probe) decides it; contradicting capture →
     doc-reconciliation blocker, never silent re-classification.
- **M3.5 doctrine pair (REQ-WBG-F-009 / AC-009)**: Query-vs-mutate bullet +
  forbidden-table row 2 flag enumeration rewritten to name the extended
  mutation class (shorts `-d|-D|-m|-M|-c|-C|-f|-u|-t`, longs
  `--force`/`--delete`/`--move`/`--copy`/`--set-upstream-to`/`--unset-upstream`/`--track`/`--edit-description`,
  combined clusters `-df`/`-vD`/`-vux`, positional/option-prefixed creation,
  operand-free query forms, `--dele`-abbreviation fail-open). Version
  1.3.2 → 1.3.3. Template copy edited FIRST
  (`internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`),
  local mirror in the SAME commit. Parity: `diff` shows the only delta is
  the local copy's two pre-existing SPEC-ID cross-ref lines (unchanged
  sanitized-pair convention); template copy carries 0 SPEC-ID/date/SHA
  tokens (`grep -c 'SPEC-\|t467\|20[0-9][0-9]-' → 0`).
  `go test ./internal/template/ -run Mirror` → `ok … 1.671s`.

### §E.2.1 Self-verification evidence (E1-E8, attribution triples)

All measurements on branch `WT-branchguard-flagclass`, post-M2/M3 tree
(final tree SHA = the M3 commit; see §E.3).

- **E1 AC matrix (acceptance.md §D.2)**:

| AC | Verdict | Evidence |
|----|---------|----------|
| AC-WBG-F-001 matrix convergence | **PASS** | `go test ./internal/hook/ -run TestBranchGuardFlagClassMatrix -count=1` → `ok … 1.192s`; 66/66 cells (33 M-rows/37 variants deny, 17 Q-rows allow, P-01/P-02 pairs, E-1..E-6) |
| AC-WBG-F-002 combined clusters | **PASS** | M-09/M-10/M-11/M-29/M-33 cells + end-to-end `-df`/`-fm`/`-vD`/`-vux` deny |
| AC-WBG-F-003 query allowlist regression | **PASS** | Q-01..Q-17 cells + end-to-end allow test; existing pins at `branch_guard_test.go` green unmodified |
| AC-WBG-F-004 whole-token classification | **PASS** | P-01a/P-01b, P-02a/c cells; `--format` allow vs `--force` deny; no embedded-letter denies |
| AC-WBG-F-005 preserved surfaces | **PASS** | `go test ./internal/hook/... -count=1` → all 11 packages `ok`; existing test files byte-untouched |
| AC-WBG-F-006 opt-in gate | **PASS** | `pre_tool_branch_guard_optin_test.go` green unmodified (in the full-package run) |
| AC-WBG-F-007 fail-open preserved | **PASS** | existing fail-open tests green unmodified; classifier adds no blocking path on uncertainty (E-5 arms) |
| AC-WBG-F-008 exemption-axes conditions | **PASS** (AgentType axis left contested per D12) | `TestBranchGuardFlagClass_SubagentNegativePath` PASS; documented conditions in its doc comment; capture impracticable — see M3.4 |
| AC-WBG-F-009 doctrine alignment | **PASS** | Template copy v1.3.3 + local mirror in the same commit; diff = pre-existing 2-line SPEC-ID delta only; `go test ./internal/template/ -run Mirror` → ok |

- **E2 builds**: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (both re-verified post-M3).
- **E3 coverage**: `go test -cover ./internal/hook/` →
  `ok  github.com/modu-ai/moai-adk/internal/hook  99.430s  coverage: 85.5% of statements`
  (≥ 85% package target; the matrix test drives every branch of
  `classifyGitBranchTail`/`matchGitBranchMutation`/`isGitBranchFlagAbbreviation`).
- **E4 subagent boundary**: `grep -rn 'AskUserQuestion' internal/hook | grep -v _test.go | grep -v '// '` → 1 match:
  `pre_tool.go:647: if input.ToolName == "AskUserQuestion"` — PRE-EXISTING on
  base tree `38274782c` (t401 judgment-first observer ToolName check, not an
  invocation; the file header itself asserts "No AskUserQuestion calls").
  NEW matches introduced by this card: 0.
- **E5 lint vs baseline**: `golangci-lint run --timeout=2m` → 1 finding =
  the pre-flight baseline finding (`internal/template/catalog_tree_hash.go:60`
  errcheck). NEW findings vs baseline: **none**. `golangci-lint run ./internal/hook/...` → `0 issues.`
- **E6 commits**: M1 `e6ea01064`, M2 `7d6f11687`, M3 = this commit
  (branch HEAD; ahead-count vs `origin/develop` recorded in §E.3). NO push
  (lane protocol).
- **E7 blockers**: none on the §D.1 axis (M1 measured matrix matched §D.1
  exactly; zero contradictions). One documented non-blocker: D12 capture
  impracticability (M3.4) — not a blocker because no capture contradicted
  the guard's reading; axis left contested per the no-fabrication rule.
- **E8 RED evidence**: M1 verbatim RED output + tree SHA `38274782c…`
  recorded above (this section, M1 block) BEFORE M2's green was accepted.

Known observation for the debt sweep: one earlier `go test ./internal/hook/...`
invocation (post-M3, output truncated to last 2 lines by `tail`) ended FAIL;
the immediate re-run and the confirmation runs were green across all 11
packages (`internal/hook` 74.5s/99.4s; subpackages ok). The truncated first
FAIL is recorded as unattributed-flaky (candidate: the load-sensitive
`perf` subpackage), NOT as evidence of a regression — the full green runs
are the cited verdicts.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: "M3 commit = final branch HEAD of WT-branchguard-flagclass (M1 e6ea01064, M2 7d6f11687; placeholder — backfill final HEAD from the completion report / git log)"
run_status: complete
milestones:
  M1: "matrix test + verbatim RED (34 cells = §D.1's 28 rows/29 variants + 5 expected-deny duplicate rows) on tree 38274782c — commit e6ea01064 (also: spec.md draft→in-progress)"
  M2: "token-level git branch flag-class classifier inside branchStatePatterns (predicate-matcher entry) — commit 7d6f11687; matrix 66/66 GREEN"
  M3: "end-to-end denial + query-allowlist + D7 negative-path condition tests; REQ-009 doctrine pair v1.3.3 (template-first, same commit); D12 capture impracticable → axis left contested — this commit"
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 6   # non-branch patterns ×5 regex entries + predicate entry; all byte-identical behavior
l44_pre_commit_fetch: not-applicable-lane-protocol   # no push; card branch only
l44_post_push_fetch: not-applicable-no-push
new_warnings_or_lints_introduced: 0   # baseline 1 pre-existing errcheck (catalog_tree_hash.go:60) unchanged
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
coverage_pkg_internal_hook: 85.5   # go test -cover, ≥85 target
total_run_phase_files: 4   # branch_guard.go, branch_guard_flagclass_test.go, doctrine pair (2 copies) — plus this SPEC dir's progress/spec frontmatter
m1_to_mN_commit_strategy: per-milestone Conventional Commits with card id t467 in body
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: ca64ccdac7a46a540d4815cfcc46a3c6bef7ec08   # D3 backfill — sync commit originally landed as 2d86aba52, trailer corrected (🗿 MoAI) to ca64ccdac pre-push; both unpushed
sync_status: completed
b12_self_test_a: pass   # grep -c 'SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001' CHANGELOG.md → 0 before emission (no duplicate entry)
b12_self_test_b: pass   # 9 distinct live AC identifiers in acceptance.md (AC-WBG-F-001..009; no [RETIRED]/[REF] exclusions) = CHANGELOG "Nine acceptance criteria PASS"
b12_self_test_c: pass   # implementation read directly, not from plan.md: internal/hook/branch_guard.go (matchGitBranchMutation + flag sets) + branch_guard_flagclass_test.go (4 test functions); paths verified by ls
changelog_entry_position: CHANGELOG.md [Unreleased] → ### Fixed — first entry of the Fixed block (SPEC-linked form per house pattern)
frontmatter_status_transitions:
  spec_md: in-progress -> completed   # rides the single sync commit (3-phase close; no separate Mx commit)
  plan_md: none
  acceptance_md: none
  progress_md: section-added          # §E.4 populated (this block)
canary_compliance_check:
  d12_contested_axis: remains contested — no resolution asserted in CHANGELOG or here; lead-side probe recipe at §E.2 M3.4
  push_state: not pushed — lane protocol; lead batch-push owns origin/develop
```

## §F Phase 4 Mode Selection

**Input parameters**: tier M (artifact set; scope discipline stays S) · scope 3 code/test files
(`internal/hook/branch_guard.go` + new `branch_guard_flagclass_test.go` + M3 doctrine pair) ·
domain count 1 (Go hook package; markdown pair inside M3) · file language mix Go + markdown ·
concurrency benefit LOW (coding-heavy, single package) · Agent Teams prereqs not requested.

**Mode evaluation**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | Non-trivial multi-milestone implementation, not a typo/single-line change |
| serial | **selected** | Coding-heavy TDD (M1 RED matrix → M2 GREEN fix → M3 tests), single package — Anthropic coding-task parallelism caveat |
| fanout | no | No multi-domain research; strictly sequential milestone dependencies |
| sweep | no | Not ≥~30-file mechanical uniform transform; no inter-file independence |

**Decision: serial**

**Justification**: the milestones are strictly sequential by design (M2 needs M1's measured
matrix; M3 needs M2's fix), all work lands in one package, and coding-heavy work is the
documented serial default. Parallel spawns would add coordination cost with zero concurrency
benefit.

**Gate record**: Implementation Kickoff Approval passed via the kanban lead's operator channel
(lead-1 dispatch 2026-09-03: "run 진입 — 운영자 게이트 통과"); lanes cannot open operator
gates, so approval authority rides the lead/operator exchange. Precondition completed by lead
order: origin/develop `4e4607abe` absorbed (merge commit `1f65efef2`, SPEC artifacts
unaffected). Phase 1 Plan Audit Gate re-execution armed: the artifact hash changed after the
iter-2 verdict (`5bd15b303` → `b8b48266f`, v0.3.0 fixes D8-D14), so the cached verdict is
invalid and the gate re-audits fresh — this gate verdict is the binding one before M1.
