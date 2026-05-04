# SPEC-V3R3-RETIRED-DDD-001 Acceptance Criteria (Phase 1B)

> Given/When/Then scenarios for manager-ddd retired stub standardization acceptance.
> Companion to `spec.md` v0.3.0 and `plan.md` v0.3.0.
> All 4 ACs (Positive + Edge + Boundary + Negative) cover the 12 REQs (1-to-1 mapping confirmed in plan.md §1.4 — 11/12 explicitly mapped, REQ-RD-011 deferred per predecessor REQ-RA-014).

## HISTORY

| Version | Date       | Author              | Description                                                                                |
|---------|------------|---------------------|--------------------------------------------------------------------------------------------|
| 0.3.0   | 2026-05-04 | MoAI Plan Workflow  | Force-accept (iter 3). 4 G/W/T scenarios finalized covering Positive + Edge + Boundary + Negative |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow  | iter 2 regression fix — cross-artifact partial sync resolved                                 |
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 4 G/W/T scenarios for SPEC-V3R3-RETIRED-DDD-001                                              |

---

## 1. Acceptance Criteria

### AC-RD-01 (Positive scenario): manager-ddd retired stub + 30 substitutions + embedded.go regenerated end-to-end

Mapped REQs: REQ-RD-001, REQ-RD-002, REQ-RD-006

