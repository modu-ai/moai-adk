# Progress — SPEC-AUDIT-PARTICIPANT-COUNT-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-01
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
requirements: 5
acceptance_criteria: 8
card: t284
premise_evidence: .moai/state/verify/t284/premise-probe.log
deferral_record: .moai/reports/t229/succession.md
```

Plan-phase notes carried forward for the run-phase owner:

- The shape decision and its four rejected alternatives are `plan.md` §D. The one item
  deliberately left open — whether `residual_risk_note` gains a sentence in the
  undetermined case — is a run-phase decision and must be recorded in §E.2 either way.
- Documentation surfaces are sync-phase scope (`plan.md` §G) and must not be edited
  during run phase.
- Audit-repair revision 0.2.0 (plan-audit iter 1 verdict FAIL 0.75, defects D1-D6;
  evidence `.moai/reports/t284/plan-audit-iter1.md`): REQ-APC-003 gained the
  intra-backend divergence carve-out (D2 — option (a), decided by the orchestrator and
  surfaced at the kickoff gate), AC-APC-005's enumeration re-partitioned by measured
  participant count (D1), `spec.md` §A.3's premise restated as derived-field identity
  (D3), site inventory corrected to 3 files / 12 sites (D4), §A.1/§A.4 coordinates
  corrected (D5/D6). AC count is now 8 (Tier M ceiling 16). Participant-count
  measurement: `.moai/reports/t284/participant-count-probe.log`.
- D7 debt discharged + D8 folded in (revision 0.2.1; plan-audit iter 2 PASS-WITH-DEBT
  0.85, `.moai/reports/t284/plan-audit-iter2.md`): AC-APC-002/003 Given clauses scoped
  to 0-or-1 inputs that produced no intra-backend synthesis divergence (AC-APC-008 owns
  the divergence input); AC-APC-005's coverage note and §D's option-(b) sentence
  aligned; `spec.md` §C case-#3 provenance named. Wording/scoping only.

## §F Phase 4 Mode Selection

Inputs: tier M; scope ~4 Go source files + 3 test files (12 existing assertion sites) in a single domain (`internal/cli`, Go); concurrency benefit LOW (coding-heavy).

Decision: serial

Rationale: coding-heavy single-package implementation falls to the sequential default (Anthropic coding-task parallelism caveat); one `manager-develop` TDD spawn carries the whole run. Operator approval for run entry received 2026-09-01 via lead (autonomous mode, D2 option (a) confirmed). Phase 1 gate disposition: iter2 verdict PASS-WITH-DEBT 0.85 (`.moai/reports/t284/plan-audit-iter2.md`) + the auditor-prescribed D7 debt discharge in v0.2.1 ("no iter3 needed"); the hash-invalidating discharge edit does not trigger a further audit re-run — the iter2 verdict chain is the audit record. Boundary note: the tree absorbed `origin/develop` 53a3fc1dd (card t248: `BuildCommit`/`BuildLag` fields on `ConvergenceResult`, `converge` shifted to :154) after the audit; the SPEC's §A coordinates were measured at 8c1d911df/64bba61aa and remain tree-pinned evidence — the implementer locates symbols by name.

## §E.2 Run-phase Evidence

Measured at branch `WT-audit-participant-count`, code HEAD `5713f82fa` (M3) unless a block
states its own tree. Full verbatim logs live in `.moai/reports/t284/` and are committed with
this evidence commit.

### Milestones as landed (reconciled against plan.md §E)

| Milestone | Commit | Shape |
|---|---|---|
| M1 (data model + derivation + DQ-2) | `a67dcb472` | **atomic with M2** — the `bool → *bool` narrowing breaks compilation of every `DisagreementFlag` reader, so M1 without M2 cannot produce a compiling tree; one green commit carries both |
| M2 (existing call-site repair) | folded into `a67dcb472` | **14 sites, not the inventoried 12** — see Findings F-1 |
| M3 (8 acceptance criteria) | `5713f82fa` | `internal/cli/mcp_convergence_participant_test.go` |
| M4 (documentation) | not in run phase | plan.md §E: "Deferred to sync phase; scope in §G" — no M4 commit is correct |

### E1 — AC matrix (deciding command: `go test ./internal/cli/ -run 'AC_APC' -count=1 -v`,
exit 0, 8 PASS / 0 FAIL; log `.moai/reports/t284/e1-ac-matrix.log`)

| AC | Status | Deciding test | Verbatim output |
|----|--------|---------------|-----------------|
| AC-APC-001 | PASS | `TestConverge_ParticipantCount_Table_AC_APC_001` (rows a–h) | `--- PASS: TestConverge_ParticipantCount_Table_AC_APC_001 (0.00s)` |
| AC-APC-002 | PASS | `TestConverge_BelowTwo_NoDivergence_FlagNilNotFalse_AC_APC_002` (5 sub-2 inputs + DQ-2; direct nil assertion) | `--- PASS: TestConverge_BelowTwo_NoDivergence_FlagNilNotFalse_AC_APC_002 (0.00s)` |
| AC-APC-003 | PASS | `TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003` (member presence + nil value + raw-bytes both checked) | `--- PASS: TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003 (0.00s)` |
| AC-APC-004 | PASS | `TestConverge_TwoMeasuredInputs_DerivedSummaryDiffers_AC_APC_004` | `--- PASS: TestConverge_TwoMeasuredInputs_DerivedSummaryDiffers_AC_APC_004 (0.00s)` |
| AC-APC-005 | PASS | `TestConverge_TwoOrMore_BooleanUnchanged_AC_APC_005` (six measured ≥2 cases with counts) | `--- PASS: TestConverge_TwoOrMore_BooleanUnchanged_AC_APC_005 (0.00s)` |
| AC-APC-006 | PASS | `TestConverge_Undetermined_GatesNothing_AC_APC_006` (real persist→gate round trip; block path still only overall==fail) | `--- PASS: TestConverge_Undetermined_GatesNothing_AC_APC_006 (0.00s)` |
| AC-APC-007 | PASS | `TestLoadConvergenceResult_OldStateFile_Decodes_AC_APC_007` (hand-written old-shape fixture) | `--- PASS: TestLoadConvergenceResult_OldStateFile_Decodes_AC_APC_007 (0.00s)` |
| AC-APC-008 | PASS | `TestConverge_SingleParticipantDivergence_CarveOut_AC_APC_008` + existing `TestConverge_SurfacesSignalDivergence_WithoutBlocking` (kept green through the carve-out) | `--- PASS: TestConverge_SingleParticipantDivergence_CarveOut_AC_APC_008 (0.00s)` |

Invariants (REQ-APC-005 / C2 / C3 / C5, covered by the same run + existing suite): gate ALLOW on
undetermined, block only on overall==fail (AC-APC-006 asserts both directions); no new enum
(`TestNoVerdictDisagreementEnum_EC7_AC_AMM_011` green); engine AskUserQuestion boundary
(`TestConvergence_NoAskUserQuestion_AC_AMM_024` green). Affected-package suite:
`go test -count=1 ./internal/cli/...` → exit 0 (log `green-full-internal-cli.log`, measured
pre-commit on the identical content).

### E8 — TDD RED evidence (three observation blocks)

**Block 1 — runtime RED, pre-GREEN, tree `53a3fc1dd` + new test file** (command
`go test ./internal/cli/ -run '...AC_APC_003|...AC_APC_004' -count=1`, exit 1; full log
`red-stage1-runtime.log`). Map-based tests only, so they compile against the unchanged engine
and fail on flag/byte semantics:

```text
--- FAIL: TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003/claude_only_(spec_§A.2_case_3) (0.00s)
    mcp_convergence_participant_test.go:67: disagreement_flag = false (bool), want JSON null — a sub-2 no-divergence result is undetermined, not false (bytes: {..."disagreement_flag":false,...})
    mcp_convergence_participant_test.go:70: raw bytes contain "disagreement_flag":false; forbidden below 2 participants
