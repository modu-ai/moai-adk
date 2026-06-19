---
id: SPEC-CC2178-TEAM-API-ALIGN-001
title: "Align MoAI doctrine + templates + docs to Claude Code 2.1.178 implicit-team API"
version: "0.1.0"
status: implemented
created: 2026-06-19
updated: 2026-06-20
author: manager-spec
priority: High
phase: "v3.0.0"
module: ".claude/rules, .claude/skills, CLAUDE.md, internal/template/templates, docs-site/content"
lifecycle: spec-anchored
tags: "claude-code, agent-teams, doctrine-alignment, template-mirror, docs-sync"
related_specs: [SPEC-CC2178-DOCS-ALIGN-001, SPEC-CC2178-MODEL-POLICY-REPAIR-001]
tier: M
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-19 | manager-spec | Initial draft. Umbrella docs/doctrine alignment SPEC tracking the Claude Code 2.1.178 Agent Teams API change (TeamCreate/TeamDelete removal → implicit-team model) plus adjacent 2.1.179/2.1.181/2.1.183 facts. Three axes: T1-1 (team-API doctrine alignment), T1-3 (`/config` command docs), D1 (`attribution.sessionUrl` settings key). |

---

## §A. Context and Motivation

### A.1 The upstream change

Claude Code v2.1.178 removed the `TeamCreate` and `TeamDelete` tools from the Agent Teams subsystem and replaced the explicit create/name/delete lifecycle with an **implicit team** model. MoAI doctrine, templates, and user-facing documentation currently instruct operators to use these now-removed tools, creating a documentation defect: following the doctrine verbatim would invoke tools that no longer exist.

This SPEC aligns every MoAI surface that names the removed tools, and the adjacent confirmed behavioral changes shipped in the same release band (2.1.178..2.1.183), with the current Claude Code behavior.

### A.2 Research basis (verbatim upstream evidence)

The load-bearing facts were verified against two official sources prior to this SPEC. They are cited here as the research basis; they are NOT re-fetched.

1. **Claude Code CHANGELOG v2.1.178 (verbatim)**: "Agent teams: removed the `TeamCreate` and `TeamDelete` tools. With `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` set, every session now has one implicit team — spawn teammates directly with the Agent tool's `name` parameter, no setup step needed. The `team_name` parameter on the Agent tool is still accepted but ignored."

2. **Official agent-teams doc (verbatim)**: "Before v2.1.178, you asked Claude to create and name a team first, and Claude used the `TeamCreate` and `TeamDelete` tools to set it up and remove it. Both tools no longer exist. The `team_name` input on the Agent tool is accepted but ignored, and the `team_name` field in `TaskCreated`, `TaskCompleted`, and `TeammateIdle` hook payloads carries the session-derived name and is deprecated."

3. **Adjacent confirmed facts (same doc)**: `teammateMode` default changed from `auto` to `in-process` as of v2.1.179 (split panes no longer auto-open); an idle teammate's agent-panel row hides after 30s and reappears on the next turn (v2.1.181); teams/tasks are stored under the session-derived name `session-<first8>`; one team per session; no nested teams.

Cross-reference: `.moai/research/cc-update-2.1.164-to-2.1.183.md`.

### A.3 Sibling SPEC lineage

The `CC2178` domain is established. Two siblings preceded this SPEC and are both `completed`:
- `SPEC-CC2178-DOCS-ALIGN-001` — earlier CC 2.1.169..178 documentation alignment.
- `SPEC-CC2178-MODEL-POLICY-REPAIR-001` — model×effort×cycle_type policy alignment.

This SPEC is the team-API axis of the same upstream-tracking program.

---

## §B. Scope — Three Axes

### B.1 Axis T1-1 — Team-API doctrine alignment (primary, High)

The implicit-team model MUST be reflected in MoAI doctrine that currently instructs use of the removed `TeamCreate`/`TeamDelete` tools. The corrected model:

