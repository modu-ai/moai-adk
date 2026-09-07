---
id: SPEC-CODEX-COMMAND-SKILLS-001
title: "Codex Command-to-Skill Publication Emitter — the 16 /moai Commands as codex Skill Artifacts"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template"
lifecycle: spec-anchored
tags: "codex, skills, emitter, commands, dual-harness, template, publication"
tier: M
---

# SPEC-CODEX-COMMAND-SKILLS-001 — Codex Command-to-Skill Publication Emitter

## §A User Story and Background

**User story.** As a codex-harness user of a MoAI-initialized project, I want the 16 `/moai`
slash commands to be reachable from codex through its official skill surface (`$skill-name`
sigil or `/skills` browse), so that the MoAI command catalog is available in both harnesses
without a hand-maintained fork — while the Claude-side slash surface stays byte-identical.

**Background.** codex-cli deprecates `~/.codex/prompts` and names skills as the explicit
replacement ("Use skills for reusable prompts" — custom-prompts docs, surveyed 2026-09-07 in
the t494 doc survey §2; official URLs: `learn.chatgpt.com/docs/custom-prompts`,
`learn.chatgpt.com/docs/build-skills`). Skill invocation is the `$` sigil or `/skills` browse.
The t494 survey §4 additionally confirms the scan path: codex scans REPO-scope
`$REPO_ROOT/.agents/skills/`, and `internal/template/skill_mirror.go:52` already creates
`.agents/skills/<name>` mirrors for the 34 deployed skills — so the *skills* path is already
wired; only the *commands* are missing from it.

This SPEC builds the missing piece: an emitter that **publishes** (not converts) the 16
`/moai` command sources (15 `.md.tmpl` + `todo.md` under
`internal/template/templates/.claude/commands/moai/`) as codex skill-shaped artifacts
(`SKILL.md` with `name`/`description` frontmatter). The structural sibling is the
`internal/template/agentemit` dual-publication emitter (SPEC-CODEX-DUAL-AGENTS-001): a
deterministic emit-from-source pattern with golden committed artifacts, an explicit
regeneration verb, a read-only drift check wired ahead of `build`, and a fail-closed
validation contract.

**Measured baseline (this tree, 2026-09-07).**

- 16 command sources exist under `internal/template/templates/.claude/commands/moai/`
  (15 `.md.tmpl` + `todo.md`).
- 34 skill directories exist under `internal/template/templates/.claude/skills/`.
- Name-collision check (card HARD requirement): for all 16 command names, BOTH the bare form
  and the `moai-<command>` derivation are collision-free against the 34 skill directory
  names (measured: 16/16 `clear` on both derivations; command loop output captured in the
  plan-phase record).

**Scope statement.** This SPEC is PUBLICATION ONLY. Command-body neutrality repair (the fact
that every command body invokes the Claude-only `Skill("moai")` tool) belongs to the sibling
card t497 and is explicitly out of scope (§ Out of Scope). The `[[skills.config]]` authoring
surface is NOT touched (the existing `internal/codexwiring/skills.go` consumer stays
read-only). The deprecated `~/.codex/prompts` surface is NOT used (t494 survey §2: rejected
alternative).

## §B Glossary

| Term | Meaning |
|---|---|
| Command source | One of the 16 files under `templates/.claude/commands/moai/` (`<name>.md.tmpl` or `todo.md`), carrying `description`/`argument-hint`/`allowed-tools` frontmatter and a `Use Skill("moai") with arguments: <verb> $ARGUMENTS` body |
| Published skill | The codex skill-shaped artifact derived from one command source: a real directory under the `.agents/skills/` root holding a `SKILL.md` with `name`/`description` frontmatter (the layout codex scans, per t494 §4) |
| Emitter | The deterministic generator that reads the command sources and produces the published skills (structural sibling of `internal/template/agentemit`) |
| Publication (vs conversion) | The command sources are consumed read-only; the emitter never rewrites, reformats, or re-orders them — the Claude slash surface is untouched by construction |
| Derived skill name | The skill directory/frontmatter name, derived as `moai-<command name>` |
| Drift check | The read-only build-time comparison of committed published artifacts against a fresh emission of the command sources (the `agents-emit-check` precedent) |
| Boundary flag | A recorded note that the published body references Claude-only tooling — surfaced, never repaired by this SPEC (t497 owns the repair) |

