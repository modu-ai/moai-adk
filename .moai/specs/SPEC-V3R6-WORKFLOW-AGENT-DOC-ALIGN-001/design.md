# Design — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

## §A — Design Goal

Specify HOW the ~125 archived-PROPER references (139 raw − 14 `researcher` role_profile KEEPs), 4 invalid team-agent files, 1 broken cross-reference, and 3 stale ground-truth values are remediated consistently across the live `.claude/` tree and the `internal/template/templates/.claude/` SSOT, and HOW `run.md` is trimmed below the 200-LOC ceiling without losing routing logic — while preserving every legitimate documentation pointer AND every live role_profile reference. (Directory truth: all agent bodies live under `.claude/agents/moai/`; there is NO `.claude/agents/builder/`.)

## §B — Replacement Mapping (per migration table §C)

The single authority for replacements is `.claude/rules/moai/workflow/archived-agent-rejection.md §C`. The design does NOT invent replacements; it applies the table:

The purge set is the **11 actually-archived agents** (NOT 12 — `researcher` is excluded; it is a live role_profile, see §B.1 class 4):

| Archived agent | Replacement (from §C) | Intent class to detect |
|----------------|------------------------|------------------------|
| manager-strategy | `manager-spec` (planning IS strategy) | spawn-target |
| manager-quality | `sync-auditor` (review) OR orchestrator verification batch (lint+test+coverage) | spawn-target |
| manager-brain | `Explore` → `manager-spec` sequential chain | spawn-target |
| manager-project | `manager-docs` (project docs ARE docs) | spawn-target |
| claude-code-guide (MoAI file) | **0 live occurrences — no action (D6).** Forward-guard only: DISAMBIGUATE from the valid Anthropic built-in of the same name per §C.3 | n/a (0 occurrences) |
| expert-backend/frontend/security/devops/performance/refactoring | per-spawn `Agent(general-purpose, model: opus, tools: <whitelist>, prompt: <domain instructions>)` | spawn-target |
| (GitHub-issue creation, formerly manager-quality) | orchestrator-direct `gh issue create` | spawn-target → CLI |

**`researcher` is NOT in this table (D2)** — per NOTICE.md the archived `researcher.md` file never existed; `researcher` is a live workflow.yaml `role_profiles:` entry. It is class 4 (KEEP) below, never a spawn-target to replace.

### §B.1 — Intent classification (FOUR classes)

Each occurrence is one of:
1. **Spawn-target / delegation instruction** (one of the 11 archived-proper names) → REPLACE per table.
2. **Documentation pointer to `archived-agent-rejection.md`** → KEEP verbatim.
3. **Frontmatter `agents:` list entry** (archived-proper name) → REMOVE or replace with canonical retained name.
4. **Live role_profile reference (KEEP — D2)** → a reference to one of the 7 live workflow.yaml `role_profiles:` — `researcher`, `analyst`, `architect`, `implementer`, `tester`, `designer`, `reviewer`. These are NOT archived agents; they are dynamic-team spawn profiles. KEEP every such reference untouched. The 14 `researcher` matches in the baseline are ALL class 4.

Detection heuristic:
- a line matching `subagent_type:`, `Agent(`, `delegate to`, `Delegate to`, or a milestone owner column, naming one of the 11 archived-proper names = **class 1**;
- a line containing `archived-agent-rejection` = **class 2**;
- a line under a YAML `agents:`/`triggers.agents:` key naming an archived-proper name = **class 3**;
- a line naming `researcher` (or any of the 7 role_profiles) in a `role_profiles:` context, a Team-Mode table, or an `agents: [...]` list that enumerates role profiles = **class 4 (KEEP)**.

Disambiguation rule for `researcher`: because `researcher` is ONLY ever a role_profile (it was never an archived agent file — NOTICE.md "originally absent"), EVERY `researcher` occurrence is class 4. There is no class-1 `researcher` to find. The over-purge guard (AC-WADA-001a) asserts the baseline count of 14 is preserved.

## §C — Group 2 Design Decisions

### §C.1 — REQ-WADA-010 direction: ADD the Terminology Glossary (recommended)

