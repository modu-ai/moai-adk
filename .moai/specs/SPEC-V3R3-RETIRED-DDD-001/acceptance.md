# SPEC-V3R3-RETIRED-DDD-001 Acceptance Criteria (Phase 1B)

> Given/When/Then scenarios for manager-ddd retired stub standardization acceptance.
> Companion to `spec.md` v0.3.0, `plan.md` v0.3.0, `research.md` v0.3.0.
> 4 ACs cover 11 non-deferred REQs (REQ-RD-011 deferred per spec.md §5.4).

## HISTORY

| Version | Date       | Author     | Description                                                                                                                                                                                                                                                                                                       |
|---------|------------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | Goos Kim   | 4 G/W/T scenarios + edge cases for SPEC-V3R3-RETIRED-DDD-001 (positive + edge + boundary + negative)                                                                                                                                                                                                              |
| 0.2.0   | 2026-05-04 | Goos Kim   | Audit iter 1 defects D3/D7 fix: AC-RD-04 violation messages cite REQ-RD-NNN (not predecessor REQ-RA-NNN); AC-RD-01 byte-size criterion replaced with tolerance band 1300–1700 bytes (D7); REQ 순서 sequential REQ-RD-001..012; file count taxonomy consistent with research.md §6.5 (Cat A 30 / Cat B 3 / Cat C 2). |
| 0.3.0   | 2026-05-04 | manager-spec, iter 3 atomic sync | 5 artifacts version bumped to current + companion reference 갱신. Acceptance.md 자체에는 기능 변경 없음 (이미 iter 2에서 D3/D7 적용됨); cross-artifact 동기화 위해 버전만 업데이트.                                                                                                                                                              |

---

## 1. Acceptance Criteria

### AC-RD-01 (Positive): manager-ddd retirement frontmatter conformance (REQ-RD-001, REQ-RD-002, REQ-RD-003, REQ-RD-008, REQ-RD-012)

**Given** a clean checkout of the moai-adk-go repository at branch `feature/SPEC-V3R3-RETIRED-DDD-001` after M2 implementation
**And** the predecessor SPEC-V3R3-RETIRED-AGENT-001 PR #776 (commit `20d77d931`) has been merged into main
**And** `internal/template/templates/.claude/agents/moai/manager-cycle.md` exists with full frontmatter and active definition (verified at session start, 11385 bytes)

**When** the developer runs `go test ./internal/template/ -run "TestAgentFrontmatterAudit" -v`
**And** also runs `go test ./internal/template/ -run "TestRetirementCompletenessAssertion" -v`

**Then** the test output shows:
- `TestAgentFrontmatterAudit/.claude/agents/moai/manager-ddd.md` PASSES (generic walk loop)
  - Frontmatter has all 5 retirement fields:
    - `retired: true` (boolean, NOT string `"true"`)
    - `retired_replacement: manager-cycle` (exact match to active replacement file basename)
    - `retired_param_hint: "cycle_type=ddd"` (string, parameter invocation hint)
    - `tools: []` (explicit empty YAML array)
    - `skills: []` (explicit empty YAML array)
  - Legacy `status: retired` field is ABSENT
- `TestAgentFrontmatterAudit/manager-ddd must be retired` subtest PASSES (added in M1, RED triggered before M2, GREEN after M2)
- `TestRetirementCompletenessAssertion/all retired agents have replacement in embedded FS` PASSES (manager-ddd → manager-cycle pair validated by generic loop)
- `TestRetirementCompletenessAssertion/manager-ddd replacement manager-cycle must exist` subtest PASSES (defensive symmetry)

**And** when the developer reads `internal/template/templates/.claude/agents/moai/manager-ddd.md`:
- File size is between **1300 and 1700 bytes** (audit D7 fix iter 2: tolerance band replaces previous "approximately 1500 bytes" non-binary criterion). Reference: predecessor manager-tdd retired stub is 1392 bytes; original manager-ddd is 7628 bytes. The tolerance band reflects natural variance from minor wording differences while ensuring the file is unambiguously a retired stub (not a partial rewrite).
- Body contains 5 H2 sections (mirror manager-tdd retired stub structure):
  - `# manager-ddd — Retired Agent` (H1)
  - `## Replacement` — names manager-cycle with cycle_type=ddd
  - `## Migration Guide` — table with Old Invocation | New Invocation rows
  - `## Why This Change` — cites both SPEC-V3R3-RETIRED-AGENT-001 (predecessor) and SPEC-V3R3-RETIRED-DDD-001 (current)
  - `## Active Agent` — pointer `.claude/agents/moai/manager-cycle.md`

