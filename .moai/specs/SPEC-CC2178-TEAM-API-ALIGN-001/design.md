# Design — SPEC-CC2178-TEAM-API-ALIGN-001

This design document is included because the work is cross-cutting: a 16-file template-mirrored doctrine corpus, a 9-file 4-locale docs-site surface, and a conceptual API-model migration (explicit create/delete lifecycle → implicit team). An explicit edit-mapping and wording-transformation design reduces run-phase ambiguity and prevents the recurring 4-locale-parity drift failure mode.

## §A. The conceptual migration: explicit lifecycle → implicit team

| Concept | Pre-2.1.178 (current doctrine) | Post-2.1.178 (target doctrine) |
|---------|-------------------------------|--------------------------------|
| Team creation | `TeamCreate(name)` explicit setup step | Implicit — team forms on first teammate spawn when `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` |
| Teammate spawn | `Agent(..., team_name: X, name: Y)` after `TeamCreate` | `Agent(..., name: Y)` directly; `team_name` accepted but ignored |
| Team naming | operator-chosen name | session-derived `session-<first8>` |
| Team teardown | `TeamDelete` (manual, after shutdown) | automatic on session exit |
| Hook payload `team_name` | active field | deprecated; carries session-derived name |
| Team count / nesting | (unchanged) | one team per session; no nested teams |

The coordination *primitives that remain*: `SendMessage`, `TaskCreate/Update/List/Get`, the TeammateIdle/TaskCompleted hooks. Only the create/delete bookends are removed. Doctrine rewrites should preserve the coordination narrative and surgically replace the bookends.

## §B. Wording transformation patterns (apply consistently)

| Found pattern | Replacement pattern |
|---------------|---------------------|
| "Use `TeamCreate` for persistent team coordination" | "Spawn teammates directly with the Agent tool's `name` parameter; the team forms implicitly (one team per session)" |
| "`TeamCreate, SendMessage, TaskCreate/Update/List/Get, TeamDelete`" (API enumeration) | "`SendMessage`, `TaskCreate/Update/List/Get` (teams are implicit — spawn via `Agent(name=…)`, cleanup is automatic on session exit)" |
| "Call `TeamDelete` only after all teammates have shut down" | "Team cleanup is automatic on session exit; no explicit teardown call is needed" |
| "`TeamCreate(...)` + `Agent(..., team_name: ..., name: ...)` × N" (mode-orchestration) | "`Agent(subagent_type: …, name: …)` × N (implicit team; `team_name` accepted but ignored)" |
| Mermaid node `TeamCreate → SendMessage` | `Agent(name=…) → SendMessage` (recommended per OQ-2 option a) |
| Migration/historical note (keep) | "(`TeamCreate`/`TeamDelete` were removed in Claude Code v2.1.178)" — acceptable as a one-line migration note |

## §C. Edit-mapping table (file → axis → action)

### C.1 Doctrine corpus (M1) — 15 `.claude/**` + CLAUDE.md, each mirrored in M2

| File | Refs | Action |
|------|------|--------|
| `rules/moai/core/moai-constitution.md` | TeamCreate ×1 (L36) | Rewrite "Use TeamCreate for persistent team coordination" → implicit-spawn |
| `rules/moai/development/agent-authoring.md` | TeamCreate/Delete, team_name | Rewrite spawn guidance + team_name accepted-but-ignored |
| `rules/moai/development/orchestrator-templates.md` | TeamCreate/Delete | Rewrite template snippets to implicit spawn |
| `rules/moai/workflow/orchestration-mode-selection.md` | `TeamCreate(...)` + team_name (L28) | Rewrite mode-3 spawn expression |
| `rules/moai/workflow/spec-workflow.md` | TeamCreate/Delete | Rewrite team-phase narrative |
| `rules/moai/workflow/team-pattern-cookbook.md` | TeamCreate/Delete | Rewrite cookbook patterns |
| `rules/moai/workflow/team-protocol.md` | TeamDelete ×1 (L106 ledger archive) | Reframe: ledger archive happens on session-exit cleanup, not on a `TeamDelete` call |
| `skills/moai/references/reference.md` | TeamCreate/Delete, team_name | Rewrite API reference entries |
| `skills/moai/team/debug.md` | TeamCreate/Delete, team_name | Rewrite team-debug flow |
| `skills/moai/team/glm.md` | TeamCreate/Delete | Rewrite GLM team flow |
| `skills/moai/team/plan.md` | TeamCreate/Delete, team_name | Rewrite team-plan flow |
| `skills/moai/team/review.md` | TeamCreate/Delete, team_name | Rewrite team-review flow |
| `skills/moai/team/run.md` | TeamCreate/Delete, team_name | Rewrite team-run flow — **preserve the Phase 0.5 gate markers** (skills_audit_test requiredPatterns) |
| `skills/moai/workflows/plan.md` | TeamCreate/Delete | Rewrite |
| `skills/moai/workflows/run/mode-orchestration.md` | TeamCreate/Delete | Rewrite |
| `CLAUDE.md` §15 | enumeration + cleanup line (L488, L490) | Rewrite §15 Team APIs; §4 consistency-only |

