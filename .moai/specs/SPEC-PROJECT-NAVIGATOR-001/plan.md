# Plan — SPEC-PROJECT-NAVIGATOR-001

> Plan-phase artifact. Order: highest-change-likelihood decisions first (mechanism, inputs, boundary) → mechanical/refactor steps last (template mirror, CI guards). The user will approve structure + mechanism at the Implementation Kickoff Approval gate.

## §A. Context

### §A.1 Branch + state

- Worktree: `.claude/worktrees/project-navigator-plan/` (L1 Claude-native, branch `worktree-project-navigator-plan`, base = `origin/main`).
- SPEC dir: `.moai/specs/SPEC-PROJECT-NAVIGATOR-001/` (newly created in this plan phase).
- Sibling stubs: `.moai/specs/SPEC-PROJECT-NAVIGATOR-002/`, `.moai/specs/SPEC-PROJECT-NAVIGATOR-003/` (spec.md stubs only, this same turn).

### §A.2 PRESERVE list (this SPEC writes NOTHING here)

- `.claude/rules/moai/**`, `CLAUDE.md`, `CLAUDE.local.md` — FROZEN doctrine.
- `.claude/agents/moai/**` — FROZEN retained agents.
- `internal/harness/applier.go`, `internal/harness/curator_dispatch.go` — FROZEN Go applier (LSEL territory; not touched here).
- `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` — LSEL surfaces; REQ-PN-016 forbids touching them.
- `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` and all other SPEC directories.

### §A.3 EXTEND targets (the only surfaces this SPEC adds to)

