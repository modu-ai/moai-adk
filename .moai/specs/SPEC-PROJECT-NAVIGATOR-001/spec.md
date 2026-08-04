---
id: SPEC-PROJECT-NAVIGATOR-001
title: "Project Navigator — living project-scoped navigation layer (capability-map + progress-map + entry point + resume reorientation)"
version: "0.2.0"
status: draft
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P1
phase: "v3.2 target"
module: project-navigator
lifecycle: spec-anchored
tier: M
tags: "navigator, project-context, session-continuity, reorientation, provenance, sync-phase"
related_specs: [SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003, SPEC-LSEL-LOCAL-EVOLUTION-001, SPEC-WORKFLOW-LIFECYCLE-001]
---

# SPEC-PROJECT-NAVIGATOR-001 — Project Navigator (living docs + resume reorientation)

## HISTORY

- 2026-08-05 (iter-1) — Initial draft. Decomposes the user-approved "Project Navigator" epic (P0–P3) into an Epic of three SPECs; this entry SPEC owns P0 (living Navigator documents) + P1 (resume-to-frontier reorientation). The sibling P2 (`--audit` drift check) is folded into SPEC-PROJECT-NAVIGATOR-002 and P3 (tree-sitter auto-derivation, 16-language) into SPEC-PROJECT-NAVIGATOR-003 — both stubbed at plan-phase, neither authored here. Research grounding (Aider repo map, Codebase-Memory, AGENTS.md, MemGPT/Mem0, Claude compaction) is recorded in `research.md`; prior-work boundary vs `SPEC-LSEL-LOCAL-EVOLUTION-001` is in `plan.md` § Boundary vs LSEL.
- 2026-08-05 (iter-2 fold) — plan-auditor iter-1 returned FAIL on MP-7 (3 surviving open-question markers; aggregate 0.86 ≥ Tier M 0.80 — score fine, FAIL purely the clarification firewall). User endorsed the SPEC's own proposed defaults and added three directives, applied here: (A) the three open decisions are RESOLVED at normative defaults — staleness `N=3` sync cycles (overridable via `navigator.staleness_cycles`), SessionStart hint ≤500 tokens, keep BOTH the ambient SessionStart hook and the on-demand skill mode; (B) `--resume` renamed to `--brief` everywhere — the output is a status BRIEF not a resume action, and the rename avoids collision with native Claude Code `/resume` (session resume) per the native-invocation anti-collision principle (cross-ref `.claude/rules/moai/workflow/goal-directive.md` § "Native /goal Prohibition" — the analogous case); (C) Navigator reframed as a PROJECT-LEVEL SHARED READ PRIMITIVE — `/moai project` owns maintenance (write/regenerate/audit), other subcommands consume (read) at defined orientation phases — captured in new REQ-PN-017 and REQ-PN-018 plus the plan.md §C Integration map; (D) auditor defects D2/D3/D4 applied — GEARS labels canonicalized to `Event-driven`, `tier: M` added to frontmatter, AC-PN-008 rewritten to a deterministic concurrency fixture. All live open-question markers removed (MP-7 re-audit ready).

## §A. User Story

**As a** MoAI-ADK user driving a large or long-running project whose AI coding sessions are frequently interrupted by other branches, urgent tasks, and session changes (`/clear`, compaction, multi-day gaps),
**I want** a project-scoped, living navigation layer that (1) tracks the whole architecture's feature/status inventory in one place, (2) auto-stays-current on every sync-phase, (3) reorients any returning session to the current frontier in a single read, and (4) carries provenance (commit-sha + timestamp) on every row so no claim is unobserved,
**so that** a returning agent reads ONE entry document and knows what is done, what is in progress, and what is next — instead of re-deriving the project state from many scattered files (SPEC registry, @MX tags, progress.md files, codemaps, product/structure/tech.md).

**Outcome hypotheses:**
- A session resuming after a `/clear` or a multi-day gap loads `navigator.md` as SessionStart additionalContext and reaches frontier-awareness in one pass (measured: zero re-derivation tool calls before the first work-action).
- Every row in `capability-map.md` and `progress-map.md` carries `commit-sha` + `captured-at` so a reader can pin any claim to a baseline (aligns with `.claude/rules/moai/core/verification-claim-integrity.md` — no unobserved claims).
- The Navigator regenerates on every `/moai sync`, so its staleness window equals the project's release cadence (never a separate manual refresh step).

## §B. Context and Background

### §B.1 The gap (research-backed — see research.md)