--- FAIL: TestConverge_TwoMeasuredInputs_DerivedSummaryDiffers_AC_APC_004 (0.00s)
    mcp_convergence_participant_test.go:92: three-way agreement participant_count = <nil>, want 3
    mcp_convergence_participant_test.go:95: claude-only participant_count = <nil>, want 1
    mcp_convergence_participant_test.go:101: claude-only disagreement_flag = false (bool), want JSON null
    mcp_convergence_participant_test.go:106: participant_count identical across the two inputs (<nil>); the derived summary must distinguish them
    mcp_convergence_participant_test.go:109: disagreement_flag identical across the two inputs (false); the derived summary must distinguish them
```

**Block 2 — compile RED, pre-GREEN, same tree** (command
`go test ./internal/cli/ -run 'AC_APC' -count=1`, exit 1; full log `red-stage2-compile.log`).
Pointer/count tests added against the unchanged engine:

```text
internal/cli/mcp_convergence_participant_test.go:164:9: r.ParticipantCount undefined (type ConvergenceResult has no field or method ParticipantCount)
internal/cli/mcp_convergence_participant_test.go:173:28: invalid operation: r.DisagreementFlag != nil (mismatched types bool and untyped nil)
internal/cli/mcp_convergence_participant_test.go:174:68: invalid operation: cannot indirect r.DisagreementFlag (variable of type bool)
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**Block 3 — mutant discharge, tree `5713f82fa`** (representative mutant of plan.md §F /
acceptance.md §D, applied by hand to `converge` Step 2c: sub-2 unconditionally emits a non-nil
pointer to `false`, carve-out dropped; command
`go test ./internal/cli/ -run 'AC_APC_001|AC_APC_002|AC_APC_003|AC_APC_008' -count=1`, exit 1;
full log `mutant-red.log`):