Two cross-references (`spec-workflow.md:23`, `worktree-state-guard.md:19`) point at `worktree-integration.md § Terminology Glossary` for L1/L2/L3 layer definitions. That section is absent.

**Decision: ADD the section** (rather than repoint). Rationale:
- `worktree-integration.md` is the natural canonical home for worktree-layer definitions (it owns the worktree selection decision tree, already cross-referenced from CLAUDE.md §14).
- L1/L2/L3 are referenced from multiple sites; a single canonical glossary location is more maintainable than scattering definitions.
- The L1/L2/L3 definitions already exist in prose elsewhere (CLAUDE.md §14 worktree isolation rules; session-handoff.md Block 0); the glossary consolidates them.

The added section MUST define: L1 = `Agent(isolation: "worktree")` runtime-autonomous worktree; L2 = SPEC worktree at `~/.moai/worktrees/<project>/<spec>/`; L3 = `/moai plan --worktree` opt-in initialization. (Exact wording is run-phase; design fixes the L1/L2/L3 semantics from the existing prose.)

### §C.2 — REQ-WADA-009: general-purpose pattern

`team/run.md` already uses `subagent_type: "general-purpose"` with role-profile overrides (verified plan-phase, lines 31/42/173/182/206/227). `team/plan.md` / `team/review.md` / `team/debug.md` simply adopt the same pattern, replacing `team-reader` (×3 in plan.md, ×3 in debug.md) and `team-validator` (×4 in review.md). `team/glm.md` table rows for `team-reader`/`team-validator` are either removed or relabeled to `general-purpose` role profiles.

### §C.3 — REQ-WADA-008: conditional

`release-update` dispatch not reproduced at plan-phase. Design: run-phase re-greps; if absent, no edit (vacuous pass). If present, repoint to `release.md` (confirmed to exist).

## §D — Group 4 Design: run.md LOC trim

### §D.1 — Constraint

`run.md` is 246 LOC; ceiling is 200 (`TestEntryRouterLOCCeiling`). Overage = 46 LOC. The entry-router argument-branching routing logic (added by SPEC-V3R6-ORCH-IGGDA-001) MUST be preserved.

### §D.2 — Strategy: compress-in-place first (OQ-3)

**Decision: compress prose in-place.** Rationale:
- M2's archived-agent purge already edits run.md; some archived-spawn prose may be deletable outright (e.g., a manager-quality delegation block that collapses to a one-line orchestrator-batch reference), naturally reducing LOC.
- The 46-LOC overage is prose-heavy (harness-level descriptions, anti-pattern lists) that can be tightened without removing routing branches.
- Extraction to a sub-file (`workflows/run/<x>.md`) is the fallback IF compression drops below clarity threshold — and it adds a template-mirror obligation for the new sub-file.

Sequencing: M2 does the archived purge edit; M6 does the final LOC trim as the LAST run.md edit so the `wc -l` check is stable. The two MUST NOT race (single-agent ownership of run.md across M2+M6).

## §E — Template-Mirror Design (changed-line, NOT whole-file byte parity)

**Correction (D4)**: NOT all in-scope files are byte-identical live↔template at baseline. Three agent bodies — `manager-spec.md`, `manager-docs.md`, `manager-git.md` — are ALREADY DIVERGED at baseline (verified `diff` → DIVERGED for all three) on the 3-phase-vs-legacy `§E.5`/Mx lifecycle content owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001 (OUT OF SCOPE here). A whole-file byte-parity gate would therefore be UNSATISFIABLE for these three regardless of this SPEC's work. The earlier "the mirror is byte-identical" claim was FALSE and is retracted.

The mirror requirement is therefore **changed-line parity**: each archived-purge edit this SPEC makes to a MIRRORED file must be present and identical in BOTH trees. Three file classes:

1. **Mirrored + baseline-identical** (most skill files): the changed lines land in both trees; whole-file diff happens to be clean too.
2. **Mirrored + baseline-diverged** (`manager-spec/docs/git.md`): apply the changed lines in both trees; whole-file diff REMAINS divergent (expected — the §E.5/Mx delta is not ours to reconcile). Verify only that the archived-purge edit is mirrored.
3. **Dev-only no-mirror** (`release.md`/`github.md`/`release-update.md` in `devOnlyLocalFiles`): NO template file exists; apply to the LIVE file only; EXCLUDED from `TestTemplateMirrorParity` (the test already excludes them) and from AC-WADA-016.