## §C Requirements (GEARS)

### R-001 — One published skill per command source (Ubiquitous)

The emitter shall produce exactly one codex skill-shaped artifact (a real directory under the
published-skills root containing a `SKILL.md` with `name`/`description` YAML frontmatter) for
each of the 16 command sources, and no artifact for any other input.

### R-002 — Publication, not conversion (Ubiquitous)

The emitter shall treat the 16 command sources as read-only inputs: after any emission run,
the content of `templates/.claude/commands/moai/` shall be byte-identical to its pre-run
state.

### R-003 — Skill identity and description (Event-driven)

**When** the emitter runs, each published skill shall carry (a) a name derived from its
command name as `moai-<command name>`, used consistently as both the directory name and the
frontmatter `name` value, and (b) a `description` equal to the language-neutral English
variant of that command's own `description` frontmatter (extraction mechanism: plan.md M1),
carrying no Go template syntax.

### R-004 — Name-collision refusal (Event-detected) `[card HARD]`

**When** a derived skill name matches an existing canonical skill directory name under
`templates/.claude/skills/`, the emitter shall fail the emission with a diagnostic naming the
offending derived name and the colliding canonical skill, and shall produce no partial
artifact set. The emitter shall never resolve a collision by suffixing, renaming, or
overwriting.

### R-005 — Verbatim body provenance (Ubiquitous) `[card HARD]`

Each published skill body shall be the command source body verbatim — the bytes after the
frontmatter delimiter, untransformed and unrewritten — so that the neutrality state of the
published body is exactly the neutrality state of the command body. The emitter shall record
a boundary flag per published skill stating that the body references Claude-only tooling,
instead of repairing the body (the repair is owned by the sibling card t497).

### R-006 — Emission layout and coexistence (Ubiquitous)

The published skills shall land under the same `.agents/skills/` root the existing skill
mirror occupies (`templates/.agents/skills/moai-<command>/SKILL.md` in the source tree; the
deployed equivalent in a user project), as real directories, coexisting with the mirror's
symlink/copy entries for the 34 canonical skills: neither mechanism shall overwrite, remove,
or shadow the other's entries.

### R-007 — Freshness: golden artifacts + drift check (Event-driven)

**When** any command source or the emitter changes, the committed published artifacts shall
be regenerated through an explicit regeneration verb before commit; a read-only drift check,
wired ahead of `build` in the same position as `agents-emit-check`, shall abort the build
with a diagnostic pointing at the regeneration verb when the committed artifacts differ from
a fresh emission. The drift check shall never write.

### R-008 — Idempotent regeneration (Ubiquitous)

Running the regeneration verb twice in a row without any source change shall leave the tree
byte-identical (no timestamps, no reordering, no churn).

### R-009 — Template neutrality (Ubiquitous)

The published artifacts shall satisfy the template neutrality doctrine: no internal SPEC
IDs, no internal dates, no commit SHAs, no macOS-bias paths, no `CLAUDE.local.md` references;
generated-file headers state their generated nature in neutral English (the agentemit
header precedent). All emitted prose (descriptions, headers) shall be English.

### R-010 — Distribution through the existing deploy path (Ubiquitous)

The published skills shall reach a fresh `moai init` project through the existing embed +
deploy machinery, without a new deploy-time code path: a project initialized from the built
binary contains all 16 published skills at the R-006 layout.

### R-011 — Deploy-side user-file safety (Event-detected)