### C.2 teammateMode reflection (M1, REQ-TAA-006) — only existing mentions

| File | Action |
|------|--------|
| `rules/moai/workflow/worktree-integration.md` | Add `auto→in-process` default note + idle-row-hide where teammateMode is discussed |
| `skills/moai-workflow-worktree/SKILL.md` | Same |

NOTE: `skills/moai-workflow-worktree/SKILL.md` and `worktree-integration.md` are also template-mirrored — include their mirrors in M2.

### C.3 docs-site (M3) — 9 files, 4-locale

| File group | Action |
|------------|--------|
| `{en,ko,ja,zh}/core-concepts/what-is-moai-adk.md` | Rewrite Mermaid `TeamCreate → SendMessage` node + the `TeamCreate, SendMessage, TaskList` bullet (×4 locales) |
| `{en,ko,ja,zh}/getting-started/introduction.md` | Rewrite Mermaid `TeamCreate → SendMessage` node (×4 locales) |
| `ko/advanced/hooks-guide.md` | Reconcile obsolete 3-step TeamDelete cleanup (L132-139) per OQ-3; verify en/ja/zh remain 0 |

### C.4 README (M3) — 4 locales

| File | Action |
|------|--------|
| `README.md` (L398, L421) | Rewrite Mermaid node (OQ-2 default `Agent(name=…) → SendMessage`) + `TeamCreate, SendMessage, TaskList` bullet |
| `README.ko.md` (L444, L467) | Same (ko) |
| `README.ja.md` (L447, L470) | Same (ja) |
| `README.zh.md` (L447, L470) | Same (zh) |

NOTE: README files are repo-root project-owned (NOT template-mirrored, NOT under template neutrality). Confirm during run-phase that root READMEs are project-owned and not subject to the neutrality guard.

**Acceptance coverage (D1 fix)**: README is an **orphan edit class** under the cross-cutting ACs — it is NOT template-mirrored (so AC-MIR-001 does not cover it) and NOT under docs-site (so AC-LOC-001/002 do not cover it). Its acceptance is owned by the dedicated **AC-LOC-003** (MUST-FIX): a 4-locale grep asserting zero stale `TeamCreate`/`TeamDelete` live-API references post-edit, plus a conditional Part B asserting the implicit-team form when OQ-2 default (a) is applied. Without AC-LOC-003 the README edits in this row would ship unverified.

### C.5 `/config` (M4) and `attribution` (M5)

| File | Axis | Action |
|------|------|--------|
| `rules/moai/workflow/dynamic-workflows.md` (+ any precise-grep hits) | T1-3 | Add `/config key=value`, `/config --help`, Enter/Space/Esc toggle docs (mirror) |
| `internal/template/templates/.claude/settings.json.tmpl` (attribution block) | D1 | Add `sessionUrl` sub-key — **conditional on OQ-1 in-instance type confirmation**; do NOT pin a literal until the type is confirmed in the maintainer's own CC instance (boolean most plausible per release-note, unverified) |
| `rules/moai/core/settings-management.md` (+ mirror) | D1 | Add attribution doctrine line incl. `sessionUrl` per release-note omit-session-link semantics, WITH a "verify type before pinning a default" caveat |

## §D. Risk and mitigation

| Risk | Mitigation |
|------|------------|
| 4-locale parity drift (canonical recurring failure) | Per-locale grep loop in AC-LOC-001/002, never glob-total. Edit all 4 locales in the same milestone (M3). |
| README orphan edit class (D1) | README is neither template-mirrored nor docs-site — dedicated AC-LOC-003 (MUST-FIX) provides 4-locale grep coverage; without it the README edits ship unverified. |
| Template-mirror divergence | AC-MIR-001 diff loop over every edited file; `make build` mandatory. |
| Breaking `skills_audit_test.go` | Preserve Phase 0.5 gate markers in `team/run.md`; re-run the test in M2 and M6. |
| `sessionUrl` schema unconfirmed (D4) | OQ-1 finding recorded: schemastore stale (predates 2.1.183), `/en/settings` lacks sub-key detail; type NOT confirmable from an authoritative machine-readable source. M5 conditional — confirm type in the maintainer's own CC instance before pinning a literal; otherwise M5 returns a blocker (AP-3). Doctrine line MAY document the key with a "verify type" caveat. |
| `/config` blast from bare grep | Precise command-form grep only; AC-CFG-002 baseline-count invariant (the `.moai/config/`-path count baseline is captured at run-phase entry per plan.md §C step 3). |
| Editing transient worktree copies | All greps exclude `.claude/worktrees/` and `.moai/backups/`. |
| Parallel-session race (RC2-README in flight) | Pre-spawn `git fetch` + divergence check at run-phase entry. |

## §E. Why Tier M (not S, not L)

- Not S: spans >2 files (16 mirrored doctrine + 9 docs-site + 4 README + 2 D1) and requires 4-locale + mirror-parity discipline — beyond a single-file minimal change.
- Not L: no code-logic change, no new subsystem, no architectural decision — mechanically bounded text alignment against confirmed upstream facts. The conceptual migration is documentation-only.