```text
--- FAIL: TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003 (0.00s)
    mcp_convergence_participant_test.go:74: disagreement_flag = false (bool), want JSON null — a sub-2 no-divergence result is undetermined, not false
--- FAIL: TestConverge_BelowTwo_NoDivergence_FlagNilNotFalse_AC_APC_002/claude_only_(spec_§A.2_case_3) (0.00s)
    mcp_convergence_participant_test.go:216: disagreement_flag non-nil (points at false); want nil — below 2 participants false is a claim one participant cannot ground (REQ-APC-003)
--- FAIL: TestConverge_SingleParticipantDivergence_CarveOut_AC_APC_008 (0.00s)
    mcp_convergence_participant_test.go:423: disagreement_flag = non-nil false; want non-nil true — the carve-out keeps a directly-observed divergence, the one sub-2 case where null is forbidden
```

Witness table verified exactly as acceptance.md §D records it: AC-APC-002/003/008 red (the
three witnesses); the AC-APC-001 a–h count rows green under the mutant ("the mutant counts
correctly"). One extra red: the §B empty-slice edge folded as a subtest inside the AC-APC-001
test function also asserts the nil flag, so that FUNCTION goes red while the criterion's own
rows stay green — additional discharge surface from folding §B into the tables, not a
contradiction of §D. Revert: `git checkout -- internal/cli/mcp_convergence.go`; `git diff`
empty immediately after (observation: `git diff --stat` printed nothing, status showed only
the card's untracked artifacts). Post-revert the 8 AC tests re-ran green (E1 above).

### residual_risk_note decision (plan.md §D, "Deliberately not decided here")

**Decided: NO undetermined-case sentence — the note surface is unchanged.** Reasons: (1) the
undetermined state is already legible in-band through both new-field positions — AC-APC-004's
witness proves the derived summary now distinguishes the cases, so a note sentence would be a
third surface for one fact; (2) a note would perturb note expectations beyond the two movements
this SPEC specifies (plan.md §H risk row treats any further movement as a finding to report);
(3) `blockReason` echoes the note only on the `overall_verdict == fail` path, so an
undetermined-case note could never reach the gate — its only reader already has the two
positions. No note assertion anywhere moved (full suite green with the note untouched).

### Findings (reported, not silently absorbed)

- **F-1 — call-site inventory was 12, measured 14.** The two extras are in
  `mcp_build_identity_test.go` (absorbed t248 tree, absent from plan §B's inventory measured at
  `8c1d911df`/`64bba61aa`): `:411` a `DisagreementFlag: false` literal (compile-level, same
  class as the 12) and `:423` the AC-ABI-004 **always-present key-set assertion**, which
  `participant_count` legitimately joins (REQ-APC-001 wants a visible 0 — deliberately NOT
  omitempty). Repair: literal → `boolPtr(false)`; key set += `"participant_count"` with a
  comment naming why the key set change is the SPEC's intent, not drift.
- **F-2 — E4 grep nuance.** The delegation's quick filter
  (`grep -rn 'AskUserQuestion' internal/cli --include='*.go' | grep -v _test | grep -v '// '`)
  returns 16 pre-existing hits, ALL comment lines or string literals in files this SPEC did
  not touch (the filter misses block-comment continuations and strings). Zero hits in
  `mcp_convergence.go` / `mcp_audit_multi.go` / `multi_review_gate.go`. The comment-aware
  mechanical guards all pass: 26 `*NoAskUserQuestion` tests green including
  `TestConvergence_NoAskUserQuestion_AC_AMM_024` (log: E1 run).

### E2/E3/E5 measurements (all at HEAD `5713f82fa`)

- E2 `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E3 `go test -cover ./internal/cli/...` → exit 0; root package **80.1%** (profile total 80.2%,
  16,004 statements; log + profile `e3-coverage.log` / `e3-cover.out`). Package figure is a
  pre-existing property — this SPEC adds ~20 statements (all covered), bounding its package
  impact at ≤0.12pp in the positive direction. **Modified-file coverage**: `converge` 100%,
  `countParticipants` 100%, `runMultiAudit` 100%, `HandleMultiReviewGate` 100% — every function
  this SPEC added or changed is fully covered.
- E5 `golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` exit 0 — identical to
  the Section C baseline (also `0 issues.`), **0 NEW issues**.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: "5713f82fa"
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 3   # plan §G protected groups (docs-site/ .claude/skills/ internal/template/templates/) — git diff 53a3fc1dd..HEAD over those paths = 0 files
l44_pre_commit_fetch: "omitted before M1 (not in the delegation pre-flight); measured post-hoc 2026-09-02 after fetch: origin/develop...HEAD = 45 2 — origin advanced 45 (parallel lanes), card ahead 2 (M1+M3); absorption belongs to the integration window (lead-owned)"
l44_post_push_fetch: "not applicable — lanes do not push card branches; zero pushes executed"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 12   # M1 (5 code + 3 SPEC artifacts + 3 plan-audit records) + M3 (1 test file); the evidence commit adds progress.md + 4 run-phase logs
m1_to_mN_commit_strategy: "M1+M2 atomic (compile-coupled: struct narrowing breaks every reader) -> M3 new criteria -> evidence chore commit last; M4 documentation is sync-phase scope per plan.md §E/§G"
coverage_note: "package root 80.1% (pre-existing level; SPEC delta bounded ≤ +0.12pp, statements added all covered); SPEC-touched functions 100%"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