Because the archived-agent references are plain prose (not `{{.Var}}` templated — verified), the changed lines themselves are byte-identical across trees; the divergence is confined to OTHER (out-of-scope) sections of the 3 diverged files. After all edits, `make build` regenerates `internal/template/embedded.go`.

Verification: the bidirectional changed-line grep + `TestTemplateMirrorParity` set-comparison (AC-WADA-016) + `TestTemplateNeutralityAudit` (AC-WADA-017). The bidirectional iteration (loop BOTH trees) catches a template-only residual — the #1 regression risk a unidirectional `git diff`-driven check would miss. The replacement text MUST stay neutral — e.g., "per the archived-agent migration table" is neutral; "per <internal-SPEC-ID> <internal-REQ-token>" would leak internal tokens and FAIL §25.

## §F — DDD Characterization (PRESERVE)

Per `cycle_type=ddd`: the PRESERVE artifact is the routing-intent map (M1 inventory). For each file, the characterization is "this file routes <work> to <agent> for <purpose>". The IMPROVE step swaps `<agent>` (archived) for its canonical replacement while keeping `<work>` and `<purpose>` identical. No characterization test in Go is needed (these are markdown assets); the "behavior" is the grep-verifiable routing target, captured in the AC matrix.

## §G — Risks

| Risk | Mitigation |
|------|------------|
| sed/regex batch edit mis-replaces a documentation pointer as a spawn target | Per-occurrence Edit with Read-before-Edit; class-2 lines (pointers) explicitly excluded |
| **Over-purging `researcher` (D2)** — a blanket find-replace removes the live role_profile, breaking team-mode plan-phase spawn | Class-4 KEEP rule (§B.1); AC-WADA-001a over-purge guard asserts baseline count 14 preserved; constraint in plan.md §D |
| **Template-only residual escapes a unidirectional check (D4 #1 risk)** | AC-WADA-016 is BIDIRECTIONAL (loops both trees) + `TestTemplateMirrorParity` set-comparison; NOT a `git diff`-driven live-only iteration |
| **Whole-file diff false-FAIL on `manager-spec/docs/git.md` (D4)** | Changed-line mirror requirement (§E); baseline §E.5/Mx divergence is out of scope and NOT reconciled |
| run.md M2 + M6 edits race / double-count LOC | Single-agent ownership; M6 trim is the final run.md edit |
| Editing a nonexistent `.claude/agents/builder/` path (D1) | Directory truth note throughout; agent bodies are under `.claude/agents/moai/` |
| Mirroring a dev-only no-mirror file (`release.md`) creates a spurious template file | §E class 3 + plan.md M2 note: live-only edit; `devOnlyLocalFiles` exclusion |
| `claude-code-guide` built-in mistakenly purged | 0 live occurrences (D6); §C.3 disambiguation forward-guard only |
| Multi-session race resets working tree | Pre-flight re-grep (plan.md §C) + `feedback_shared_main_orphan_race` discipline |
| Replacement text leaks internal tokens into template | AC-WADA-017 neutrality test; design §E neutral-phrasing rule |

## §H — Cross-References

- `.claude/rules/moai/workflow/archived-agent-rejection.md §C` (replacement SSOT)
- `.claude/rules/moai/workflow/worktree-integration.md` (REQ-WADA-010 target)
- `CLAUDE.md §14` / `session-handoff.md` Block 0 (existing L1/L2/L3 prose to consolidate)
- `.moai/config/sections/workflow.yaml` `role_profiles:` (line ~68) — the 7 live role_profiles incl. `researcher` (D2 KEEP source)
- `internal/skills/workflow_split_test.go:154` (LOC ceiling test) + `:159/:184` (`TestTemplateMirrorParity` + `devOnlyLocalFiles` D4/D5 source)
- `.claude/rules/moai/NOTICE.md` (researcher "originally absent" — never an agent file, D2 evidence)
- `CLAUDE.local.md §25` (template neutrality)
