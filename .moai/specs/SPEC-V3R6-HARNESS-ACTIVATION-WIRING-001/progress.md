# Progress — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
plan_complete_at: 2026-06-03
plan_status: audit-ready
plan_audit_verdict: PASS-WITH-DEBT 0.86 (GATE-2 approved)
plan_audit_remediation: D1/D2 (SHOULD-FIX) + D3/D4 (MINOR) all addressed in this commit
artifacts:
  - spec.md          # 12-field frontmatter + era:V3R6, GEARS REQ-HAW-001..016 (+013b), Exclusions EX-1..EX-8
  - plan.md          # Tier M justified, M1..M6 milestones, template-first ordering, D1 prefix disambiguation
  - acceptance.md    # AC-HAW-001..015 + AC-HAW-PROC-1..2, all grep/test-verifiable, DoD, edge cases
  - design.md        # wiring-mechanism decision (Option A recommended), main.md additive router, 3-check smoke gate
  - progress.md      # this file
authored_by: manager-spec
plan_commit_sha: aa078e1f0
```

### Plan-audit remediation log (2026-06-03)

| Defect | Severity | Resolution |
|--------|----------|------------|
| D1 | SHOULD-FIX | plan.md M4 + AC-HAW-PROC-2 disambiguated: §6.4 correction target is the code-side `my-harness-*` (NOT the EX-1 `harness-*` migration); cites `meta-harness SKILL.md:168` doctrine-vs-code drift; AC PROC-2 now asserts `my-harness-` equality + no bare `harness-*` directive (impossible to read as endorsing migration) |
| D2 | SHOULD-FIX | (preferred option) REQ-HAW-013b + AC-HAW-015 added: Phase-6 smoke gate FAILs when a generated agent OMITS `skills:` — runtime enforcement of REQ-HAW-008, closing the silent auto-discovery gap; spec.md REQ-HAW-008 binding updated; plan.md M5 + design.md §C updated |
| D3 | MINOR | design.md §B corrected: `mainMD()` already emits Domain Summary + Linked Files; only the `## Task-Shape Routing` table is additive (not a from-scratch rewrite) |
| D4 | MINOR | spec.md EX-1 future SPEC-ID `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` marked **(planned — not yet created)** |

## §E.0 Ground-Truth Diagnosis Anchors (verified during plan-phase)

- `InjectMarker` (`internal/harness/layer3.go`) — **0 non-test callers** (orphaned, but works + tested)
- `ScaffoldHarnessDir` (`internal/harness/layer5.go`, emits `main.md`) — **0 non-test callers** (orphaned)
- This repo CLAUDE.md + template CLAUDE.md — **0 `moai:harness-start` markers** each
- `project/meta-harness.md` — Phase 7 (5-Layer Activation) referenced but **body absent** (file ends at Phase 6.5)
- L4 import lines (`@.moai/harness/`) present in all of plan/run/sync/design workflows — L4 intact
- `doctor harness` L3 (marker) + L5 (`main.md`) checks already exist — smoke gate reuses them
- B3 (empty agent descriptions) REFUTED per diagnosis — codified as REQ-HAW-009 for the gate to assert

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: M (300-1000 LOC, 5-15 files)
- scope (file count): ~8-10 (2 Go source + 2 Go test + 2 template skill/workflow + 2 mirror sync)
- domain count: 2 — Go source code (`internal/cli/`, `internal/harness/`) + orchestration skill/workflow markdown (`.claude/skills/...`)
- file language mix: Go-dominant (the load-bearing wiring + smoke gate + TDD tests are Go); markdown is additive skill/workflow body
- concurrency benefit: LOW — coding-heavy, single coherent wiring chain (CLI command → InjectMarker/ScaffoldHarnessDir caller → smoke gate); milestones are sequential-dependent (M3 CLI before M4 router before M5 gate)
- Agent Teams prereqs status: not evaluated (single-domain coding work does not meet ≥3-domain gate)

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | not selected | Multi-file semantic change (new CLI command + smoke gate + tests) |
| 2 background | not selected | Write/Edit operations required (CONST-V3R2-020 forbids background writes) |
| 3 agent-team | not selected | Not multi-domain (≥3); Agent Teams capability gate not met |
| 4 parallel | not selected | Coding-heavy work — Finding A4 caveat (coding tasks have few parallelizable subtasks) |
| 5 sub-agent | **selected** | Default fallback for coding-heavy work; sequential milestone execution (M1→M6) with inter-milestone dependency |
| 6 workflow | not selected | Not ≥30-file mechanical-uniform transform; this is semantic new-code Go wiring |

