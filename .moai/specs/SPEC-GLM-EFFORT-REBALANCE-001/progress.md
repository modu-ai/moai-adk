# SPEC-GLM-EFFORT-REBALANCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M. Artifacts emitted: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
- Scope was settled with the operator before authoring; no `[NEEDS CLARIFICATION]` markers remain.
- Coupled-file inventory verified against the tree during planning, not carried over from the delegation brief. Three corrections and four additions are recorded in `plan.md` §B.
- Requirements: REQ-GER-001 through REQ-GER-014 (GEARS). Criteria: AC-GER-001 through AC-GER-016, plus six items named in `acceptance.md` §F as not mechanically checkable.
- Amendment 1 (config shadow): surfaced by the coordinator from the primary checkout, which a worktree cannot see. `llm.profiles` is consulted before the Go default, so a Go-matrix-only edit would have been inert on every populated install. Added REQ-GER-012 / REQ-GER-013, milestone M0, and AC-GER-006 / AC-GER-013 / AC-GER-014. Corrected the `plan.md` §B-1 reason: the local `llm.yaml` is gitignored runtime state, not an absent file.
- Amendment 2 (plan-audit FAIL 0.64, Testability 0.50 — six criteria could not decide what they claimed). Resolved: two `internal/cli` test breakages invisible to the package set (D1); the false "sole consumer" claim, which hid a second main-session delivery path (D2); an unowned template `profiles:` block (D7); a false `llm.glm` premise (D8); an AC-GER-008 grep vacuous on the two `agent-authoring.md` surfaces it was meant to protect (D9); AC-GER-010 verifying nothing because `runModelProfile` never reads the embed (D4); AC-GER-013 written for the wrong YAML shape (D6); and the build step not refreshing the binary the criteria execute (D5). Added REQ-GER-014, AC-GER-015, AC-GER-016, and §B.0 Freshness gate.
- Status: `draft`. Awaiting re-audit and Implementation Kickoff Approval.

### Open gap for sync — MP-3 is PASS-by-inspection, not PASS-by-tool

`moai spec lint` exceeded the auditor's 120s window and never completed, so this SPEC's frontmatter validity (MP-3) was cleared by reading the 12 canonical fields, not by running the linter. Per `verification-claim-integrity.md` §1.1 surface 3, an inspection is a hypothesis where a dedicated tool exists. **The sync phase must re-run `moai spec lint` (with an extended timeout) against this SPEC and record the verbatim result** before treating frontmatter validity as verified.

### Audit-history note — D3 was real, then cleared by an unrelated session

The plan audit recorded `go vet ./...` and `GOOS=windows go vet ./...` both exiting 1 on `internal/cli/preference/home_isolation_test.go:75` (`undefined: userHomeDir`), which would have made AC-GER-009a/009b unsatisfiable. A re-run in the primary checkout after the audit exited 0 with zero output, and both `home_isolation_test.go` files no longer exist — a parallel session deleted them mid-audit. The auditor was not wrong; the tree moved underneath it.

The recommendation was kept anyway and AC-GER-009a/009b are now baseline-relative rather than absolute-exit-0. Three sessions share this checkout, so an absolute assertion is hostage to unrelated churn in both directions: it can fail correct work, and it can pass by luck.

## §E.2 Run-phase Evidence

### Baseline gate (captured BEFORE the first edit, at HEAD `7b7772638`)

| Artifact | Command | Result |
|---|---|---|
| `baseline-vet-host.txt` | `go vet ./... > …/baseline-vet-host.txt 2>&1` | exit 0, **0 lines** (clean) |
| `baseline-vet-windows.txt` | `GOOS=windows go vet ./... > …/baseline-vet-windows.txt 2>&1` | exit 0, **0 lines** (clean) |
| `baseline-llm.yaml` | `cp .moai/config/sections/llm.yaml …/baseline-llm.yaml` | 5121 bytes copied from the primary checkout |

Pre-edit resolver reading (primary checkout, config present) — the before half of
the AC-GER-006 delta:

```
manager-spec opus/max
plan-auditor opus/max
manager-docs opus/high
```

All three baseline files are committed so the paths cited below still resolve at
audit time; they are working artifacts and may be removed at sync.

