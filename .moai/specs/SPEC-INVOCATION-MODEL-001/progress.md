# Progress — SPEC-INVOCATION-MODEL-001

> Lifecycle: plan → run → sync. This file tracks phase evidence. Plan-phase populates §E.1 only; §E.2/§E.3 are owned by manager-develop (run-phase); §E.4 by manager-docs (sync-phase).

---

## §E.1 Plan-phase Audit-Ready Signal

- **Status**: draft (plan-phase artifacts created 2026-07-01).
- **SPEC ID self-check**: `SPEC-INVOCATION-MODEL-001` decomposition → SPEC ✓ | INVOCATION ✓ | MODEL ✓ | 001 ✓ → PASS (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Artifacts**: spec.md, plan.md, acceptance.md, progress.md (4-file Tier M set).
- **REQ coverage**: REQ-IM-001..014 (5 doctrine + 6 feedback + 3 cross-cutting) → AC-IM-001..022 (AC-IM-021 per-command citation anchor + AC-IM-022 run-phase official-docs verification added per plan-auditor D1).
- **Tier**: M (frontmatter `tier: M` — sets plan-auditor 0.80 threshold; aligns frontmatter with plan/progress).
- **plan-auditor**: iter-1 FAIL 0.81 (no blocking defects; all additive). D1-D7 corrections applied — D5 tier, D1 per-command citations + `/loop` framing correction, D2 orchestrator-direct reframe of REQ-IM-014, D3 2-guaranteed + `go version` best-effort diagnostics, D4 definite loader wiring (`feedbackFileWrapper`/`loadFeedbackSection`/`Loader.Load()`), D6 GEARS Event-driven relabels (REQ-IM-005/009), D7 integration/file-load test for AC-IM-012.
- **Deliverables**: (1) invocation-model alignment doctrine rule; (2) /moai feedback enhancement (diagnostic attach, dup detection, gh-fallback, target-repo config-ization).
- **Design decisions RESOLVED** (all 3 fixed by orchestrator, plan.md §H): (1) rule path = `.claude/rules/moai/workflow/native-invocation-model.md`; (2) config = `.moai/config/sections/feedback.yaml` key `feedback.repository` default `modu-ai/moai-adk`; (3) diagnostic = HYBRID — 2 guaranteed items (`moai version`, `uname`/OS) + `go version` best-effort (tool build-provenance) + optional orchestrator-mediated error context (best-effort). No open questions remain.
- **Milestones**: M1 doctrine rule → M2 config-ization (TDD) → M3 feedback workflow → M4 build+verify.
- **Frontmatter**: 12 canonical fields present; validated against spec-frontmatter-schema.md.
- **Out of Scope**: 5 `### Out of Scope —` sub-headings present in spec.md §E (Axis-A reuse, legacy retirement, runtime enforcement, config schema, feedback UX).

## §E.2 Run-phase Evidence

### AC-IM-022 — Official-docs classification verification (RECORDED BEFORE the M1 doctrine matrix commit)

Per verification-claim-integrity §1.1, the 9-command PROGRAMMATIC/HUMAN-ONLY classification matrix is NOT asserted from memory. It was verified at run-phase by fetching the official Claude Code docs.

- **Command run** (WebFetch unavailable in the isolated run-phase agent context → curl fallback of the same official pages):
  - `curl -sL https://code.claude.com/docs/en/commands` → HTTP 200, 1259176 bytes (the commands reference — `[Skill]`/`[Workflow]` marker source)
  - `curl -sL https://code.claude.com/docs/en/skills` → HTTP 200 (the `skills.md` — prompt-based defining quote + `Skill`-tool bridge statement)
- **Observed evidence** (verbatim extracts):
  - commands reference definition entries: `/code-review [low|medium|high|xhigh|max|ultra] [--fix] [--comment] [target]` **Skill** ; `/simplify [target]` **Skill** ; `/loop [interval] [prompt]` **Skill** ("Omit the interval and Claude self-paces between iterations") ; `/deep-research <question>` **Workflow** ; `/goal [condition|clear]` (no marker) ; `/clear [name]` (no marker) ; `/compact [instructions]` (no marker)
  - `skills.md`: "Unlike most built-in commands, which execute fixed logic directly, bundled skills are prompt-based: they give Claude detailed instructions and let it orchestrate the work using its tools."
  - `skills.md`: "A few built-in commands are also available through the Skill tool, including /init, /review, and /security-review. Other built-in commands such as /compact are not."
  - `skills.md`: "Add disable-model-invocation: true to prevent Claude from triggering it automatically." + "available in every session unless disabled with the disableBundledSkills setting" + "Disable all skills by denying the Skill tool in /permissions".
- **Verified classification** (6 PROGRAMMATIC / 3 HUMAN-ONLY):
  - PROGRAMMATIC bundled skill/workflow: `/code-review`, `/simplify`, `/loop`, `/deep-research`
  - PROGRAMMATIC built-in-exposed-via-Skill-tool: `/security-review`, `/review`
  - HUMAN-ONLY built-in (no Skill-tool bridge): `/goal`, `/clear`, `/compact`
- **DIVERGENCE recorded** (LIVE observation overrides plan.md §D1 provisional table): plan.md provisionally tagged `/security-review` and `/review` as HUMAN-ONLY. The official `skills.md` shows both ARE available through the Skill tool → they are orchestrator-invocable → reclassified PROGRAMMATIC (built-in-exposed sub-case). The `/moai review` per-subcommand note in the doctrine reflects this: its justification rests on Axis A + broader-orchestration composition, NOT on a HUMAN-ONLY premise. This divergence is documented in the doctrine's "Classification-divergence note".
- **`/loop` framing correction**: commands reference confirms native `/loop` is a bundled `[Skill]` that self-paces when the interval is omitted — NOT merely a "fixed time-interval scheduler" as goal-directive.md frames it. The doctrine carries the acknowledge+correct cross-reference (goal-directive.md NOT edited).

### Per-milestone commit SHAs

- M1 (doctrine rule + mirror + progress §E.2 evidence + spec.md draft→in-progress): _<recorded after M1 commit — see §E.3>_
- M2 (feedback config-ization TDD): _<recorded after M2 commit>_
- M3 (feedback workflow enhancement): _<recorded after M3 commit>_
- M4 (build + parity + verification): _<recorded after M4 commit>_

> Run-phase environment note: this run-phase executed in a runtime-materialized L1 isolation worktree (`Agent(isolation: "worktree")`), NOT the main checkout that Section A of the delegation assumed. Commits land on the worktree branch `worktree-agent-abae346a5e201f249`; the orchestrator reconciles them to main (the untracked SPEC plan-dir in main requires the mv-backup merge-unblock procedure per the worktree-untracked-plandir hazard). WebFetch was unavailable in this agent context; the official-docs verification above used the curl fallback of the same pages.

## §E.3 Run-phase Audit-Ready Signal

_<AC matrix summary + build/coverage/lint results recorded at M4 completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