**Edge cases**:
- File size outside 1300–1700 byte band: REJECTED. Below 1300: likely missing required H2 sections; above 1700: likely retains active-agent body content.
- File size at exactly 1500 bytes: PASS (within band).
- Frontmatter `retired: "true"` (string): REJECTED by audit — must be boolean (audit emits `RETIREMENT_INCOMPLETE_manager-ddd`)
- Frontmatter `tools:` (no value, just key): REJECTED — must be explicit `[]`
- `retired_replacement: manager_cycle` (underscore typo): REJECTED — must match real file basename `manager-cycle`
- Body content lacks Migration Guide table: REJECTED by REQ-RD-008 — `## Migration Guide` H2 section MUST contain a Markdown table
- Body content uses different H2 ordering than manager-tdd: WARNING by manual review (M2 commit) — structural mirror is preferred but not enforced by automated test
- `retired_param_hint: cycle_type=ddd` (no quotes): REJECTED — research.md §2.2 schema requires quoted string

**Test Anchor**:
- `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit` (path-loop subtest)
- `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit/manager-ddd must be retired` (M1-added subtest)
- `internal/template/agent_frontmatter_audit_test.go TestRetirementCompletenessAssertion/all retired agents have replacement in embedded FS` (generic loop)
- `internal/template/agent_frontmatter_audit_test.go TestRetirementCompletenessAssertion/manager-ddd replacement manager-cycle must exist` (M1-added subtest, defensive)
- File size assertion: `len(data) >= 1300 && len(data) <= 1700` in M2 verification step

---

### AC-RD-02 (Positive + Edge): SubagentStart runtime guard regression check (REQ-RD-004, REQ-RD-005, REQ-RD-009)

**Given** the M2 implementation deployed (manager-ddd.md frontmatter `retired: true` + replacement `manager-cycle`)
**And** the predecessor `agentStartHandler` (in `internal/hook/subagent_start.go`) is unchanged from PR #776 merge state
**And** the project structure has `.claude/agents/moai/manager-ddd.md` with the standardized retired stub frontmatter (deployed to a test project via `moai update`)

**When** Claude Code triggers a SubagentStart hook event with stdin JSON `{"agentName": "manager-ddd", "agent_id": "<uuid>", "session": {...}}`
**And** `agentStartHandler.Handle()` is invoked

**Then** the handler returns within ≤500ms (predecessor REQ-RA-012 budget; observed 0.056ms)
**And** the response is `&HookOutput{Decision: DecisionBlock, Reason: "agent manager-ddd retired (SPEC-V3R3-RETIRED-AGENT-001), use manager-cycle with cycle_type=ddd"}`
**And** Claude Code surfaces the block reason to the orchestrator
**And** the spawn is blocked at the runtime layer (worktree allocation does NOT proceed)
**And** the orchestrator can re-route the invocation to manager-cycle with cycle_type=ddd

**Edge cases**:
- Hook fires after worktree allocation (Claude Code spec ambiguity per predecessor research.md §4.1): predecessor research verified SubagentStart timing; if Claude Code semantics shift, behavior degrades to "block-with-message-after-allocation" — still safer than no block
- Multiple `retired: true` agents in same session (manager-ddd + manager-tdd both invoked): each hook invocation is independent; both blocked correctly
- Stdin malformed JSON: handler returns exit 0 (allow, fail-safe); does NOT crash. (Predecessor research.md §10.3 verified)
- `agentName` is empty string: handler pass-through (REQ-RA-008 generic semantic, inherited by REQ-RD-005 by extension)
- `agentName` contains path traversal (`..`, `/`): handler rejects with warning log, returns pass-through (predecessor implementation)
- Performance regression: if Handle() exceeds 500ms (e.g., due to large frontmatter file), audit-test does not directly measure. Manual benchmark via `go test -bench` (deferred to M5 manual smoke)

**Test Anchor**:
- Existing `internal/hook/agent_start_test.go TestAgentStartBlocksRetiredAgent` (predecessor-added; covers any retired:true agent generically)
- Existing `internal/hook/agent_start_test.go TestAgentStartHandlerPerformance` (predecessor-added; verifies ≤500ms budget)
- Manual integration test in M5 smoke test phase (`/tmp/test-ddd-retire` simulation)