**Given** a clean checkout of moai-adk-go at branch `feature/SPEC-V3R3-RETIRED-DDD-001`, base commit `origin/main` `20d77d931` (SPEC-V3R3-RETIRED-AGENT-001 / PR #776 머지된 상태)

**And** M1 (audit row added → RED), M2 (Cat B B-01 REWRITE → GREEN), M3 (Cat A 30 SUBSTITUTE → GREEN) milestones implementations are completed in sequence

**When** the developer runs:
1. `cat internal/template/templates/.claude/agents/moai/manager-ddd.md`
2. `grep -rln "manager-ddd" internal/template/templates/.claude/agents/moai/ internal/template/templates/.claude/rules/moai/ internal/template/templates/.claude/skills/ internal/template/templates/.claude/output-styles/ internal/template/templates/CLAUDE.md`
3. `make build`
4. `git diff --stat internal/template/embedded.go`

**Then** the following all hold:

- (1) `manager-ddd.md` is a retired stub of size approximately 1.4KB (`wc -c` returns a value between 1300 and 1500 bytes), with frontmatter containing exactly:
  - `name: manager-ddd`
  - `retired: true` (boolean)
  - `retired_replacement: manager-cycle`
  - `retired_param_hint: "cycle_type=ddd"`
  - `tools: []`
  - `skills: []`
- (1) The body contains a Migration Guide table with old → new invocation patterns specific to DDD (ANALYZE-PRESERVE-IMPROVE cycle).
- (2) The grep returns hits ONLY in:
  - `internal/template/templates/.claude/agents/moai/manager-ddd.md` (retired stub self-reference: `name: manager-ddd`, body migration table)
  - `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` (Cat C C-01 — annotated row)
  - `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (Cat C C-02 — annotated listing)
- (2) The grep returns NO hits in any of the 30 Cat A SUBSTITUTE files (per plan.md §3 manifest).
- (3) `make build` exits 0 with no errors.
- (4) `git diff --stat internal/template/embedded.go` shows non-empty changes covering all edits.

**Edge cases**:
- `manager-ddd` 식별자가 `ddd-pre-transformation` 같은 hyphen-suffixed action key의 prefix로 등장하는 경우 — 이는 substitute 대상이 아님 (action key 보존). grep 결과에 `ddd-` 접두 식별자가 hit하면 별도 확인.
- `manager-ddd.md` self-references (frontmatter `name: manager-ddd` + migration table 옛 호출 예시) — substitution 대상에서 제외.

**Test Anchor**: `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit` (manager-ddd row) + manual grep verification.

---

### AC-RD-02 (Edge scenario): Cat C UPDATE-WITH-ANNOTATION preserves action-key dispatch + retired marker visible + retire stub body matches predecessor format

Mapped REQs: REQ-RD-003, REQ-RD-008, REQ-RD-009

**Given** M2 (Cat B REWRITE) and M4 (Cat C UPDATE-WITH-ANNOTATION) milestones are completed

**When** the developer reads:
1. `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` § Agent Hook Actions table
2. `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` § Manager Agents listing
3. `internal/template/templates/.claude/agents/moai/manager-ddd.md` body (post-REWRITE)
4. `internal/template/templates/.claude/agents/moai/manager-tdd.md` body (predecessor reference)

**Then**:
- (1) The Agent Hook Actions table preserves the manager-ddd row containing action keys `ddd-pre-transformation`, `ddd-post-transformation`, `ddd-completion` (REQ-RD-007 invariant — backward compat with `factory.go` `ddd_handler.go` dispatch).
- (1) The manager-ddd row label includes a retired marker (e.g., `manager-ddd (retired — see manager-cycle)`).
- (1) An explanatory note exists below the table clarifying the dispatch path is preserved only for backward compat with hook events; new spawns are blocked at SubagentStart per predecessor REQ-RA-007.
- (2) The Manager Agents listing reflects `manager-cycle` as the unified DDD/TDD agent (Option A from plan.md M4.7); `manager-ddd` is removed from the active list OR retained with explicit retired marker (consistent with predecessor's manager-tdd treatment).
- (3) The new `manager-ddd.md` body contains, in order: `# manager-ddd — Retired Agent` heading, Replacement section pointing to manager-cycle with `cycle_type=ddd`, Migration Guide table (≥2 rows showing old → new invocation pairs), Why This Change section, Active Agent reference.
- (4) The structure of (3) matches the predecessor `manager-tdd.md` body (1392 bytes) form-for-form: same heading levels, same section ordering, same migration table column count.

**Edge cases**:
- `agent-hooks.md` table has manager-ddd row deleted entirely (rather than annotated) → REJECTED. `factory.go` `ddd_handler.go` dispatch would become orphan documentation.
- `agent-authoring.md` Manager Agents count "8" becomes inconsistent → ACCEPTED if accompanied by explicit count adjustment AND consistent with predecessor's manager-tdd treatment ("8" = effective active managers including manager-cycle).
- Migration table missing `cycle_type=ddd` parameter column → REJECTED (REQ-RD-008 requires explicit migration command).

**Test Anchor**: Manual review by reviewer + `grep "manager-ddd" internal/template/templates/.claude/rules/` returning only the 2 Cat C files with annotations.

---

### AC-RD-03 (Boundary scenario): audit test fails fast on any frontmatter regression + RETIREMENT_INCOMPLETE_manager-ddd assertion

Mapped REQs: REQ-RD-004, REQ-RD-010, REQ-RD-012

**Given** M1 + M2 (audit row added + Cat B REWRITE → audit GREEN) milestones are completed

**And** the test table in `internal/template/agent_frontmatter_audit_test.go` contains both manager-tdd (predecessor) and manager-ddd (this SPEC) rows with identical schema

**When** the developer intentionally introduces ONE of the following regressions and runs `go test ./internal/template/ -run TestAgentFrontmatterAudit -v`:

- **Regression A**: Edit `manager-ddd.md` frontmatter to set `retired: "true"` (string instead of boolean).
- **Regression B**: Edit `manager-ddd.md` frontmatter to set `retired_replacement: cycle-manager` (typo).
- **Regression C**: Edit `manager-ddd.md` frontmatter to remove `tools: []` field entirely.
- **Regression D**: Edit `manager-ddd.md` frontmatter to set `retired_param_hint: ""` (empty string).
- **Regression E**: Add `manager-ddd` reference back to one Cat A file (e.g., `internal/template/templates/CLAUDE.md` §4 Manager Agents listing) — if extended audit also detects cross-file references.

**Then** the test fails for each regression with a clear message, distinguishable per failure mode:
- A → `expected boolean, got string` style error
- B → `expected_replacement mismatch: want manager-cycle, got cycle-manager`
- C → `tools field absent or non-empty`
- D → `retired_param_hint must be non-empty`
- E → (only if extended) `RETIREMENT_INCOMPLETE_manager-ddd: cross-reference still cites manager-ddd in CLAUDE.md`

**And** the test exit code is non-zero, blocking CI.

**And** when each regression is reverted, the test passes again.

**Edge cases**:
- Audit test schema diverges from predecessor (e.g., predecessor uses `expected_tools_empty bool`, this SPEC accidentally adds `expected_tools_count int`) → REJECTED. plan.md M1.1 mandates schema reuse.
- Regression E is out of scope if the predecessor audit only checks frontmatter, not cross-references — that case falls to M3.3 grep verification (manual gate).

**Test Anchor**: `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit` with deliberate regression injection.

---

### AC-RD-04 (Negative scenario): SubagentStart guard blocks manager-ddd spawn ≤500ms; no worktreePath leak; factory.go ddd_handler dispatch path intact

Mapped REQs: REQ-RD-005, REQ-RD-007, REQ-RD-012

**Given** all milestones M1-M5 are completed and the SPEC is merged to main via PR

**And** a downstream user has run `moai update` and synced the standardized `manager-ddd.md` retired stub to their `.claude/agents/moai/manager-ddd.md`

**When** the user (or automation) attempts an `Agent({subagent_type: "manager-ddd", isolation: "worktree", ...})` invocation

**Then**:
- (a) The SubagentStart hook (predecessor REQ-RA-007 / agent_start.go from PR #776) parses `manager-ddd.md` frontmatter, detects `retired: true`, extracts `retired_replacement: manager-cycle` and `retired_param_hint: "cycle_type=ddd"`.
- (b) The hook emits stdout JSON `{"decision":"block","reason":"agent manager-ddd retired (SPEC-V3R3-RETIRED-DDD-001), use manager-cycle with cycle_type=ddd"}` and exits with code 2.
- (c) Total handler latency is ≤500ms (predecessor REQ-RA-012 inheritance — single YAML parse + stdout write, no network).
- (d) NO worktree is allocated for manager-ddd. The Agent() return value, if any, does NOT include a `worktreePath: {}` empty object that would propagate to fallback re-delegation.
- (e) The mo.ai.kr-class incident path (`/Users/.../{}/{}` literal) does NOT recur, because the spawn is blocked before worktree allocation.
- (f) Separately, `internal/hook/agents/factory.go` continues to dispatch hook events with action keys `ddd-pre-transformation`, `ddd-post-transformation`, `ddd-completion` to `ddd_handler.go` for backward compatibility (REQ-RD-007 invariant). This dispatch path is unaffected by the SubagentStart guard, which only fires on agent spawn events, not on PreToolUse/PostToolUse/SubagentStop hook events.

**Verification commands**:
1. Trigger spawn attempt: e.g., `claude /agent manager-ddd "test"` or scripted Agent() invocation.
2. Inspect hook log: `cat .claude/hooks/log/subagent-start.log` (or equivalent) for the block decision.
3. Time the response: `time <spawn-attempt>`. Expect ≤500ms.
4. Verify factory.go: `grep -A5 "manager-ddd\|ddd_handler" internal/hook/agents/factory.go` shows dispatch routing intact.

**Edge cases**:
- User's local `manager-ddd.md` was modified pre-update (e.g., they removed `retired: true` to "revive" the agent locally). In that case, SubagentStart guard does NOT block (correct — guard respects user override). Audit test in template only catches template state, not user state. Documented as expected behavior.
- Hook event dispatched by `factory.go` for legacy `ddd-completion` action key continues to call `ddd_handler.go.Handle()`. This is correct (REQ-RD-007 invariant). The handler may be a stub if predecessor or this SPEC simplified it, but dispatch routing must remain.
- Latency >500ms due to slow YAML parser on large frontmatter — REJECTED for this SPEC; predecessor REQ-RA-012 already constrained the implementation.

**Test Anchor**: Integration test in `internal/hook/agent_start_test.go` (predecessor M3 산출) — extend with manager-ddd test case mirroring manager-tdd case from PR #776. Plus manual end-to-end verification on a local moai project after `moai update`.

---

## 2. AC ↔ REQ Coverage Matrix

| AC | Mapped REQs | Coverage Type |
|---|---|---|
| AC-RD-01 | REQ-RD-001, REQ-RD-002, REQ-RD-006 | Positive (golden path) |
| AC-RD-02 | REQ-RD-003, REQ-RD-008, REQ-RD-009 | Edge (annotation + body format) |
| AC-RD-03 | REQ-RD-004, REQ-RD-010, REQ-RD-012 | Boundary (audit fail-fast) |
| AC-RD-04 | REQ-RD-005, REQ-RD-007, REQ-RD-012 | Negative (runtime block) |

**Total**: 4 ACs × 3 average REQs = 12 REQ slots assigned. 11/12 unique REQs covered (REQ-RD-011 deferred per predecessor REQ-RA-014). REQ-RD-012 is composite and appears in both AC-RD-03 (audit aspect) and AC-RD-04 (runtime aspect).

## 3. Out-of-AC Scenarios (informational)

The following scenarios are NOT acceptance criteria but should be considered during plan execution:
- mo.ai.kr 사이드 프로젝트 동작 검증 — out of scope (`moai update` 자동 sync 정책).
- `moai agents list --retired` CLI 출력 — REQ-RD-011 deferred.
- predecessor RETIRED-AGENT의 retired stub (manager-tdd.md) 회귀 검증 — predecessor PR #776 CI에서 이미 검증됨; 본 SPEC은 추가 회귀 테스트하지 않음.
- text/template path interpolation — RETIRED-AGENT REQ-RA-006 영역.

---

End of acceptance.md