### Acceptance matrix

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-GER-001 | PASS | the two `grep -nE` patterns over `internal/template/profile_matrix.go` | both exit 1 (zero matches) |
| AC-GER-002 | PASS | `go test ./internal/template/... -run 'TestDefaultProfileMatrix_(Shape\|Monotone)' -count=1` | `--- PASS: TestDefaultProfileMatrix_Shape`, `--- PASS: TestDefaultProfileMatrix_Monotone`, `ok` |
| AC-GER-003 | PASS | `go test ./internal/template/... -run 'TestSessionGLMReasoningState' -count=1` | `--- PASS: TestSessionGLMReasoningState`, `--- PASS: TestSessionGLMReasoningStateForEffort` |
| AC-GER-004 | PASS | `moai agent lint \| grep -E 'LR-12.*(manager-spec\|plan-auditor\|manager-docs)'` + `grep -H '^effort:'` on the six files | grep exit 1 (zero matches); `moai agent lint` exit 0 (`25 total (0 errors, 25 warnings)` — all LR-08 skill-preload drift, pre-existing); frontmatter reads `high`/`high`/`low` in both trees |
| AC-GER-005 | PASS | `grep -nE '^\s+(synthesize\|research):'` on the shipped `llm.yaml` + `go test -run TestResolveHarnessAgentModelEffort` | high/medium columns: `synthesize { effort: low }`, `research { effort: high }`; low column unchanged; test `ok` |
| AC-GER-006 | PASS | `moai model profile --json \| jq …` run from the **primary checkout**, config present (5123 bytes), binary `b72b6ab01` | `manager-spec opus/high` / `plan-auditor opus/high` / `manager-docs opus/low` (profile=medium, backend=claude) |
| AC-GER-013 | PASS | `grep -A2` on the local block-style file + `grep` on the flow-style template + harness test | local: high col research `high` / synthesize `low` (L124-129), medium col research `high` / synthesize `low` (L168-173), low col unchanged (L146-151); template L142/143 + L150/151 correct; test `ok` |
| AC-GER-014 | PASS | `diff …/baseline-llm.yaml .moai/config/sections/llm.yaml` | exactly 10 changed lines — the 6 `profiles` cells + the 4 `harness_agents` cells. `profile`, `performance_tier`, `team_mode`, `glm_env_var`, `mode`, the whole `glm:` block and `agent_overrides` are byte-unchanged |
| AC-GER-007 | PASS-WITH-DEBT | `go test ./internal/template/... ./internal/cli/... ./internal/web/... -count=1` | exit 1; 18 packages `ok`; the sole failure is `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey`, which reproduces **identically on the pre-change baseline tree** (`git archive 7b7772638` → same `expected BLOCK …; got decision="" err=<nil>`). Live-codex environment dependency, not introduced here |
| AC-GER-008 | PASS | grep (a) over both `model-policy.md`, grep (b) over both `agent-authoring.md` | (a) exit 1 — every pre-change phrasing is gone; (b) `manager-spec \| high`, `plan-auditor \| high`, `manager-docs \| low` in both trees. Required states verified line by line: 122 `max` on manager-develop + super-advisor; 123 no `max`; 131 verified-not-rewritten (now true); 196 "Two matrix cells"; 198 no longer says plan takes `max` |
| AC-GER-009a | PASS | `go vet ./...` + `diff` vs baseline | exit 0, diff empty — no new finding |
| AC-GER-009b | PASS | `GOOS=windows go vet ./...` + `diff` vs baseline | exit 0, diff empty — no new finding (test layer compiles) |
| AC-GER-010 | PASS | `go test -run 'TestEmbeddedLLMYAMLMatchesMatrix' -v \| grep -c '^--- PASS:'` | `--- PASS: TestEmbeddedLLMYAMLMatchesMatrix`, count 1 — the test ran, not `[no tests to run]` |
| AC-GER-011 | PASS | `go test -run TestRuleTemplateMirrorDrift` + three frontmatter `diff`s + `cmp` on both rule files | test `ok`; all three diffs empty; `model-policy.md` and `agent-authoring.md` byte-identical across trees |
| AC-GER-012 | PASS | commit-scoped grep for `SPEC-GLM-EFFORT-REBALANCE\|REQ-GER-\|AC-GER-` over added template lines + tree-wide grep + `TestTemplateNoInternalContentLeak` | zero matches for tokens introduced by this change (the tree-wide hits are pre-existing `SPEC-AUTH-001` examples and `SPEC-AGENT-ARCH-V2-001` citations in files this change also touched); leak test `ok` |
| AC-GER-015 | PASS | `go test ./internal/cli/ -run 'TestGLMReasoningEnvVars\|TestGLMReasoningEnvVarsForEffort' -v` | `--- PASS: TestGLMReasoningEnvVars_SessionHigh` (sub-agent wire = `high`) and all six `TestGLMReasoningEnvVarsForEffort` sub-cases including `empty_→_session_default_(reasoning_high)` |
| AC-GER-016 | PASS | `sed -n '/^    high:/,/^    low:/p' … \| grep -E '(manager-spec\|plan-auditor\|manager-docs):'` | six lines: `opus/high`, `opus/high`, `opus/low` under each of the high and medium columns |
| §E DoD 3 | PASS | report-content check | neither the commit messages nor this report claims a GLM cost reduction, and both state that the matrix change does not alter delivered GLM behavior (REQ-GER-011) |