MoAI's existing continuity primitives are all **session-scoped** or **feature-scoped**:

| Primitive | Scope | Why it does not close the gap |
|-----------|-------|-------------------------------|
| `session-handoff` (`session-handoff.md`) | ONE SPEC across `/clear` | Resumes one feature; does not map the whole project |
| auto-memory (`MEMORY.md` + `feedback_*.md`) | Past-tense lessons/incidents | Records what was learned, not what the project IS |
| `/moai goal` | One session, completion-condition | Within-session continuation, not cross-session reorientation |
| `/moai codemaps` | CODE structural map | Static snapshot, regenerated on demand; does not track feature/status |
| `@MX` tags | Symbol-level annotations | Micro-context, no project-level rollup |
| SPEC lifecycle (plan/run/sync) | Per-feature artifacts | Each SPEC is one feature; nothing aggregates them |
| `/moai project` (skill) | Generates product/structure/tech.md | Generated ONCE at init; not living |

The **GAP**: no PROJECT-SCOPED, LIVING navigation map that aggregates the SPEC registry into a single reorientable view. A returning agent must re-derive the current state by reading many files — the exact failure mode the user named ("the AI loses its way").

### §B.2 Navigator artifact set (P0)

Three files under `.moai/project/navigator/`:

| File | Content | Regenerated when |
|------|---------|------------------|
| `navigator.md` | Single entry point: "current frontier + next task + missing specs (link to 002) + recent activity" | every sync |
| `capability-map.md` | Feature inventory rows: `<capability> \| owning-spec \| status \| implementation-path \| commit-sha \| captured-at` | every sync |
| `progress-map.md` | Per-SPEC progress rollup: `<spec-id \| phase \| last-commit-sha \| last-commit-at \| frontier-milestone>` | every sync |

