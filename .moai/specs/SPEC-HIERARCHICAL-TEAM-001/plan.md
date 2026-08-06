# plan.md — SPEC-HIERARCHICAL-TEAM-001

> Tier M implementation plan. 3-artifact set (spec + plan + acceptance). The `manager-lead` agent FILE is created at run-phase by `builder-harness` — this plan SPECIFIES the agent's role/tools/constraints/AC; it does NOT author the agent file at plan-phase. Ordered by decision-reversibility (highest-change-likelihood first) per manager-spec discipline.

## §A. Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team` (branch `feat/spec-hierarchical-team-001`, based on `origin/main`).
- **Branch HEAD at plan-phase start**: clean worktree, on `feat/spec-hierarchical-team-001`.
- **SPEC artifacts**: `.moai/specs/SPEC-HIERARCHICAL-TEAM-001/{spec.md, plan.md, acceptance.md, progress.md}`.
- **Sibling SPEC (completed, consumed)**: `SPEC-AUTONOMY-TIERS-001` — the `MOAI_AUTONOMY_TIER` token + 3-mode selection + tier→bundle renderer. This SPEC's `manager-lead` delegation predicate keys off Tier L scope.
- **plan-auditor verdict**: pending (plan-phase authoring now; audit immediately after).
- **Existing infrastructure to PRESERVE** (cross-cutting — every milestone respects these):
  - `internal/template/templates/**`, `CLAUDE.md`, distributed `.claude/rules/moai/**` — every distributed-surface edit mirrors to template source per §2 Template-First Rule (CLAUDE.local.md §2).
  - `internal/spec/era.go` `hasProgressMarker` / `hasAnyProgressMarker` / `extractProgressField` string-matchers — the fold row format (REQ-FOLD-001) MUST NOT collide with the `§E.2`/`§E.3`/`§E.4`/`§E.5` heading tokens or the `sync_commit_sha` / `mx_commit_sha` field names.
  - The existing `plan-research-fanout` skill's fixed-heading markdown schema (REQ-FANOUT-001 consumes it — does NOT redefine it).
  - The 11→12 retained-agent catalog integrity (REQ-LEAD-003) — adding `manager-lead` does NOT displace or archive any existing retained agent.
- **Existing infrastructure to EXTEND** (NOT replace):
  - CLAUDE.md §4 Selection Decision Tree (add 13th row), Retained Agents table (add 12th entry), Watch note (amend the flat-hierarchy claim).
  - `.claude/rules/moai/workflow/worktree-integration.md` § Worktree Selection Rules (re-key team-mode → parallel-write-workers-within-hierarchical-team).
  - `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution (same re-key, retain concurrency safeguard verbatim).
  - `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A Mode catalog — UNCHANGED (REQ-CLOSE-001); add a §G.2 note that `manager-lead` is a Mode-5-shaped delegation target, NOT a Mode 7.

## §B. Known Issues (manager-develop auto-injection, filtered to relevant categories)