---

### AC-RD-03 (Boundary): Backward compatibility preservation via factory.go case "ddd" (REQ-RD-006)

**Given** the M4 implementation deployed (factory.go switch-level @MX:NOTE expanded; no code change)
**And** legacy user projects exist that have NOT yet run `moai update` (e.g., a test installation at `/tmp/legacy-project` with the older active manager-ddd.md frontmatter still containing `hooks:` block referencing `ddd-pre-transformation` etc.)

**When** Claude Code in the legacy project fires a hook event with action string `ddd-pre-transformation` (legacy hook still configured)
**And** `internal/hook/agents/factory.go.CreateHandler("ddd-pre-transformation")` is invoked

**Then** the factory returns `NewDDDHandler("pre-transformation")` (the existing DDD handler in `internal/hook/agents/ddd_handler.go`)
**And** the handler executes its existing logic (no behavioral regression for legacy users)
**And** the legacy user can continue working with their pre-update setup (verified via existing factory_test.go tests; "behavioral regression" defined as DDD handler test failures)
**And** when the user eventually runs `moai update`, manager-ddd.md is replaced with the retired stub (`hooks:` block removed) and SubagentStart hook (AC-RD-02) blocks future spawns

**And** when the developer reads the switch statement comment in `internal/hook/agents/factory.go` lines 19–28:
- Comment cites both SPEC-V3R3-RETIRED-AGENT-001 (predecessor) and SPEC-V3R3-RETIRED-DDD-001 (current) for `case "tdd":` and `case "ddd":` preservation rationale
- @MX:NOTE explicitly states: "preserved for backward compatibility with legacy user projects that have not run `moai update`"

**Edge cases**:
- User invokes `moai update` mid-session: switch case behavior unchanged; only manager-ddd.md frontmatter is updated; future SubagentStart events block via AC-RD-02
- Future cleanup SPEC removes `case "ddd":`: legacy users must update first; CHANGELOG warning documented in spec.md §10
- `factory.go` `case "ddd"` accidentally removed during this SPEC's implementation: M4 verification step `go test ./internal/hook/agents/ -run "TestFactory"` catches this — `factory_test.go` lines 200, 229, 231 reference `NewDDDHandler` and would fail
- `ddd_handler.go` accidentally deleted: `go build ./internal/hook/...` fails; build error blocks merge

**Test Anchor**:
- `internal/hook/agents/factory_test.go TestFactory*` (existing tests, lines 200, 229, 231 cover NewDDDHandler dispatch)
- Manual M4 verification: `go test ./internal/hook/agents/ -run "TestFactory" -v` shows DDD handler tests pass

---

### AC-RD-04 (Negative): CI assertion failure semantics for incomplete retirement (REQ-RD-002, REQ-RD-007, REQ-RD-010, REQ-RD-012)

**Given** the M1 implementation deployed (extended `agent_frontmatter_audit_test.go`)
**And** an incomplete retirement state where one of the following violations exists (testing each branch independently):
- (a) `manager-ddd.md` frontmatter is missing `retired_replacement` field
- (b) `manager-ddd.md` frontmatter has legacy `status: retired` field (no `retired: true` boolean)
- (c) `manager-ddd.md` frontmatter has `tools:` key without explicit `[]` (malformed retirement)
- (d) Some Cat A substitution-target file (e.g., `agents/moai/expert-backend.md`) still contains a `manager-ddd` substring
- (e) `manager-cycle.md` is accidentally deleted from the template

**When** the developer (or CI pipeline) runs `go test ./internal/template/ -run "TestAgentFrontmatterAudit|TestRetirementCompletenessAssertion|TestNoOrphanedManagerDDDReference" -v`