- Teams form **implicitly on first teammate spawn** when `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is set — there is no setup step.
- Teammates are spawned via the Agent tool's `name` parameter (e.g., `Agent(subagent_type: ..., name: ...)`) — no `TeamCreate` precedes them.
- The `team_name` parameter on the Agent tool is **accepted but ignored**.
- The `team_name` field in the `TaskCreated` / `TaskCompleted` / `TeammateIdle` hook payloads carries the **session-derived name** and is **deprecated**.
- Team cleanup is **automatic on session exit** — instructions of the form "Call TeamDelete only after all teammates have shut down" MUST be removed or replaced with the automatic-cleanup statement.
- Teams/tasks are stored under the session-derived name `session-<first8>`; one team per session; no nested teams.

Adjacent facts to reflect **where the doctrine already describes `teammateMode`** (no scope expansion beyond existing mentions):
- `teammateMode` default `auto → in-process` (v2.1.179); split panes no longer auto-open.
- Idle teammate agent-panel row hides after 30s and reappears on next turn (v2.1.181).

### B.2 Axis T1-3 — `/config` command documentation (secondary, Med)

Update the genuine Claude Code `/config` command documentation to reflect the v2.1.178-band behavior:
- `/config key=value` direct-set form.
- `/config --help` lists shorthand keys.
- Toggle-key behavior change: Enter AND Space both change the selected setting; Esc now **saves-and-closes** instead of reverting.

This axis is gated on a PRECISE grep at authoring/run time: bare `/config` is extremely noisy (it matches `.moai/config/` paths). Only the genuine Claude Code `/config` command references are in scope.

### B.3 Axis D1 — `attribution.sessionUrl` settings key (small, Med)

Add the new `attribution.sessionUrl` settings key (CC 2.1.183: omits the claude.ai session link from commits/PRs in web and Remote Control sessions) to the template `settings.json.tmpl` `attribution` block, and add one doctrine line documenting it in `settings-management.md` (which currently has zero `attribution` coverage). The exact value type and sub-key schema MUST be confirmed against the official Claude Code settings reference before the edit.

---

## §C. Requirements (GEARS notation)

### C.1 Team-API doctrine alignment (Axis T1-1)

- **REQ-TAA-001** (Ubiquitous): The aligned doctrine corpus **shall** describe the implicit-team model — teams form on first teammate spawn, with no `TeamCreate`/`TeamDelete` setup or teardown step.

- **REQ-TAA-002** (Ubiquitous): The aligned doctrine corpus **shall not** instruct any actor to call the removed `TeamCreate` or `TeamDelete` tools as a live action.

- **REQ-TAA-003** (Event-driven): **When** the doctrine describes spawning a teammate, the doctrine **shall** present spawning via the Agent tool's `name` parameter as the canonical mechanism.

- **REQ-TAA-004** (Ubiquitous): The aligned doctrine corpus **shall** state that the `team_name` parameter on the Agent tool is accepted but ignored, and that the `team_name` field in the `TaskCreated` / `TaskCompleted` / `TeammateIdle` hook payloads is deprecated and carries the session-derived name.

- **REQ-TAA-005** (Ubiquitous): The aligned doctrine corpus **shall** state that team cleanup is automatic on session exit, and **shall not** retain the instruction "Call TeamDelete only after all teammates have shut down" (or equivalent manual-teardown instructions).

- **REQ-TAA-006** (Where capability gate): **Where** existing doctrine already describes `teammateMode`, the doctrine **shall** reflect the `auto → in-process` default change (v2.1.179) and the split-pane no-longer-auto-opens behavior, and **shall** note the v2.1.181 idle-row-hide behavior in the same locations — without introducing `teammateMode` discussion to files that do not already mention it.

- **REQ-TAA-007** (Ubiquitous): The CLAUDE.md §15 "Team APIs" enumeration (`TeamCreate, SendMessage, TaskCreate/Update/List/Get, TeamDelete`) and the "Call TeamDelete only after all teammates have shut down" line **shall** be rewritten to the implicit-team API surface (no `TeamCreate`, no `TeamDelete`), and its template mirror `internal/template/templates/CLAUDE.md` **shall** carry the identical rewrite.

- **REQ-TAA-008** (Ubiquitous): The CLAUDE.md §4 "Watch (Claude Code 2.1.172)" nested-subagent note **shall** remain factually consistent with the implicit-team alignment; this SPEC **shall not** expand scope beyond team-API facts when touching that note (consistency check only).

### C.2 `/config` command documentation (Axis T1-3)

- **REQ-CFG-001** (Event-driven): **When** doctrine documents the genuine Claude Code `/config` command, the doctrine **shall** present the `/config key=value` direct-set form and the `/config --help` shorthand-key listing.

- **REQ-CFG-002** (State-driven): **While** the `/config` settings selector is focused, the doctrine **shall** describe that Enter AND Space both change the selected setting and that Esc saves-and-closes (no longer reverts).

- **REQ-CFG-003** (Ubiquitous): The `/config` edits **shall** be applied ONLY to genuine Claude Code `/config` command references, identified by a precise backtick-quoted (`` `/config` ``) or command-form grep, and **shall not** touch `.moai/config/` filesystem-path references.

### C.3 `attribution.sessionUrl` settings key (Axis D1)

- **REQ-ATT-001** (Ubiquitous): The template `settings.json.tmpl` `attribution` block **shall** carry a `sessionUrl` sub-key alongside the existing `commit` and `pr` keys, with the value type confirmed against the official Claude Code settings reference.

- **REQ-ATT-002** (Ubiquitous): `settings-management.md` **shall** document the `attribution` block including the new `sessionUrl` key in at least one doctrine line.

### C.4 Cross-cutting constraints

- **REQ-MIR-001** (Ubiquitous): Every edit to a local `.claude/**` file or `CLAUDE.md` **shall** have a byte-identical corresponding edit in its `internal/template/templates/**` mirror (template-mirror parity per CLAUDE.local.md §2/§24).

- **REQ-MIR-002** (Event-driven): **When** any template-managed source file under `internal/template/templates/` is edited, the build **shall** regenerate `internal/template/embedded.go` via `make build`.

- **REQ-NEU-001** (Ubiquitous): Template-side edits **shall** remain language-neutral and free of internal-content leak (no SPEC IDs, REQ tokens, internal dates, commit SHAs in template files) per CLAUDE.local.md §15/§25; the public-CC-mechanism text describing the implicit-team model is acceptable content.

- **REQ-LOC-001** (Ubiquitous): docs-site edits **shall** be applied across the `en` / `ko` / `ja` / `zh` locales equally (4-locale sync per CLAUDE.local.md §17), with per-locale verification.

- **REQ-LOC-002** (Event-driven): **When** a team-API reference exists in one locale but is absent in the sibling locales (locale divergence), the alignment **shall** reconcile the divergence — either by removing the obsolete reference from the divergent locale or by confirming the sibling locales legitimately never carried it.

- **REQ-LOC-003** (Ubiquitous): The 4 repo-root README locale files (`README.md` / `README.ko.md` / `README.ja.md` / `README.zh.md`) **shall** carry zero stale `TeamCreate` / `TeamDelete` live-API references after alignment. README is repo-root project-owned — NOT template-mirrored (outside REQ-MIR-001) and NOT docs-site (outside REQ-LOC-001/002) — so this orphan edit class **shall** have dedicated acceptance coverage (AC-LOC-003).

- **REQ-QG-001** (Event-driven): **When** the run-phase edits complete, the quality gate (`go test ./...`, `golangci-lint`, and the template-neutrality test `TestTemplateNeutralityAudit`) **shall** pass.

- **REQ-GO-001** (Ubiquitous): The Go test/source references to the removed-tool names (`internal/cli/team_spawn_test.go`, `internal/cli/team_run_audit_gate_test.go`, `internal/runtime/audit_gate.go`, `internal/template/skills_audit_test.go`) **shall** be assessed for whether each is a live-behavior reference (requiring a code change) or a doctrine-text/comment reference (tracking the doctrine edits); the assessment finding **shall** be recorded and only genuine live-behavior references **shall** be changed.

---

## §D. Acceptance Criteria Summary

Full Given-When-Then enumeration and grep/command verification live in `acceptance.md` (18 ACs). The acceptance matrix covers: zero residual `TeamCreate`/`TeamDelete` live-action instructions in the 16-file doctrine corpus + 4-locale docs-site + 4-locale README (AC-LOC-003, dedicated orphan-class coverage); implicit-team model present; template-mirror parity (16 == 16); 4-locale parity; `attribution.sessionUrl` added conditional on in-instance type confirmation; `/config` precise-scope-only edits (baseline-count invariant); quality-gate green; Go-reference assessment recorded.

---

## §E. Self-Verification Signal

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored: spec.md + plan.md + acceptance.md + design.md + progress.md.
- SPEC ID self-check: `decomposition: SPEC ✓ | CC2178 ✓ | TEAM ✓ | API ✓ | ALIGN ✓ | 001 ✓ → PASS`.
- Frontmatter 12-canonical-field schema validated.
- Blast radius measured live (not assumed): 16 doctrine files (15 `.claude/**` + CLAUDE.md) == 16 template mirrors; 9 docs-site files (4×2 + ko/advanced/hooks-guide); 4 README locales; 4 Go references assessed (all comment/text, none live-behavior).
- Out of Scope section present (§J).
- plan-auditor verdict: PASS 0.89 (1 major + 3 minor defects). Pre-run-phase fixes applied (status remains draft): D1 → AC-LOC-003 (README orphan-class, MUST-FIX) + REQ-LOC-003; D2 → `.moai/config/`-path baseline capture added to plan.md §C pre-flight; D3 → AC-CFG-001 tightened to 3 non-vacuous anchored checks; D4 → OQ-1 resolution status recorded (schemastore stale, type unconfirmed) + M5 conditional on in-instance confirmation; D5 → OQ-2/OQ-3 recommended defaults encoded as Implementation-Kickoff-Approval-overridable. Counts after fixes: 21 REQs, 18 ACs (added REQ-LOC-003 + AC-LOC-003).
- _Audit-ready: pending Implementation Kickoff Approval (plan→run human gate)._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §J. Exclusions (Out of Scope)

The following are explicitly out of scope. This SPEC focuses on WHAT and WHY, not HOW; implementation details (exact wording, function names) are deferred to run-phase.

### Out of Scope — Historical and immutable artifacts

- `.moai/design/v3-legacy/**`, `.moai/design/v3-redesign/**`, `.moai/design/v3-research/**` — frozen design history.
- `docs/design/major-v3-master.md` — historical design master.
- root `CHANGELOG.md` historical entries — immutable provenance.
- `.moai/research/**` — research reports are read-only inputs, not edit targets.
- `.moai/state/last-cc-version.json` — runtime/dev-only state.

### Out of Scope — Transient agent worktrees

- `.claude/worktrees/agent-*/**` — transient agent worktree copies (e.g., `agent-a04f3ccc3d653e2bd`); these are ephemeral and MUST NOT be edited. The live `.claude/**` tree and its template mirror are the only doctrine edit targets.
- `.moai/backups/**` — archived agent/template backups.

### Out of Scope — Behavioral / code-logic changes to the team subsystem

- Changing the actual MoAI Go team-spawn implementation logic. This SPEC is doctrine/template/docs alignment; it does NOT alter `internal/cli` team-spawn behavior, `internal/runtime/audit_gate.go` gate logic, or the swarm-registry contract. Go references are assessed (REQ-GO-001), and only references that are genuine live invocations of a removed Claude Code tool (none found at authoring time) would be changed.

### Out of Scope — Nested-subagent adoption

- Adopting or revising the nested-delegation capability described in the CLAUDE.md §4 "Watch (Claude Code 2.1.172)" note. That note is touched only for factual consistency (REQ-TAA-008); nested-delegation adoption remains deferred per the note's own guidance.

### Out of Scope — New team-API features

- Documenting team APIs beyond the confirmed implicit-team facts (e.g., speculative future team commands). Only the verbatim-confirmed 2.1.178..183 facts (§A.2) are in scope.

### Out of Scope — `/config` filesystem-path references

- `.moai/config/` and any non-command `/config` substring matches. Only the genuine Claude Code `/config` command references (REQ-CFG-003) are edited.