- **B2 — Cross-SPEC Policy Conflict Pre-Scan**: this SPEC re-keys worktree gating previously coupled to the RETIRED Agent Teams static layer (`SPEC-V3R6-AGENT-TEAM-REBUILD-001` retired Mode 3). Run `grep -rn "team mode\|agent-team\|RETIRED" .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/core/agent-common-protocol.md` pre-M1 — confirm the re-key does NOT leave residual team-mode language that would re-introduce Mode 3 confusion.
- **B4 — Frontmatter Canonical Schema**: use `created: 2026-08-07` / `updated: 2026-08-07` / `tags:` (NOT snake_case aliases); `tier: M` (NOT `tier: "M"` quoted — enum is bare). `phase: "v3.x target"` — do NOT write `phase: plan` or `phase: run` (lifecycle-stage tokens are prohibited per spec-frontmatter-schema.md § Prohibited phase values; this is an error-severity lint).
- **B6 — spec-lint Heading Convention**: `## Out of Scope` (h2) alone triggers `MissingExclusions`. Each exclusion block MUST be `### Out of Scope — <topic>` (h3) with at least one `-` bullet. spec.md §B already follows this convention.
- **B10 — Untouched Paths PRESERVE**: do NOT touch `.moai/specs/SPEC-AUTONOMY-TIERS-001/`, `.moai/specs/SPEC-GOAL-HTML-WIRING-001/`, `.moai/specs/SPEC-STOPCHAIN-TRIM-001/`, or any sibling SPEC. Do NOT touch `internal/spec/era.go` (the fold row format is designed to coexist with its matchers — not to require matcher changes).
- **B11 — AskUserQuestion Prohibited (Subagent Boundary)**: `manager-lead` (a subagent) returns structured blocker reports; it does NOT call `AskUserQuestion`. The orchestrator runs the AskUser round on REQ-PEER-002 FAIL/PARTIAL (spec.md §C.4). This is already encoded in REQ-PEER-002.
- **Template-First mirror**: CLAUDE.md, `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/workflow/orchestration-mode-selection.md`, and `.claude/agents/moai/manager-lead.md` are ALL distributed surfaces — every edit mirrors to `internal/template/templates/` and runs `make build`. Per §25 template-neutrality, no SPEC IDs / REQ tokens / internal dates leak into the template copies (the distributed `manager-lead.md` carries generic role prose; the SPEC-HIERARCHICAL-TEAM-001 ownership is recorded in this SPEC's `progress.md` §E only, NOT in the template agent body).

## §C. Pre-flight (read-only reconnaissance — before M1)

```bash
# 1. Confirm branch + baseline
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team branch --show-current
# expect: feat/spec-hierarchical-team-001
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team rev-parse HEAD

# 2. Confirm the worktree-integration.md team-mode gating sites that re-keying will touch
grep -n "team mode" /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/.claude/rules/moai/workflow/worktree-integration.md
# expect: ≥2 matches (decision-tree line ~173, HARD-rule line ~194)

# 3. Confirm the plan-research-fanout skill's fixed-heading schema (REQ-FANOUT-001 contract source)
ls /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/.claude/skills/plan-research-fanout/ 2>/dev/null && \
  grep -n "^##\|^# " /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/.claude/skills/plan-research-fanout/SKILL.md 2>/dev/null | head -20

# 4. Confirm era.go §E.2 matcher is line-start anchored (the fold row prefix M<n>: must NOT be mis-parsed)
grep -n "§E.2\|hasAnyProgressMarker\|extractProgressField" /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/internal/spec/era.go | head -10

# 5. Cross-SPEC conflict pre-scan (B2)
grep -rn "manager-lead" /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/.claude/ /Users/goos/MoAI/moai-adk-go/.claude/worktrees/hier-team/.moai/specs/ 2>/dev/null | head -5
# expect: 0 matches in distributed doctrine + 0 matches in sibling SPECs (manager-lead is NEW); the ONLY pre-existing references are in SPEC-AUTONOMY-TIERS-001/spec.md §B (the sibling-scope pointer, line 71) which stays as-is.

# 6. spec-lint baseline on this SPEC (before any run-phase commit)
moai spec lint .moai/specs/SPEC-HIERARCHICAL-TEAM-001/ 2>&1 | tail -10
# expect: 0 errors, 0 strict errors
```

## §D. Constraints (recap from spec.md §D — binding on the plan)

- The `manager-lead` agent FILE is created at run-phase by `builder-harness` (NOT by this plan-phase). This plan SPECIFIES it; the agent body is a deliverable described in acceptance.md AC-LEAD-001.
- Distributed-surface edits (CLAUDE.md, `.claude/rules/moai/**`, `.claude/agents/moai/manager-lead.md`) ALL mirror to `internal/template/templates/` + `make build`. Template-neutrality (CLAUDE.local.md §25) preserved: no SPEC IDs / REQ tokens / internal dates / commit SHAs in template copies.
- Subagent boundary preserved: `manager-lead` does NOT call `AskUserQuestion`; it returns blocker reports on REQ-PEER-002 FAIL/PARTIAL.
- Flat-hierarchy carve-out: `manager-lead` is the SOLE retained agent carrying `Agent` in `tools:`; all other retained agents' `tools:` lists continue to omit it. Leaf workers spawned by manager-lead ALSO omit `Agent` (depth-2 seal — REQ-LEAD-001).
- Phase 4 mode catalog (orchestration-mode-selection.md §A) UNCHANGED — `manager-lead` is a Mode-5-shaped delegation target, NOT a Mode 7; Mode 3 (agent-team) tombstone preserved (REQ-CLOSE-001).
- The `/compact` slash command availability inside a subagent context is changelog-sourced and run-phase-verified (assumption 3, OQ-1).

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

Each item follows the 5-section Evidence-Bearing Report format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) per verification-claim-integrity §3.