**Then** the failing test output emits the appropriate sentinel (audit D3 fix iter 2: REQ-RD-NNN cited, NOT predecessor REQ-RA-NNN, since these are this SPEC's own assertions targeting manager-ddd):
- For violation (a): `RETIREMENT_INCOMPLETE_manager-ddd: retired:true 에이전트에 'retired_replacement' 필드 없음 (REQ-RD-002)`
- For violation (b): `RETIREMENT_INCOMPLETE: legacy 'status: retired' 필드 감지. 'retired: true' boolean 필드로 교체 필요 (REQ-RD-002)`
- For violation (c): `RETIREMENT_INCOMPLETE_manager-ddd: retired:true 에이전트에 'tools: []' 명시적 빈 배열 없음 (REQ-RD-002)`
- For violation (d): `ORPHANED_MANAGER_DDD_REFERENCE in agents/moai/expert-backend.md: <line>: <line content>. SPEC-V3R3-RETIRED-DDD-001 M3에서 'manager-cycle'로 교체 필요 (REQ-RD-010)`
- For violation (e): `RETIREMENT_INCOMPLETE_manager-ddd: retired_replacement 'manager-cycle' 파일이 embedded FS에 없음 (.claude/agents/moai/manager-cycle.md)` (generic loop) AND `RETIREMENT_INCOMPLETE_manager-ddd: 교체 에이전트 '.claude/agents/moai/manager-cycle.md'가 embedded FS에 없음. SPEC-V3R3-RETIRED-DDD-001 M2에서 manager-cycle.md 검증 필요 (REQ-RD-012)` (defensive subtest)
- The CI pipeline FAILS (exit non-zero), blocking the PR merge

> **Note on REQ-RD-NNN citation**: The implementation in `agent_frontmatter_audit_test.go` is shared infrastructure introduced by predecessor SPEC-V3R3-RETIRED-AGENT-001 (so the helper's predecessor-era error messages may still cite REQ-RA-002 for legacy reasons). However, this SPEC's M1 added subtests and new top-level function MUST emit messages citing `REQ-RD-002`, `REQ-RD-010`, or `REQ-RD-012` as appropriate. Specifically: the manager-ddd-specific subtests (line 139–161 mirror in M1, defensive subtest in `TestRetirementCompletenessAssertion`) and `TestNoOrphanedManagerDDDReference` MUST cite REQ-RD-NNN. Only the generic walk loop subtests (which existed pre-this-SPEC) may emit pre-existing REQ-RA-NNN strings — those are predecessor-managed.

**And** the failing tests do not produce false positives for the predecessor SPEC-V3R3-RETIRED-AGENT-001 manager-tdd assertions (regression-clean: existing manager-tdd subtests continue PASS independently)

**Edge cases**:
- Multiple violations simultaneously (e.g., (a) AND (d)): each test reports independently; both sentinels emitted; CI sees ≥2 failed subtests
- Cat B file (manager-cycle.md migration table phrase `manager-tdd and manager-ddd`): NOT in `TestNoOrphanedManagerDDDReference.checkFiles` slice (Cat B excluded entirely per research.md §6.6); helper allow-list does NOT need to handle this
- Cat C file (`agent-hooks.md` line 48 `manager-ddd | ddd-pre-transformation | ...`): NOT in checkFiles slice; substring preservation acceptable
- Allow-list rule for `# deprecated` headers: defensive allow (not expected in Cat A files)
- Allow-list rule for HTML comments `<!-- ... -->`: defensive allow (general markdown convention)
- Test runs with `-count=1` flag: results consistent (no test caching of stale embedded FS); `make build` precondition required
- File path mismatch in checkFiles slice: M1 implementation MUST verify all 30 paths exist via `fs.Stat`; if path missing, subtest SKIPS with "make build 필요" message (not FAIL)
- Frontmatter parse error: M1 fail-safe is `t.Fatalf` (predecessor pattern preserved); does not silently mask retirement issues

**Test Anchor**:
- `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit` (frontmatter validation, walk loop)
- `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit/manager-ddd must be retired` (M1-added explicit subtest; cites REQ-RD-002)
- `internal/template/agent_frontmatter_audit_test.go TestRetirementCompletenessAssertion/all retired agents have replacement in embedded FS` (generic loop)
- `internal/template/agent_frontmatter_audit_test.go TestRetirementCompletenessAssertion/manager-ddd replacement manager-cycle must exist` (M1-added defensive subtest; cites REQ-RD-012)
- `internal/template/agent_frontmatter_audit_test.go TestNoOrphanedManagerDDDReference` (M1-added new top-level test, **30 Cat A file scope**; cites REQ-RD-010)
- `internal/template/agent_frontmatter_audit_test.go findManagerDDDReferences` (M1-added helper, allow-list)

---

## 2. Quality Gate Criteria (Definition of Done)

A SPEC-V3R3-RETIRED-DDD-001 Run phase is COMPLETE when ALL of the following PASS:

| # | Criterion | Verification Method |
|---|-----------|---------------------|
| 1 | manager-ddd.md retired stub conforms to 5-field frontmatter schema | `TestAgentFrontmatterAudit` PASSES |
| 2 | manager-ddd retirement explicit subtests PASS | `TestAgentFrontmatterAudit/manager-ddd must be retired` PASSES |
| 3 | manager-cycle replacement validated for manager-ddd | `TestRetirementCompletenessAssertion` (both generic loop + defensive subtest) PASSES |
| 4 | All **30 Cat A substitution-target files** free of orphan `manager-ddd` references | `TestNoOrphanedManagerDDDReference` ALL 30 subtests PASS |
| 5 | factory.go `case "ddd":` preserved (backward compat) | `factory_test.go TestFactory*` continues PASSES |
| 6 | factory.go switch-level @MX:NOTE expanded with SPEC-V3R3-RETIRED-DDD-001 citation | Manual diff review of `internal/hook/agents/factory.go` lines 19–28 |
| 7 | Cat C1 `agent-hooks.md` @MX:NOTE added (manager-ddd substring preserved) | Manual diff review; `grep -c "manager-ddd" rules/moai/core/agent-hooks.md` returns 2 |
| 8 | Cat C2 `handle-agent-hook.sh.tmpl` @MX:NOTE added | Manual diff review; `bash -n <rendered>` syntax check passes |
| 9 | CHANGELOG.md Unreleased entry added with `moai update` user action directive | Manual review of CHANGELOG.md head |
| 10 | All M1 RED tests transition to GREEN after M2-M4 | Final `go test ./internal/template/ -run "TestAgentFrontmatterAudit\|TestRetirementCompletenessAssertion\|TestNoOrphanedManagerDDDReference" -v` shows all subtests PASS |
| 11 | No predecessor manager-tdd test regressions | All pre-existing tests in `agent_frontmatter_audit_test.go` PASS |
| 12 | Full repository test suite PASS | `go test ./...` returns exit 0 |
| 13 | No race conditions | `go test -race ./...` PASS |
| 14 | Lint clean | `golangci-lint run ./...` 0 issues |
| 15 | Embedded template build clean | `make build` succeeds, `git diff internal/template/embedded.go` shows expected changes |
| 16 | Local install succeeds | `make install` produces binary, `moai version` reflects new version |
| 17 | Manual smoke test passes | M5 `/tmp/test-ddd-retire` simulation: manager-ddd retired stub deploys; SubagentStart guard blocks invocation |

All 17 criteria MUST PASS before PR merge to main.

---

## 3. REQ → AC Coverage Verification

REQ numbering: sequential REQ-RD-001..REQ-RD-012 (audit D1 fix iter 2; no gaps).

| REQ ID | Category | Requirement Summary | AC Coverage | Verification |
|---|---|---|---|---|
| REQ-RD-001 | Ubiquitous | Standardized retirement frontmatter | AC-RD-01 | TestAgentFrontmatterAudit walk loop |
| REQ-RD-002 | Ubiquitous | manager-ddd audit assertions (3 sub-deliverables) | AC-RD-01, AC-RD-04 | Explicit subtest + new top-level test function |
| REQ-RD-003 | Ubiquitous | Embedded FS regeneration | AC-RD-01 | TestRetirementCompletenessAssertion generic loop |
| REQ-RD-004 | Ubiquitous | Predecessor agentStartHandler covers manager-ddd | AC-RD-02 | Manual smoke test (`agentStartHandler` is generic; M5 verification) |
| REQ-RD-005 | Event-Driven | SubagentStart block decision for manager-ddd | AC-RD-02 | Existing `TestAgentStartBlocksRetiredAgent` covers any retired:true agent |
| REQ-RD-006 | Event-Driven | factory.go case "ddd" backward compat | AC-RD-03 | Existing `TestFactory*` + manual code review |
| REQ-RD-007 | Event-Driven | Test sentinel emission | AC-RD-04 | TestNoOrphanedManagerDDDReference emits `ORPHANED_MANAGER_DDD_REFERENCE` on violation |
| REQ-RD-008 | State-Driven | Retired stub body structure | AC-RD-01 | Manual diff review (M2) confirms 5 H2 sections |
| REQ-RD-009 | State-Driven | ≤500ms guard performance | AC-RD-02 | Existing `TestAgentStartHandlerPerformance` covers manager-ddd via generic logic |
| REQ-RD-010 | State-Driven | 30 Cat A file documentation substitution + Cat B/C disposition taxonomy | AC-RD-04 | TestNoOrphanedManagerDDDReference + grep verification |
| REQ-RD-011 | Optional | Optional `moai agents list --retired` | (deferred) | spec.md §5.4 explicit deferral |
| REQ-RD-012 | Unwanted (composite) | CI assertion composite | AC-RD-01, AC-RD-04 | Multiple sentinel branches |

Coverage: **11/12 REQs mapped, REQ-RD-011 explicitly deferred**, 100% non-deferred coverage.

---

## 4. Cross-Test Boundary Considerations

### 4.1 Independence from Predecessor Tests

The M1-added tests MUST NOT modify or duplicate predecessor test functions. Specifically:
- Do NOT modify `TestAgentFrontmatterAudit/manager-tdd must be retired` subtest (lines 139–161)
- Do NOT modify `TestRetirementCompletenessAssertion/manager-tdd replacement manager-cycle must exist` subtest (lines 179–188)
- Do NOT modify `TestNoOrphanedManagerTDDReference` top-level function (lines 235–298)
- Do NOT modify `findManagerTDDReferences` helper (lines 315–338)

All M1 additions are NEW code: 2 new subtests + 1 new top-level function + 1 new helper. Total ~118 lines added (~457 lines after addition).

### 4.2 KEEP/UPDATE Allow-List Boundary

The `findManagerDDDReferences` helper MUST exclude these patterns from orphan detection. Sources from research.md §6.3 (Cat B) + §6.4 (Cat C) + §6.6 (allow-list):

| Pattern | Reason | Source |
|---------|--------|--------|
| `# deprecated` (markdown headers) | Defensive allow (not expected in Cat A files) | predecessor pattern |
| `<!--` (HTML comment open) | Defensive allow (general markdown convention) | predecessor pattern |

Cat B (3 files) and Cat C (2 files) are EXCLUDED from `checkFiles` slice entirely (research.md §6.6), so no allow-list entry is required for:
- `manager-ddd.md` frontmatter `name:` (Cat B1; not in checkFiles)
- `manager-cycle.md` migration table at lines 61/65/70 (Cat B2; not in checkFiles)
- `manager-tdd.md` body line 31 (Cat B3; not in checkFiles)
- `agent-hooks.md` lines 48, 79 (Cat C1; not in checkFiles)
- `handle-agent-hook.sh.tmpl` (Cat C2; not in checkFiles AND has 0 manager-ddd substrings)
- `manager-tdd and manager-ddd` migration phrase (only appears in Cat B2 manager-cycle.md, which is excluded)
- `ddd-pre-transformation` etc. hook actions (only appear in Cat C2 file, excluded)

If false positive detected during M3 substitution: ADD allow-list entry to helper; do NOT remove orphan reference from substitution targets.

### 4.3 Pre-merge Boundary Tests

Before PR creation, manual verification:

```bash
# Substitution count (post-M3+M4)
grep -rln "manager-ddd" internal/template/templates/ | wc -l
# Expected: 4 files (Cat B1 manager-ddd.md frontmatter + Cat B2 manager-cycle.md + Cat B3 manager-tdd.md + Cat C1 agent-hooks.md)

# Sentinel detection (post-M3)
go test ./internal/template/ -run "TestNoOrphanedManagerDDDReference" -v
# Expected: ALL 30 Cat A subtests PASS

# Predecessor regression check
go test ./internal/template/ -run "TestNoOrphanedManagerTDDReference" -v
# Expected: ALL subtests PASS (no regression)

# factory.go regression check
go test ./internal/hook/agents/ -run "TestFactory" -v
# Expected: ALL DDD + TDD + cycle handler tests PASS

# Predecessor agentStartHandler regression check
go test ./internal/hook/ -run "TestAgentStart" -v
# Expected: ALL retired-rejection guard tests PASS (covers manager-ddd via generic logic)

# manager-ddd.md byte-size tolerance check (M2 verification)
wc -c internal/template/templates/.claude/agents/moai/manager-ddd.md
# Expected: between 1300 and 1700 bytes (audit D7 fix tolerance band)
```

---

End of acceptance.md.