**When** deployment encounters a pre-existing non-managed entry at a published-skill path in
the user project, the deployer shall leave it untouched and report the skip (the
`skill_mirror.go` skip-and-report precedent), rather than overwrite user-owned content.

## §D Acceptance Criteria

The full AC matrix (Given-When-Then, each AC naming its verifying command) lives in
`acceptance.md`. Coverage map: R-001→AC-001; R-002→AC-002; R-003→AC-003, AC-004;
R-004→AC-005; R-005→AC-006, AC-007; R-006→AC-008, AC-009; R-007→AC-010; R-008→AC-011;
R-009→AC-012; R-010→AC-009; R-011→AC-009. (AC-013 is the cross-platform build gate, B1 —
not an R-coverage target.)

## §E Dependencies and Related Work

- **t494** (doc survey, read-only prior art): the deprecation + `$` sigil + `.agents/skills`
  scan-path evidence this SPEC stands on. Survey §7 row C1 (adopted).
- **t497** (sibling card, NOT this SPEC): command-body neutrality repair. This SPEC flags the
  boundary (R-005); t497 repairs the bodies.
- **SPEC-CODEX-DUAL-AGENTS-001** (`completed`): the agentemit precedent this emitter mirrors
  structurally.
- **SPEC-CODEX-SKILL-LOADER-001** (`completed`): measured that `<repo>/.agents/skills/<name>/SKILL.md`
  is actually loaded by codex.

## Out of Scope

### Out of Scope — Command-body neutrality repair (t497)

- Rewriting, annotating, or conditionally transforming any command body so it works without
  Claude-only tooling. The published body is verbatim (R-005); every repair belongs to t497.
- Measuring or fixing `$ARGUMENTS` placeholder semantics on the codex skill surface (custom
  prompts documented `$ARGUMENTS`; the skills docs surveyed in t494 §4 do not document an
  argument placeholder — that gap is a t497 boundary question, not this SPEC's).

### Out of Scope — Other codex surfaces

- `[[skills.config]]` authoring in `~/.codex/config.toml` — registration is by path
  convention, and the existing consumer stays read-only (t494 §4: the table is
  disable-only).
- Publication to `~/.codex/prompts` — deprecated surface, rejected alternative (t494 §2 C2).
- Hooks, agents TOML, plugin packaging, MCP/statusline fields — separate cards in the same
  series.

### Out of Scope — Description authoring

- Re-authoring, translating, or improving the command descriptions. The published
  description is a mechanical extraction of the existing English variant (R-003); new
  description prose is a separate decision.

### Out of Scope — Published-skill lifecycle cleanup

- Removing a published skill whose command source was renamed or retired (the mirror
  precedent: lifecycle management belongs to the clean path, `skill_mirror.go` header note).
  The drift check (R-007) makes the stale artifact visible; its removal is follow-up work.

## §F References

- t494 doc survey: `.moai/reports/t494/codex-doc-survey.md` §2 (custom prompts deprecation,
  `$` sigil) and §4 (`.agents/skills` scan paths, `SKILL.md` layout, `name`/`description`
  frontmatter, `[[skills.config]]` disable-only) — read-only reference from the t494
  worktree; URLs cited there: `learn.chatgpt.com/docs/custom-prompts`,
  `learn.chatgpt.com/docs/build-skills`.
- `internal/template/skill_mirror.go` — existing `.agents/skills` mirror mechanism.
- `internal/template/agentemit/` — the agent dual-publication emitter (golden artifacts,
  drift check, fail-closed contract, Makefile wiring).
- `internal/codexwiring/skills.go` — read-only `[[skills.config]]` consumer (untouched).
- `.github/workflows/template-neutrality-check.yaml` — existing CI guard whose
  `internal/template/templates/**` path filter automatically covers the emitted subtree.

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-09-07 | 0.1.0 | Initial plan-phase draft (card t503, Tier M, Class C) |