- **E1. AC Binary PASS/FAIL Matrix** — every AC in acceptance.md §D verified PASS via the peer cross-validation of REQ-PEER-001 (or marked GAP with a blocker report).
- **E2. Cross-Platform Build result** — `go build ./...` exits 0; `GOOS=windows GOARCH=amd64 go build ./...` exits 0. (This SPEC touches mostly markdown; Go changes are limited to any CI guard added per OQ-4. If no Go changes, E2 trivially PASSes.)
- **E3. Coverage measurement** — ≥85% threshold on any Go package touched (likely zero Go packages if OQ-4 is decided "no CI guard").
- **E4. Subagent Boundary Grep (C-HRA-008 family)** — `grep -rn 'AskUserQuestion' .claude/agents/moai/manager-lead.md` returns 0 matches; the agent returns blocker reports, NOT prompts.
- **E5. Lint Status** — `moai spec lint --strict .moai/specs/SPEC-HIERARCHICAL-TEAM-001/` exits 0; `golangci-lint run` on any Go package touched exits 0.
- **E6. Branch HEAD + Push state** — feature-branch commits land on `feat/spec-hierarchical-team-001`; push succeeds; CI green per `repo-local-pr-policy.md` (Route B for all tiers in this repo).
- **E7. Blocker Report (if any)** — any unresolved OQ (OQ-1/OQ-2/OQ-3/OQ-4) decided at Implementation Kickoff that surfaces a downstream issue returns as a structured blocker; never an in-agent AskUserQuestion.

## §F. Milestones

> Ordered by decision-reversibility (most-likely-to-change first). The Implementation Kickoff Approval gate settles OQ-1/OQ-2/OQ-3/OQ-4 BEFORE M1 starts; the OQ decisions are M0 (pre-flight) inputs.

### M0 — Pre-flight + OQ resolution (post-Kickoff, pre-M1)

**Owner**: orchestrator (OQ decisions surfaced via AskUserQuestion at Implementation Kickoff Approval) + manager-develop (pre-flight command batch).

- Run the §C Pre-flight command batch; capture outputs.
- Record OQ-1/OQ-2/OQ-3/OQ-4 decisions from Implementation Kickoff Approval in `progress.md` §E.1 plan-phase audit-ready signal.
- Confirm the `plan-research-fanout` skill's schema is stable (Pre-flight §3); if NOT, escalate as a blocker BEFORE M1.

**Priority**: High (blocks M1).

### M1 — `manager-lead` agent file + CLAUDE.md §4 catalog (Axis 1, the most-reversible design decision)

**Owner**: builder-harness (agent file) + manager-develop (CLAUDE.md distributed-surface edits + template mirror).

- Author `.claude/agents/moai/manager-lead.md` (builder-harness): coordination-only role, `tools:` includes `Agent`, leaf-worker spawn contract (depth-2 seal).
- Amend CLAUDE.md §4: add the 13th Selection Decision Tree row; add the 12th Retained Agents table entry; amend the Watch note's flat-hierarchy claim to name `manager-lead` as the sole Agent-carrier.
- Amend CLAUDE.md §4 Supersession note for `SPEC-SUBAGENT-NESTING-DOCTRINE-001` to reference this SPEC's depth-2 seal as the active flat-hierarchy guarantee.
- Mirror all three edits to `internal/template/templates/.claude/agents/moai/manager-lead.md` + `internal/template/templates/CLAUDE.md`; run `make build`.
- Add the OQ-4 CI guard (if OQ-4 = "add guard"): a new test in `internal/template/` mirroring `subagent_boundary_test.go` that greps `manager-lead`-spawned leaf-worker agent files for `Agent` in `tools:` and fails on match.