**Decision: sub-agent**

**Justification**: This is coding-heavy Go work — the orphaned-installer wiring, the `moai harness install` CLI command, and the TDD smoke-gate extension form a single coherent dependency chain (the CLI command calls `InjectMarker`/`ScaffoldHarnessDir`; the smoke gate must follow the router restructure). Per Anthropic Finding A4, coding tasks involve fewer truly parallelizable subtasks than research, so the sequential sub-agent path (Mode 5) is the correct default. Mode 6 (workflow) is explicitly rejected: this is not a ≥30-file uniform mechanical transform but a small set of semantic new-code edits with inter-file dependency. Modes 3/4 are rejected because the scope is single-domain (Go + companion markdown) and below the multi-domain parallelism gate.

## §E.2 Run-phase Evidence

### M1 — RED baseline (ground truth captured 2026-06-03)

The four wiring breaks asserted as failing pre-conditions at run-phase start:

| Ground-truth assertion | Command | Result (RED) |
|------------------------|---------|--------------|
| `InjectMarker` orphaned | `grep -rn "InjectMarker(" --include="*.go" internal/ \| grep -v "_test.go" \| grep -v "func InjectMarker"` | **0 callers** (orphaned) |
| `ScaffoldHarnessDir` orphaned | `grep -rn "ScaffoldHarnessDir(" --include="*.go" internal/ \| grep -v "_test.go" \| grep -v "func ScaffoldHarnessDir"` | **0 callers** (orphaned) |
| CLAUDE.md markers absent | `grep -c "moai:harness-start" CLAUDE.md` | **0 markers** |
| Phase 7 body absent | `grep -n "Phase 7\|5-Layer Activation" project/meta-harness.md` | file ends at Phase 6.5 (only a forward reference at L306) |

Pre-flight baseline (Section C):
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `golangci-lint run ./internal/cli/... ./internal/harness/...` → 0 issues (clean baseline)
- existing `TestRunHarnessCheck*` / `TestInjectMarker*` / `TestScaffold*` → all PASS (L1-L5 + installers unit-tested but unwired)

### M2-M6 milestone summary

| Milestone | Commit | What changed |
|-----------|--------|--------------|
| M1 | `0ebed07b9` | RED baseline + Phase 0.95 Mode Selection; draft→in-progress |
| M2/M3 | `f45bc90d9` | Option A confirmed; `moai harness install` CLI (install.go + install_test.go) wiring InjectMarker+ScaffoldHarnessDir; harness_route.go registration; project/meta-harness.md Phase 7 body |
| M4 | `6becb8550` | layer5.go mainMD() `## Task-Shape Routing` table; meta-harness emission contract (skills: preload + non-empty description); §6.4 `moai-harness-*` → `my-harness-*` prefix correction |
| M5 | `2a90a28d7` | doctor_harness.go L6 smoke gate (checkLayer6AgentActivation) + frontmatter parser; 6 TDD cases |
| M6 | f9b54c07c | layer5.go readmeMD() Activation/Retrofit note; final verification batch + §E.2/E.3 population |

### Run-phase Evidence — AC PASS/FAIL matrix