### Guard falsifiability probe (AC-GER-010)

`TestEmbeddedLLMYAMLMatchesMatrix` was written after the template edit, so it never
saw a natural RED. Its discriminating power was demonstrated instead: the shipped
`manager-docs` cell was temporarily reverted to `effort: high` and the test failed
with `high/manager-docs: embedded llm.yaml has {Model:opus Effort:high},
DefaultProfileMatrix() has {Model:opus Effort:low}` on both affected columns. The
probe was restored from a byte-verified backup (`cmp` → identical) and the test
returned to green.

### Surfaces the plan did not enumerate (found during run)

Two additional assertion surfaces broke and were repaired in the same scope
envelope; both are one-value updates of the same mechanical class as `plan.md` §B-3:

- `internal/cli/agentlint/agent_lint_test.go` — `TestLintLR12_MatrixDrift_CleanAgent`
  seeds a synthetic `manager-spec` with `effort: max` and asserts no LR-12 drift.
  The `internal/cli/agentlint` package is a **sibling** of the `./internal/cli/...`
  parent that §B-3 correctly flagged, so it was inside the tested package set but
  outside the enumerated file list.
- `internal/cli/profile_setup_translations.go` + `internal/cli/wizard/{questions,translations}.go`
  — the 4-locale tier labels state the matrix opus effort span, asserted by
  `TestModelPolicyLabels_AgreeWithProfileMatrix`. High moved `(max~medium)` →
  `(max~low)`; medium moved `(max~low)` → `(high~low)`. The wizard copies carry the
  same spans and were updated for consistency (no test asserts them today).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-14
run_commit_sha: "763582247"          # M1 implementation commit
run_followup_commit_sha: "b72b6ab01" # gofmt of the one edited test file
run_status: complete-with-debt
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt_count: 1           # AC-GER-007 (pre-existing live-codex failure)
preserve_list_post_run_count: 0      # no PRESERVE-listed file modified
l44_pre_commit_fetch: not-run        # worktree branch, no push performed
l44_post_push_fetch: not-run         # no push performed (out of scope by instruction)
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues; vet diff empty on both platforms
cross_platform_build:
  host_vet: "exit 0, diff vs baseline empty"
  windows_vet: "exit 0, diff vs baseline empty"
total_run_phase_files: 24            # 23 modified + 1 new test file
m1_to_mN_commit_strategy: "single M1 implementation commit + one style follow-up; no push"
```

### Open items carried to sync

- `moai spec lint` was still not run against this SPEC (the plan-phase MP-3 gap
  above is unchanged) — the sync phase owes that verbatim result.
- The live-codex gate failure is pre-existing and unrelated; it is recorded here so
  a sync-phase reader does not attribute it to this change.
- Other machines carrying a pre-change `profiles:` block in their gitignored
  `llm.yaml` keep resolving the old cells. Known limitation per `spec.md` §4; no
  migrator ships here.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