**Self-verify**: AC-LEAD-001, AC-LEAD-002, AC-LEAD-003; `moai spec lint --strict` on this SPEC's directory exits 0; `grep -rn 'AskUserQuestion' .claude/agents/moai/manager-lead.md` returns 0.

**Priority**: High (this is the load-bearing Axis 1; everything else composes on top of it).

### M2 — Worktree re-keying (Axis 2, second-most-reversible — pure predicate substitution)

**Owner**: manager-develop.

- Re-key `worktree-integration.md` § Worktree Selection Rules decision tree line ~173 + HARD rule line ~194 from "team mode" → "parallel write workers within a hierarchical team (e.g., manager-lead fan-out)".
- Re-key `agent-common-protocol.md` § Background Agent Execution's stale team-mode framing; RETAIN the concurrency safeguard ("MoAI does not run two write-capable agents concurrently") verbatim.
- Mirror to `internal/template/templates/`; run `make build`.

**Self-verify**: AC-WORKTREE-001, AC-WORKTREE-002; `grep -rn "team mode" .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/core/agent-common-protocol.md` returns 0 matches in the re-keyed sections (the RETIRED Mode 3 references in orchestration-mode-selection.md are NOT in scope — they are the Mode-3 tombstone, preserved verbatim per REQ-CLOSE-001).

**Priority**: High.

### M3 — Context-Folding procedure (Axis 3, third-most-reversible — depends on M1's manager-lead)