User-project surfaces (generated, not scaffolded — these are the Navigator's *output*, created at runtime by the skill):

- NEW (runtime-generated): `.moai/project/navigator/{navigator.md, capability-map.md, progress-map.md}` in each user project.
- NEW (runtime state): `.moai/state/navigator/{last-regen-commit.txt, warnings-count.txt}` (cache + staleness signal).
- NEW (runtime log): `.moai/logs/navigator-warnings.log`.

Template-distributed surfaces (the *mechanism*, shipped to all users via `internal/template/templates/`):

- EDIT: `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` — extend with a Navigator regeneration phase + a `--brief` skill mode.
- NEW: `internal/template/templates/.claude/skills/moai-workflow-project/references/navigator.md` — the regeneration procedure (Level 3 reference).
- NEW: `internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh` — SessionStart hint hook (fail-open, time-boxed).
- EDIT: `internal/template/templates/.claude/settings.json.tmpl` — register the new SessionStart hook.
- EDIT (CI guard): `internal/template/internal_content_leak_test.go` — extend forbidden-token set with `navigator` SPEC-ID leak guards (project-local SPEC-ID strip already enforced; add `SPEC-PROJECT-NAVIGATOR-` sentinel).

### §A.4 Epic decomposition decision

**Decision: Epic of three SPECs, NOT one Tier-L SPEC.** Tradeoff stated:

| Option | Pro | Con |
|--------|-----|-----|
| (a) ONE Tier-L SPEC with M1..Mn | Single audit gate; one PR series | P3 (tree-sitter, 16-lang) has a fundamentally different cost surface and ships on a different cadence; bundling it with P0/P1/P2 forces one giant 5-artifact plan whose milestones don't share a concrete acceptance surface |
| (b) Epic of 3 SPECs (**CHOSEN**) | Each SPEC has a focused audit gate and PR series; P3 can ship later without blocking P0/P1/P2; per-SPEC tier classification is honest (001=M, 002=S/M, 003=L) | Three plan-phase audit gates instead of one; slightly more ceremony |

Choosing (b) because the three sub-scopes have **different changeability profiles**: P0+P1 (living docs + reorientation) are tightly coupled (the SessionStart hint reads what regeneration produces, so they ship together); P2 (audit) is exploratory and its algorithm depends on what 001 surfaces; P3 (tree-sitter) is large, orthogonal, and 16-language-gated. Folding them into one SPEC would create milestones that don't share an acceptance surface.

This entry SPEC (001) owns **P0 + P1 only** and is classified **Tier M** (spec.md + plan.md + acceptance.md + progress.md + research.md).

## §B. Known Issues + [DECISION] markers

### §B.1 [DECISION] Mechanism for `--brief` and `--audit` (constraint #5)

Constraint #5 requires picking the simplest mechanism that fits the existing architecture. Survey of options:

| Mechanism | Fit | Verdict |
|-----------|-----|---------|
| New Go CLI subcommand (`moai navigator ...`) | Couples a project-level feature to the Go binary; violates "/moai project is a skill, not a CLI" | REJECTED |
| Skill mode (`/moai project --brief`) | `/moai project` is already a skill; modes are native to slash commands | **CHOSEN for --brief (on-demand surface)** |
| SessionStart hook (additionalContext) | Existing pattern (`handle-session-start.sh` already emits `hookSpecificOutput.additionalContext`); ambient, auto-fire on every session | **CHOSEN for --brief (ambient auto-brief surface)** |
| Agent instruction (manager-docs) | Sync-phase already runs manager-docs; chaining regeneration there is natural | **CHOSEN for regeneration trigger** |
| Stop hook | Wrong lifecycle surface — reorientation is a session-START need, not session-END | REJECTED |
| PostToolUse hook | Too granular; Navigator regen is a sync-cycle concern, not a per-edit concern | REJECTED |

**Final mechanism decision**:
- **Regeneration**: extend `moai-workflow-project` skill body. Triggered by `/moai project` AND chained into `/moai sync` (via manager-docs sync-phase instructions). No new CLI subcommand.
- **`--brief` (P1) — TWO surfaces, deliberately retained**: (a) SessionStart hook `handle-session-start-navigator.sh` emits a bounded additionalContext (≤500 tokens) on every new session — the **ambient auto-brief** (ambient / automatic / light: frontier + next task + one link; fail-open, time-boxed per Advisory-Check Discipline); AND (b) `/moai project --brief` skill mode — the **on-demand full brief** (loads the entire `navigator.md` entry brief + the current-frontier section of `progress-map.md`). Both read the same `navigator.md`. They serve DIFFERENT use cases — a returning session gets the ambient brief automatically (orientation at session start), while a mid-session re-orientation needs the explicit full brief (deep re-orientation during active work). Dropping either leaves a real gap: hook-only loses the full brief; skill-mode-only loses the zero-touch auto-orientation.
- **`--audit` (P2)**: skill mode only, owned by SPEC-002. NOT a hook — audit output can be large and is not latency-sensitive; making it a hook would violate Advisory-Check Discipline.

**Naming rationale (`--brief`, not `--resume`)** [iter-2 directive B]: (1) the output is a status BRIEF, not a resume action — it orients the session to project state, it does not resume an interrupted SPEC (that is `/moai run` + session-handoff's job); (2) avoids collision with native Claude Code `/resume` (session resume) per the native-invocation anti-collision principle — the same reasoning that blocks native `/goal` emission in `.claude/rules/moai/workflow/goal-directive.md` § "Native /goal Prohibition". The hook script filename stays `handle-session-start-navigator.sh`; its role text says "ambient auto-brief".

**Rationale**: this composes with existing primitives instead of inventing new ones. The hook layer already exists and already emits additionalContext; the skill body already exists and already has modes; manager-docs already runs in sync-phase. Zero new infrastructure categories.

### §B.2 [DECISION] Atomic-rename write strategy (REQ-PN-008)

Concurrent regeneration is a real risk in multi-session projects (this repo dogfoods it). Write strategy: each Navigator file is written to `<file>.tmp` then `mv`'d into place atomically. Readers either see the old version or the new version, never a partial. `last-regen-commit.txt` is written LAST, after all three files have landed, so its commit-sha is a sentinel for "all three files reflect this commit". The AC for this is deterministic (AC-PN-008 uses a synchronized-barrier fixture, NOT millisecond polling).

### §B.3 Resolved decisions (iter-2 fold — MP-7 clearance)

The three open-question markers raised at iter-1 are RESOLVED at the SPEC's own proposed defaults (user-endorsed "proceed"; override window preserved via config where stated):

1. **Staleness threshold** (REQ-PN-012 / AC-PN-012): hard-coded default `N=3` sync cycles behind HEAD. Overridable via the `navigator.staleness_cycles` config key (NOT an in-SPEC variable) — defaults live at the framework config layer, not the SPEC layer.
2. **SessionStart hint token budget** (REQ-PN-009 / AC-PN-009): hard-coded ceiling of 500 tokens (frontier + next task + one link to the full file).
3. **Hook vs skill-mode overlap** (REQ-PN-009 vs REQ-PN-011): both surfaces KEPT — see §B.1 for the differential-use rationale (ambient auto-brief vs on-demand full brief).

No live open-question markers remain in this artifact set.

## §C. Integration map (constraint #2)

### §C.1 Component ↔ primitive

| Navigator component | Existing primitive | Relationship |
|---------------------|--------------------|-------------|
| `navigator.md` entry point | `/moai project` skill (`moai-workflow-project/SKILL.md`) | EXTENDS — generated by the skill, sits alongside product/structure/tech.md |
| Regeneration trigger (sync) | `/moai sync` → manager-docs | CHAINS INTO — manager-docs invokes the skill's Navigator phase before sync-commit |
| Regeneration trigger (on-demand) | `/moai project` | EXTENDS — single invocation regenerates product + structure + tech + Navigator |
| SessionStart ambient auto-brief | `handle-session-start.sh` (+ Go `session_start.go`) | ADDS a sibling hook (`handle-session-start-navigator.sh`) emitting additionalContext — does not modify the existing hook |
| `capability-map.md` rows (spec-id, status) | SPEC frontmatter (`.moai/specs/SPEC-*/spec.md`) | READ-ONLY consumer — does not modify SPEC frontmatter |
| `progress-map.md` rows (frontier, last-commit) | `.moai/specs/SPEC-*/progress.md` | READ-ONLY consumer — does not modify progress.md |
| Provenance (commit-sha, captured-at) | `git log` | READ-ONLY consumer — single source of truth for "when did this last change" |
| `@MX` tag corpus | `@MX` protocol (`.claude/rules/moai/workflow/mx-tag-protocol.md`) | NOT USED by 001 — reserved for SPEC-003 auto-derivation |
| Design-intent diff (P2) | `product.md` / `structure.md` / `tech.md` | NOT USED by 001 — consumed by SPEC-002 `--audit` |
| Session handoff | `session-handoff.md` | COMPLEMENTS — handoff resumes ONE SPEC; Navigator reorients to the WHOLE project. A paste-ready resume MAY reference `navigator.md` as the first read. |
| Auto-memory | `MEMORY.md` + `feedback_*.md` | COMPLEMENTS — memory is past-tense lessons; Navigator is present-tense project state. |
| `/moai goal` | goal-directive.md | COMPLEMENTS — goal is within-session continuation; Navigator is cross-session reorientation. |
| `/moai codemaps` | `.claude/skills/moai/workflows/codemaps.md` | SIBLING — codemaps is the CODE structural map; Navigator is the FEATURE/status map. SPEC-003 will bridge them. |

### §C.2 Cross-subcommand consultation map (Navigator as shared read primitive) [iter-2 directive C]

The Navigator is a PROJECT-LEVEL SHARED READ PRIMITIVE, not a siloed `/moai project` feature. Ownership split: **`/moai project` OWNS maintenance** (write / regenerate / audit); other subcommands **CONSUME (read)** at defined orientation phases. The consultation points are opt-in per subcommand ("use where needed"), NOT forced on every call — each subcommand's skill/workflow gets a READ instruction ("consult `.moai/project/navigator/navigator.md` at <phase>"), and the subcommand decides whether the consultation is relevant to the current invocation.

| Subcommand | Phase where it consults | What it reads | What it gains |
|------------|-------------------------|---------------|---------------|
| `/moai plan` | Phase 1 context-load | the brief | Scopes new SPEC against current frontier; prevents duplicate SPECs (REQ-PN-017 / AC-PN-017) |
| `/moai run` | start-of-run | the brief + owning SPEC's `progress-map.md` row | Implementing agent oriented before first action — no re-derivation from many files (REQ-PN-018 / AC-PN-018) |
| `/moai sync` | sync tail | WRITES (regenerates Navigator) | Navigator stays current; staleness window ≤ 1 sync cycle (REQ-PN-003) |
| `/moai project --brief` / `--audit` | on-demand | full brief / drift report | Explicit re-orientation or drift audit (REQ-PN-011 / SPEC-002) |
| SessionStart hook | every session | ambient auto-brief (≤500 tokens) | Zero-touch re-orientation at session start (REQ-PN-009) |

**Mechanism**: each consuming subcommand's skill/workflow body gets a single READ instruction pointing at `.moai/project/navigator/navigator.md` (and optionally `progress-map.md`). NO new Go CLI subcommand — this composes with the existing mechanism decision (§B.1). The READ instruction is opt-in per subcommand: it fires only at the named phase, never on every tool call.

**Boundary preservation (vs LSEL)**: the read-side integrations touch ONLY `.moai/project/navigator/` and `.moai/specs/*/progress.md` — they do NOT read or write any LSEL surface. REQ-PN-016 (non-overlap) is intact. Orientation ≠ harness self-evolution.

## §D. Template mirror plan (constraint #1)

Every new file under `.claude/` or `.moai/` is mirrored to `internal/template/templates/` and regenerated via `make build` (go:embed). Template-neutrality (CLAUDE.local.md §25.1) applies to each: no SPEC IDs, no REQ tokens, no internal dates, no commit SHAs (C2/C3/C7 forbidden classes).

| Template path (source of truth) | Local mirror (rendered) | Action |
|----------------------------------|-------------------------|--------|
| `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` | `.claude/skills/moai-workflow-project/SKILL.md` | EDIT — add Navigator phase + `--brief` mode |
| `internal/template/templates/.claude/skills/moai-workflow-project/references/navigator.md` | `.claude/skills/moai-workflow-project/references/navigator.md` | NEW — regeneration procedure (Level 3 reference) |
| `internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh` | `.claude/hooks/moai/handle-session-start-navigator.sh` | NEW — SessionStart hint hook (fail-open, time-boxed) |
| `internal/template/templates/.claude/settings.json.tmpl` | `.claude/settings.json` | EDIT — register the new SessionStart hook in the SessionStart hooks array |
| `internal/template/internal_content_leak_test.go` | (test, runs in CI) | EDIT — extend forbidden-token sentinel for `SPEC-PROJECT-NAVIGATOR-` |

**Note on `.moai/project/navigator/`**: these files are GENERATED at runtime in each user project, not scaffolded. The template ships no fixed content there (would violate neutrality — a generated file's content depends on the user's SPEC registry).

## §E. Boundary vs SPEC-LSEL-LOCAL-EVOLUTION-001

**Navigator owns**: project-scoped, present-tense state — "what does the project look like right now?" Inputs: SPEC registry + git log + (for 002) design docs. Outputs: `.moai/project/navigator/*.md` + `.moai/state/navigator/*`. Touches ONLY project-scoped artifacts.

**LSEL owns**: harness self-evolution — "how should the harness itself refine based on observation?" Inputs: `.moai/lessons-inbox.jsonl` + usage-log. Outputs: candidate `feedback_*.md` topics + proposed edits to the 6 evolvable surfaces under `hns-lsel-*` + `CLAUDE.local.md`. Touches ONLY user-owned harness surfaces.

**Why they don't overlap**:

| Axis | Navigator | LSEL |
|------|-----------|------|
| Question answered | "what is the project?" | "how does the harness improve?" |
| Primary input | SPEC registry + git log | lessons-inbox + usage-log |
| Primary output | `.moai/project/navigator/*.md` | `feedback_*.md` + `hns-lsel-*` proposals |
| Write surface | project-scoped (`.moai/project/navigator/`) | user-owned harness (`memory/`, `hns-lsel-*`, `CLAUDE.local.md`) |
| Direction | present-tense (rollup of what IS) | past→future (lessons → proposals) |
| Touchpoint | Navigator does NOT consume lessons-inbox; LSEL does NOT consume SPEC registry | One-way future integration: Navigator's `--audit` (SPEC-002) output could *feed* LSEL proposals ("missing SPEC detected → propose creating it"). That is a **future cross-reference**, not in scope for either SPEC today. |

REQ-PN-016 codifies the non-overlap mechanically: Navigator grep must touch no LSEL surface.

## §F. Milestones (priority-ordered; no time estimates)

Order rationale: highest-change-likelihood first. M1 (data model + format) is the most likely to be revised during review; M4 (CI guards) is the most mechanical.

### M1 — Navigator data model + format [Priority High]

- Define the exact markdown schema for `navigator.md`, `capability-map.md`, `progress-map.md` (column order, provenance fields, row ordering).
- Define the empty-project form (REQ-PN-006).
- Define the malformed-frontmatter tolerance behavior (REQ-PN-007).
- Define atomic-rename write strategy (REQ-PN-008 / §B.2).
- Define idempotence contract (REQ-PN-005).
- Deliverable: a documented schema + a fixture project (under `/tmp` or the SPEC's test corpus) that exercises all three forms.

### M2 — Regeneration procedure (skill body) [Priority High]

- Extend `moai-workflow-project/SKILL.md` with a Navigator regeneration phase.
- Add `references/navigator.md` (Level 3) with the step-by-step procedure.
- Wire regeneration into `/moai project` (on-demand) and `/moai sync` (chained via manager-docs).
- Implement inputs: read SPEC registry, parse frontmatter, read `progress.md`, read `git log` for provenance.
- Deliverable: regeneration works end-to-end on a fixture project with ≥3 SPECs across mixed statuses.

### M3 — SessionStart ambient auto-brief hook + `/moai project --brief` skill mode + cross-subcommand READ integration (P1) [Priority High]

- Implement `handle-session-start-navigator.sh` (role: ambient auto-brief; fail-open, time-boxed, ≤500-token bounded additionalContext).
- Register in `internal/template/templates/.claude/settings.json.tmpl`.
- Implement `/moai project --brief` skill mode (load full brief into context).
- Implement staleness signal (REQ-PN-012, default N=3 sync cycles, overridable via `navigator.staleness_cycles`).
- Add the cross-subcommand READ instructions (§C.2): `/moai plan` Phase 1 context-load consults `navigator.md` (REQ-PN-017); `/moai run` start-of-run consults the brief (REQ-PN-018). Opt-in per subcommand, never forced on every call.
- Deliverable: a new session in a fixture project receives the ambient auto-brief automatically; `/moai project --brief` loads the full brief on demand; `/moai plan` and `/moai run` consult the brief at their defined phases.

### M4 — Template mirror + CI guards [Priority Medium]

- Mirror all template edits to `internal/template/templates/` + run `make build`.
- Extend `internal/template/internal_content_leak_test.go` with `SPEC-PROJECT-NAVIGATOR-` sentinel.
- Verify §25 neutrality checklist (5-item pre-commit) on each new template file.
- Verify 16-language neutrality (run generator on a non-Go fixture — e.g. a Python-only project).
- Deliverable: `make build` clean; CI guards green; neutrality checklist passes.

### M5 — Plan-audit gate + Implementation Kickoff Approval [Priority Medium]

- Run plan-auditor on this plan-phase artifact set; resolve findings.
- Implementation Kickoff Approval human gate at plan→run (constraint #6) — the three open questions are already resolved (§B.3); the gate confirms progression mode and any user-override of the hard-coded defaults.
- Deliverable: plan-auditor PASS; user-confirmed progression-mode + gate pass.

## §G. Anti-Patterns (what NOT to do)

- **AP-PN-001** — Generate the Navigator into a fourth/fifth top-level file. Stick to exactly three (REQ-PN-001).
- **AP-PN-002** — Copy SPEC body content into `capability-map.md` rows (violates REQ-PN-013 non-duplication). Rows are references, not copies.
- **AP-PN-003** — Add a new Go CLI subcommand for `--brief` / `--audit` (violates the mechanism decision §B.1; /moai project is a skill).
- **AP-PN-004** — Make the SessionStart hook a blocking gate (violates REQ-PN-010 fail-open + Advisory-Check Discipline).
- **AP-PN-005** — Hardcode this repo's SPEC IDs or commit SHAs into template-distributed files (violates §25.1 C2/C3/C7).
- **AP-PN-006** — Have Navigator read `.moai/lessons-inbox.jsonl` or any LSEL surface (violates REQ-PN-016 boundary).
- **AP-PN-007** — Regenerate the Navigator on every PostToolUse edit (wrong cadence — regeneration is a sync-cycle concern, not per-edit; violates the integration map §C).

## §H. Cross-References

- `research.md` (this SPEC dir) — web grounding + prior-work review.
- `acceptance.md` (this SPEC dir) — Given-When-Then ACs for every REQ.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` emitted here.
- `.claude/rules/moai/workflow/spec-workflow.md` — Plan-Run-Sync chain this SPEC plugs into.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — provenance attribution rationale (REQ-PN-002).
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — fail-open + time-boxed SessionStart (REQ-PN-010).
- CLAUDE.local.md §2 (Template-First), §25 (template-neutrality) — constraint #1.