| AC | Binds | Status | Verification | Actual Output |
|----|-------|--------|--------------|---------------|
| AC-HAW-001 | REQ-HAW-001 | PASS | `grep InjectMarker( ... -v _test.go -v func` | 1 caller `internal/cli/harness/install.go:85` (was 0) |
| AC-HAW-002 | REQ-HAW-002 | PASS | `TestRunInstall_Idempotent` (double install → 1 start/1 end/1 heading) | PASS |
| AC-HAW-003 | REQ-HAW-003 | PASS | `grep 'AskUserQuestion(' install.go/layer3.go/layer5.go/doctor_harness.go` | 0 invocation matches; `TestPropose_NoAskUserQuestion` PASS |
| AC-HAW-004 | REQ-HAW-004 | PASS | `TestRunInstall_MissingClaudeMd` (absent CLAUDE.md → wrapped error) | PASS |
| AC-HAW-005 | REQ-HAW-005 | PASS | `grep ScaffoldHarnessDir( ... -v _test.go -v func` | 1 caller `internal/cli/harness/install.go:74` (was 0) |
| AC-HAW-006 | REQ-HAW-006 | PASS | `TestScaffoldHarnessDir_MainMDIsRouterManifest` (Domain + Task-Shape Routing + Linked Files) | PASS |
| AC-HAW-007 | REQ-HAW-007 | PASS | `TestRunHarnessCheck_L5Missing` (agents present + main.md removed → L5 FAIL) | PASS (existing L5 check) |
| AC-HAW-008 | REQ-HAW-008,013b | PASS | `grep skills:` in SKILL.md + project workflow ≥1 | 8 matches; `TestRunHarnessCheck_MissingSkillsKey` PASS |
| AC-HAW-009 | REQ-HAW-009 | PASS | `TestRunHarnessCheck_EmptyAgentDescription` (empty desc → FAIL) | PASS |
| AC-HAW-010 | REQ-HAW-010 | PASS | `TestRunHarnessCheck_L5Missing` (main.md removed → FAIL) | PASS (L5) |
| AC-HAW-011 | REQ-HAW-011 | PASS | `TestRunHarnessCheck_L3MarkerUnpaired` (markers unpaired → L3 FAIL) | PASS (L3) |
| AC-HAW-012 | REQ-HAW-012 | PASS | `TestRunHarnessCheck_EmptyAgentDescription` detail names "description" + agent | PASS |
| AC-HAW-013 | REQ-HAW-013 | PASS | `TestRunHarnessCheck_DanglingSkillReference` (my-harness-nonexistent → FAIL) | PASS |
| AC-HAW-015 | REQ-HAW-008,013b | PASS | `TestRunHarnessCheck_MissingSkillsKey` (no skills: key → FAIL naming agent + skills) | PASS |
| AC-HAW-014 | REQ-HAW-014 | PASS | `grep 'L1:..L5:..runHarnessCheck' doctor_harness.go` = 12; all 6 existing L1-L5 cases PASS | PASS (L6 additive, L1-L5 preserved) |
| AC-HAW-PROC-1 | REQ-HAW-015 | PASS | template/working mirror byte-identity (meta-harness workflow + SKILL.md) | MIRROR OK both files; `make build` ran |
| AC-HAW-PROC-2 | REQ-HAW-016 | PASS | (a) `my-harness-[a-z]` ≥1 = 3 matches; (b) no bare `harness-*` directive = 0 | PASS (EX-1 boundary held) |

EC-4 (template `moai-*` skill ref not dangling): `TestRunHarnessCheck_TemplateSkillNotDangling` PASS.

### Invariant verification