**Owner**: manager-develop (procedure prose) + builder-harness (manager-lead agent body's fold sequence).

- Encode the 3-step Context-Folding procedure into `manager-lead.md`'s body: at Mn completion → (a) persist evidence to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`, (b) append `progress.md` §E.2 fold row, (c) invoke `/compact` with retain-current-milestone + retain-fold-rows instructions.
- Verify the fold row prefix `M<n>:` does NOT collide with era.go's `§E.*` heading matchers (Pre-flight §4).
- Apply the OQ-1 decision: `auto` at Tier L / `opt-in` at Tier M / `off` config flag wired into `.moai/config/sections/workflow.yaml` (default per OQ-1).
- Mirror to `internal/template/templates/`; run `make build`.

**Self-verify**: AC-FOLD-001, AC-FOLD-002, AC-FOLD-003; evidence path resolves at audit time; `/compact` invocation verified available in a subagent context (assumption 3 — if NOT available, blocker report and re-plan).

**Priority**: Medium (depends on M1; high value for Tier L but lower than M1/M2 because Tier L runs can still fall back to `/clear` + handoff without it).

### M4 — Peer cross-validation orchestration (Axis 4, depends on M1 + M3)

**Owner**: manager-develop (procedure prose) + builder-harness (manager-lead agent body's peer-spawn step).

- Encode the peer cross-validation spawn into `manager-lead.md`'s body: on `manager-develop` AC PASS → spawn second `Agent(general-purpose)` read-only to re-run acceptance.md §D commands → PASS/PARTIAL/FAIL.
- Encode REQ-PEER-002 blocker behavior: FAIL/PARTIAL → return blocker report; orchestrator runs AskUserQuestion.
- Apply the OQ-2 decision: all ACs at Tier M/L; Tier S skips.
- Mirror to `internal/template/templates/`; run `make build`.

**Self-verify**: AC-PEER-001, AC-PEER-002; `grep -rn 'AskUserQuestion' .claude/agents/moai/manager-lead.md` returns 0 (the blocker report path is taken, NOT the AskUser path).

**Priority**: Medium (depends on M1 + M3; structural strength of the per-AC validation is the value).

### M5 — Schema-driven fan-out reduce contract (Axis 5, the least-reversible in design but the most encapsulated)

**Owner**: manager-develop (procedure prose) + builder-harness (manager-lead agent body's reduce step).

- Encode the reduce step into `manager-lead.md`'s body: explorer returns must use `plan-research-fanout`'s fixed-heading markdown schema; reduce is mechanical merge; contradictions annotated as a named section.
- Encode REQ-FANOUT-002 concurrency ceiling (≤ 5 concurrent leaf-worker spawns).
- Mirror to `internal/template/templates/`; run `make build`.

**Self-verify**: AC-FANOUT-001, AC-FANOUT-002; reduce step verified mechanical (no per-spawn re-derivation).

**Priority**: Low (encapsulated; composes last).

### M6 — Non-regression on Phase 4 mode taxonomy + close

**Owner**: manager-develop.

- Add a §G.2 note to `orchestration-mode-selection.md` documenting that `manager-lead` is a Mode-5-shaped delegation target, NOT a Mode 7; Mode 1-6 catalog unchanged.
- Run the §E Self-Verification matrix end-to-end.
- Update `progress.md` §E.1 + §E.2 with the milestone-completion fold rows per REQ-FOLD-001.
- Confirm `moai spec lint --strict` exits 0 on this SPEC.

**Self-verify**: AC-CLOSE-001, AC-REGRESS-001; full §E matrix green.

**Priority**: Low (close-out).

## §G. Anti-Patterns (plan-specific, additive to spec.md §B exclusions)

- **AP-1 — Authoring `manager-lead.md` at plan-phase**: the agent file is a run-phase deliverable by builder-harness (CLAUDE.local.md §16 — `3+ agent/skill creation → builder-harness 强制`). Plan-phase SPECIFIES the agent; run-phase authors it.
- **AP-2 — Resurrecting Mode 3 (Agent Teams)**: the worktree re-keying (M2) decouples worktree use from team-mode WITHOUT resurrecting the static team-orchestration layer. Mode 3 stays RETIRED; `MODE_TEAM_UNAVAILABLE` stays; `--team` dispatch-axis value stays rejected. Any M2 edit that re-introduces team-mode orchestration machinery is scope-creep — re-route to a separate SPEC.
- **AP-3 — Compacting away the goal condition**: REQ-FOLD-003's retain-current-milestone + retain-fold-rows instruction to `/compact` MUST also retain any armed goal condition (context-window-management.md § Compaction Preservation). A fold that drops the goal condition violates the Compaction Preservation rule.
- **AP-4 — Peer worker = sync-auditor conflation**: peer cross-validation (run-phase, per-AC, binary PASS/PARTIAL/FAIL) is DISTINCT from sync-auditor (sync-phase, 4-dimension harmonic-mean score). Do NOT fold the two into a single "validator agent" — they operate at different phases with different scopes.
- **AP-5 — Reinventing the plan-research-fanout schema**: REQ-FANOUT-001 CONSUMES the existing skill's schema. Do NOT author a parallel `hierarchical-team-fanout-schema` skill — that is duplicate-prevention violation (coding-standards.md § Duplicate Prevention).
- **AP-6 — Treating OQ-1/OQ-2/OQ-3/OQ-4 as resolved at plan-phase**: the OQs are DEFERRED to Implementation Kickoff per the AUTONOMY-TIERS precedent. Plan-phase does NOT pre-decide them; the Implementation Kickoff Approval gate surfaces them.

## §H. Cross-References

- `.moai/specs/SPEC-AUTONOMY-TIERS-001/spec.md` — sibling completed SPEC whose OQ-1..OQ-4 pattern this plan's OQ-1..OQ-4 mirror.
- `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.3 — the design authority (cited via AUTONOMY-TIERS §G; the report is gitignored from the worktree but lives in the main checkout).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — the `(none) → draft` transition this plan-phase artifacts emission triggers (owner: manager-spec; commit subject `feat(SPEC-HIERARCHICAL-TEAM-001): plan-phase artifacts (Tier M, 3 artifacts)` per the canonical pattern).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M = 3 artifacts (spec + plan + acceptance), ceiling 16 REQs / 16 ACs. This SPEC carries 13 REQs + 14 ACs (within budget).
- `.claude/agents/moai/builder-harness.md` — the agent that authors `manager-lead.md` at run-phase M1 (CLAUDE.local.md §16 — 3+ agent creation mandates builder-harness).
- `.claude/skills/plan-research-fanout/` — the skill REQ-FANOUT-001 consumes (fixed-heading markdown schema for explorer returns).
- `internal/spec/era.go` — the §E.2 / §E.3 / §E.4 / §E.5 heading matchers the fold row format MUST NOT collide with (Pre-flight §4 verifies).