The directory is **generated**, not scaffolded — the template ships no fixed content (template-neutrality, constraint #1). The skill body + hook script are the only template-distributed surfaces.

### §B.3 Inputs (read-only)

The Navigator consumes; it does not own:

- `.moai/specs/SPEC-*/spec.md` (frontmatter `id`, `title`, `status`, `phase`, `module`, `tags`, `updated`) — the canonical SPEC registry
- `.moai/specs/SPEC-*/progress.md` (last-commit-sha, current milestone, blockers) — per-SPEC state
- `.moai/project/{product,structure,tech}.md` — design intent (the `--audit` mode of SPEC-002 diffs these against the capability-map)
- `@MX` tag corpus — symbol-level status (used by SPEC-003 for auto-derivation, NOT by 001)
- `git log` (commit-sha + timestamp for provenance) — the single source of truth for "when did this last change"

### §B.4 Mechanism (decided — see plan.md §C for full rationale)

- **Regeneration**: extended `moai-workflow-project` skill, invoked on `/moai project` AND chained into `/moai sync`. NOT a new Go CLI subcommand.
- **`--brief` reorientation (P1)**: (a) a SessionStart hook (`handle-session-start-navigator.sh`) reads `navigator.md` and emits a bounded additionalContext on every new session — the **ambient auto-brief** (fail-open, time-boxed per the Advisory-Check Discipline); AND (b) a `/moai project --brief` skill mode for on-demand reorientation that loads the full entry brief into context. Both surfaces are retained deliberately: the hook is ambient / automatic / light (frontier + next task + one link, ≤500 tokens); the skill mode is on-demand / explicit / full (the entire entry brief plus the current-frontier section of `progress-map.md`). They serve different use cases — a returning session gets the ambient brief automatically, while a mid-session re-orientation needs the explicit full brief.
- **`--audit` drift check (P2)**: owned by SPEC-002 (skill mode, not a hook — audit output can be large and is not latency-sensitive).

## §C. Requirements (GEARS)

### §C.1 Navigator artifact generation (P0)

#### REQ-PN-001 (Ubiquitous)
The Project Navigator SHALL produce exactly three living documents under `.moai/project/navigator/` — `navigator.md` (entry point), `capability-map.md` (feature inventory), and `progress-map.md` (per-SPEC progress rollup) — and no other top-level Navigator files.

#### REQ-PN-002 (Ubiquitous — provenance, aligns with verification-claim-integrity.md §2)
Every row in `capability-map.md` and `progress-map.md` SHALL carry a `commit-sha` plus an ISO-8601 `captured-at` timestamp drawn from `git log` for the owning file's last commit, such that any Navigator claim is attributable to a measured baseline.

#### REQ-PN-003 (Event-driven — sync-phase regeneration)
**When** `/moai sync` completes for any SPEC in the project, the Navigator documents SHALL be regenerated before the sync commit lands, so the Navigator's staleness window never exceeds one sync cycle.

#### REQ-PN-004 (Event-driven — `/moai project` regeneration)
**When** `/moai project` is invoked, the Navigator documents SHALL be regenerated alongside `product.md` / `structure.md` / `tech.md`, so a single `/moai project` invocation leaves the full project-context surface current.

#### REQ-PN-005 (Ubiquitous — idempotence)
The Navigator generator SHALL be idempotent: regenerating twice on the same commit SHA with no intervening SPEC change SHALL produce byte-identical output, so a no-op regeneration is a safe operation.

#### REQ-PN-006 (Capability-gate — empty-project resilience)
**Where** a project carries zero SPECs OR zero commits, the Navigator SHALL emit a minimal "no features tracked yet" entry document rather than failing, so initialization of the Navigator does not block first-time use.

#### REQ-PN-007 (Event-driven — malformed-frontmatter tolerance)
**When** the Navigator generator encounters a SPEC whose frontmatter is unparseable or missing required fields, the generator SHALL skip that SPEC row, emit a warning line into `.moai/logs/navigator-warnings.log`, and continue regenerating the remaining rows, rather than aborting the whole regeneration.

#### REQ-PN-008 (Event-driven — concurrent-regeneration safety)
**When** two sessions regenerate the Navigator concurrently, the write SHALL land via an atomic-rename strategy (write to `<file>.tmp` then `mv` into place), so the later write wins and no reader observes a partially-written file.

### §C.2 Resume reorientation (P1)

#### REQ-PN-009 (State-driven — SessionStart ambient auto-brief)
**While** `.moai/project/navigator/navigator.md` exists on disk at session start, the SessionStart hook (`handle-session-start-navigator.sh`, role: ambient auto-brief) SHALL emit a bounded additionalContext of at most 500 tokens that surfaces the Navigator entry brief — current frontier, next task, link to the full file — into the new session's context.

#### REQ-PN-010 (Ubiquitous — fail-open + time-boxed, per Advisory-Check Discipline)
The SessionStart Navigator hint SHALL be fail-open and time-boxed: on missing files, unreadable files, or a generation deadline exceed, the hook SHALL degrade to emitting no hint and exit 0, never blocking session start.

#### REQ-PN-011 (Capability-gate — on-demand full brief)
**Where** `/moai project --brief` is invoked, the `moai-workflow-project` skill SHALL load the full `navigator.md` entry brief plus the current-frontier section of `progress-map.md` into the active context as a structured reorientation brief, so a returning session reaches frontier-awareness in a single skill invocation.

#### REQ-PN-012 (Event-driven — staleness signal)
**When** the Navigator's most recent regeneration commit-sha is more than 3 sync cycles behind `HEAD` (default `N=3`, overridable via the `navigator.staleness_cycles` config key), the SessionStart hint SHALL include a staleness advisory naming the gap, so a returning session is warned that the Navigator may be missing recent work.

### §C.3 Inputs / non-duplication

#### REQ-PN-013 (Ubiquitous — non-duplication)
The Navigator SHALL consume the SPEC registry, `progress.md` files, `product/structure/tech.md`, and `git log` as read-only inputs and SHALL NOT duplicate their bodies into Navigator documents — Navigator rows are rollup references (spec-id, status, frontier, provenance), not copies of the underlying content.

### §C.4 Template / distribution

#### REQ-PN-014 (Ubiquitous — template-neutrality, constraint #1 + #4)
The template-distributed Navigator surfaces (the extended `moai-workflow-project` skill body and the `handle-session-start-navigator.sh` hook script under `internal/template/templates/.claude/hooks/moai/`) SHALL be template-neutral: they SHALL NOT contain internal SPEC IDs, REQ tokens, audit citations, internal dates, or commit SHAs (the C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1), and they SHALL carry no Go-bias in any per-language example.

#### REQ-PN-015 (Ubiquitous — 16-language neutrality)
The Navigator generator SHALL derive its inputs from project-agnostic sources (SPEC registry, git log, optional @MX corpus) and SHALL NOT assume a Go codebase, so the feature ships equally to all 16 supported languages.

### §C.5 Boundary vs LSEL (non-overlap)

#### REQ-PN-016 (Ubiquitous — non-overlap with SPEC-LSEL-LOCAL-EVOLUTION-001)
The Navigator SHALL operate strictly on project-scoped artifacts (`.moai/project/navigator/`, `.moai/specs/`, `git log`) and SHALL NOT read or write the LSEL surfaces (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` skills), so the two systems do not duplicate responsibility — Navigator answers "what does the project look like now", LSEL answers "how should the harness itself evolve".

### §C.6 Cross-subcommand integration (Navigator as shared read primitive)

The Navigator is a PROJECT-LEVEL SHARED READ PRIMITIVE, not a siloed `/moai project` feature. `/moai project` OWNS maintenance (write / regenerate / audit); other subcommands CONSUME (read) at defined orientation phases. The consultation points below are opt-in per subcommand ("use where needed"), NOT forced on every call.

#### REQ-PN-017 (Event-driven — plan-phase consultation)
**When** `/moai plan` enters its Phase 1 context-load step, the skill SHALL consult `.moai/project/navigator/navigator.md` (if present) to scope the candidate SPEC against the current frontier, so that duplicate SPECs are prevented and the new SPEC's boundary is drawn against the real project state.

#### REQ-PN-018 (Event-driven — run-phase consultation)
**When** `/moai run` begins a run-phase, the implementing agent SHALL consult the Navigator brief (current frontier + owning-SPEC's progress-map row) before implementation, so the agent is oriented to what is already done and what is next without re-deriving project state from many files.

The integration is read-only on the consumer side: `/moai plan` and `/moai run` only READ the Navigator; only `/moai sync` and `/moai project` WRITE it. REQ-PN-016 (LSEL non-overlap) is preserved — orientation reads from `.moai/project/navigator/` do not touch any LSEL surface.

## §D. Acceptance Criteria Matrix (summary)

Full Given-When-Then scenarios live in `acceptance.md`. Summary:

| AC | REQ | Check |
|----|-----|-------|
| AC-PN-001 | REQ-PN-001 | Three files exist after regeneration, no extras |
| AC-PN-002 | REQ-PN-002 | Every row carries `commit-sha` + ISO-8601 `captured-at` |
| AC-PN-003 | REQ-PN-003 | `/moai sync` regenerates Navigator before sync-commit |
| AC-PN-004 | REQ-PN-004 | `/moai project` regenerates Navigator alongside product docs |
| AC-PN-005 | REQ-PN-005 | Two regenerations on same commit → byte-identical |
| AC-PN-006 | REQ-PN-006 | Zero-SPEC project → "no features tracked yet" entry, exit 0 |
| AC-PN-007 | REQ-PN-007 | Malformed SPEC → row skipped + warning logged, others regenerated |
| AC-PN-008 | REQ-PN-008 | Atomic-rename write strategy verified |
| AC-PN-009 | REQ-PN-009 | SessionStart emits additionalContext with frontier brief |
| AC-PN-010 | REQ-PN-010 | Hook fails open on missing file / timeout |
| AC-PN-011 | REQ-PN-011 | `/moai project --brief` loads full brief |
| AC-PN-012 | REQ-PN-012 | Staleness advisory fires when >3 sync cycles behind HEAD (default; overridable via `navigator.staleness_cycles`) |
| AC-PN-013 | REQ-PN-013 | Navigator rows are references, not content copies |
| AC-PN-014 | REQ-PN-014 | Template neutrality grep clean (C2/C3/C7 classes absent) |
| AC-PN-015 | REQ-PN-015 | Generator runs on a non-Go fixture project |
| AC-PN-016 | REQ-PN-016 | Navigator grep touches no LSEL surface |
| AC-PN-017 | REQ-PN-017 | `/moai plan` Phase 1 consults `navigator.md` to prevent duplicate SPECs |
| AC-PN-018 | REQ-PN-018 | `/moai run` consults the brief before implementation |

## §E. Out of Scope

### Out of Scope — drift / completeness audit (`--audit`)

- The `--audit` mode that diffs design intent (`product.md` / `structure.md` / `tech.md`) against the implemented capability-map and flags missing SPECs / unimplemented features (명세 누락 detection) is owned by **SPEC-PROJECT-NAVIGATOR-002** (stubbed, not authored here). This SPEC delivers the artifact set that 002 will audit; 002 owns the audit algorithm.

### Out of Scope — tree-sitter auto-derivation (P3)

- Auto-deriving the capability-map's file/symbol/status columns from tree-sitter AST analysis (16-language) and integrating that into `/moai codemaps` is owned by **SPEC-PROJECT-NAVIGATOR-003** (Tier L, stubbed). This SPEC's `capability-map.md` is hand-derivation from the SPEC registry + git log only; AST-based enrichment is deferred.

### Out of Scope — Navigator-as-search-index

- The Navigator is a reorientation rollup, NOT a semantic search index over the codebase. Semantic search / RAG over source files is a separate concern; Aider-style PageRank symbol ranking (research.md §Aider) is the P3 territory of SPEC-003, not this SPEC.

### Out of Scope — cross-project / multi-repo navigation

- The Navigator is project-scoped (one repository). Federated navigation across multiple repositories (monorepo-of-monorepos) is out of scope; a future SPEC may address it if a real need surfaces.

### Out of Scope — LSEL-owned harness self-evolution

- Closing the PROPOSE→APPLY seam for harness self-edit (lessons-inbox drain, `feedback_*.md` generation, `hns-*` edits) is owned by **SPEC-LSEL-LOCAL-EVOLUTION-001**. The Navigator consumes the SPEC registry; it does not consume the lessons-inbox. See plan.md § Boundary vs LSEL.

## §F. Constraints (Non-Functional)

- **Provenance**: every row carries `commit-sha` + ISO-8601 `captured-at` (constraint #3, aligns with verification-claim-integrity.md).
- **Fail-open + time-boxed SessionStart**: the hook's read+render is bounded by a deadline; on exceed it degrades to no hint (per coding-standards.md § Advisory-Check Discipline).
- **Idempotent regeneration**: same commit + same SPEC set → byte-identical output (REQ-PN-005).
- **Atomic writes**: `<file>.tmp` → `mv` (REQ-PN-008).
- **16-language neutrality**: no Go-bias (constraint #4).
- **Template-First**: all new files under `.claude/` or `.moai/` ship via `internal/template/templates/` + `make build` + §25 neutrality (constraint #1).
- **No new Go CLI subcommand**: `/moai project` stays a skill; `--brief` is a skill mode, SessionStart hint is a hook (constraint #5, decided in plan.md §C).
- **PR-mandatory**: enforce_admins:true; plan-audit gate + Implementation Kickoff Approval human gate at plan→run (constraint #6).

## §G. Dependencies / Related SPECs

- **SPEC-PROJECT-NAVIGATOR-002** (stub) — consumes this SPEC's artifact set to run `--audit`.
- **SPEC-PROJECT-NAVIGATOR-003** (stub) — will feed enriched rows into `capability-map.md` via tree-sitter.
- **SPEC-LSEL-LOCAL-EVOLUTION-001** — non-overlap boundary (this SPEC's REQ-PN-016).
- **SPEC-WORKFLOW-LIFECYCLE-001** — sync-phase chain; Navigator regeneration hooks into the lifecycle this SPEC codifies.
- `moai-workflow-project` skill (`.claude/skills/moai-workflow-project/SKILL.md`) — extended, not replaced.
- `/moai codemaps` workflow (`.claude/skills/moai/workflows/codemaps.md`) — sibling; SPEC-003 integrates, this SPEC does not.

## §H. Resolved decisions (iter-2 fold — recorded for traceability)

The three open-question markers raised at iter-1 were resolved at the SPEC's own proposed defaults (user-endorsed "proceed", override window preserved via config):

1. **Staleness threshold** (REQ-PN-012): hard-coded default `N=3` sync cycles behind HEAD. Overridable via the `navigator.staleness_cycles` config key (NOT an in-SPEC variable). Configured at the framework layer, not the SPEC layer.
2. **SessionStart hint token budget** (REQ-PN-009): hard-coded ceiling of 500 tokens (frontier + next task + one link to the full file).
3. **Hook vs skill-mode overlap** (REQ-PN-009 vs REQ-PN-011): both surfaces KEPT. The SessionStart hook is the ambient / automatic / light path (auto-fires every session, ≤500 tokens); the `/moai project --brief` skill mode is the on-demand / explicit / full path (loads the entire entry brief + the current-frontier section of `progress-map.md`). Different use cases — automatic reorientation vs mid-session full re-orientation.

The `--resume` naming was renamed to `--brief` (iter-2 directive B): the output is a status BRIEF, not a resume action; and the rename avoids collision with native Claude Code `/resume` (session resume) per the native-invocation anti-collision principle — the same reasoning that blocks native `/goal` emission in `.claude/rules/moai/workflow/goal-directive.md` § "Native /goal Prohibition".