| Invariant | Status | Evidence |
|-----------|--------|----------|
| All existing tests still pass | PASS-WITH-DEBT | `go test ./...` green EXCEPT `TestOutputStylesTemplateLiveParity` (einstein.md drift) — **pre-existing** at pre-M1 baseline `e83864047`, output-styles domain, NOT touched by this SPEC (EX-8 / B10 scope discipline) |
| InjectMarker/ScaffoldHarnessDir core logic unchanged (EX-6/D-4) | PASS | layer3.go byte-unchanged; layer5.go change confined to mainMD()/readmeMD() body content (no scaffolding-algorithm edit) |
| Prefix stays `my-harness-*` (D-2/EX-1) | PASS | no bare `harness-*` generation directive introduced; §6.4 corrected to `my-harness-*` |
| No new lint/type errors | PASS | `golangci-lint run ./internal/cli/... ./internal/harness/...` = 0 issues; `go vet` clean |
| Cross-platform build | PASS | `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
run_complete_at: 2026-06-03
run_commit_sha: f9b54c07c
run_status: implemented
ac_pass_count: 17        # AC-HAW-001..015 (013b folded into 015) + PROC-1..2
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list file modified outside scope
l44_pre_commit_fetch: n/a (worktree-isolated agent; orchestrator owns push)
l44_post_push_fetch: n/a (push deferred to orchestrator)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 10   # install.go install_test.go harness_route.go harness_retirement_test.go layer5.go layer5_test.go doctor_harness.go doctor_harness_test.go + 2 mirrored markdown (template+working ×2 files counted once each pair)
m1_to_mN_commit_strategy: per-milestone separate commits (M1 / M2+M3 / M4 / M5 / M6), local only — push coordinated by orchestrator post-run (active parallel session on shared branch)
known_preexisting_failure: TestOutputStylesTemplateLiveParity (einstein.md template/live drift) — present at pre-M1 baseline e83864047, output-styles domain, out of SPEC scope (EX-8)
authored_by: manager-develop
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
sync_complete_at: 2026-06-03
sync_commit_sha: 16df5274f
sync_status: implemented
ac_pass_count: 17        # AC-HAW-001..015 (013b folded into 015) + PROC-1..2
ac_fail_count: 0
version_transition: 0.1.0 -> 0.2.0
changelog_entry: CHANGELOG.md [Unreleased] Added — Harness activation wiring restoration
readme_touched: false        # internal-mechanism SPEC; no user-facing README API surface change
docs_site_touched: false     # internal wiring; not a docs-site user-doc topic
deliverables:
  - CHANGELOG.md (Added entry)
  - spec.md frontmatter (status in-progress -> implemented; version 0.1.0 -> 0.2.0)
  - progress.md (run_commit_sha backfill f9b54c07c + §E.4 population)
integration_method: cherry-pick (worktree base e83864047 -> patch∪run spec.md/progress.md auto-merge union; code 13 files disjoint)
run_commit_sha_backfilled: f9b54c07c   # integrated M6 SHA; cherry-pick reassigned from worktree tip c3a9872b4
integration_attributed_test_failures: 0   # proven via 8a6329b63..HEAD footprint (no internal/hook, no output-styles)
known_preexisting_failures:                # NOT integration-attributed
  - TestOutputStylesTemplateLiveParity (einstein.md template/live drift, output-styles domain, EX-8 / B10 out-of-scope)
  - TestHookWrapper_ValidJSON / TestHookWrapper_MoaiBinaryFallback (internal/hook flaky ~5s timeout under full-suite load; isolated re-run PASS 1.07s/0.88s; CI-FLAKY-STABILIZE-003 candidate)
authored_by: orchestrator-direct
```

### (Migrated from §E.5)

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
mx_complete_at: 2026-06-03
mx_commit_sha: 845bad045
final_status: completed
four_phase_close:
  plan: aa078e1f0 (plan-phase artifacts, 5 artifacts) + 532226a3c (plan-audit D1/D2/D3/D4 patch)
  run: 92ce0b1ce..f9b54c07c (M1-M6, 5 commits cherry-pick integrated; worktree origin 0ebed07b9..c3a9872b4)
  sync: 16df5274f
  mx: 845bad045
plan_audit_verdict: PASS-WITH-DEBT 0.86 (GATE-2 approved)
ac_pass_count: 17
ac_fail_count: 0
integration_attributed_test_failures: 0   # proven via 8a6329b63..HEAD footprint
known_preexisting_failures: TestOutputStylesTemplateLiveParity (einstein drift, EX-8) + TestHookWrapper flaky (CI-FLAKY-STABILIZE-003 candidate) — both NOT integration-attributed
close_method: orchestrator-direct (Tier M bounded internal SPEC; manager-docs sync failure-mode avoidance, active multi-session race)
authored_by: orchestrator-direct
```
