# plan.md — SPEC-CODEX-DUAL-AGENTS-001 (Codex Dual Harness M5)

> Plan-phase implementation plan. Tier M. Mapping table (§A.3) is a first-class deliverable
> named by the card. Milestones are ordered by decision-reversibility: the decisions most
> likely to change come first.
> Errata (updated 2026-08-22): §A.3 row 9 disposition corrected by run-phase measurement —
> `mcp_servers` emits as the `[mcp_servers.moai]` table, not the array literal (MS3b,
> commit e6c2239e5; spec.md R-009 carries the same errata).

## §A Context

### §A.1 Problem

The 11 retained agent definitions exist today only as Claude Code `.md` files in the template
tree (`internal/template/templates/.claude/agents/moai/`). Codex (codex-cli 0.147.0, verified by
M0/t91) loads agent definitions from `.codex/agents/*.toml`. Hand-maintaining a second,
parallel set of 11 TOML files would guarantee drift. M5 introduces a neutral definition layer +
a deterministic emitter so both forms are generated from one source, with the `.md` side
byte-frozen (regression ban).

### §A.2 Verified agent inventory (ground truth = TEMPLATE tree, read 2026-08-22)

> The orchestrator's delegation table contained errors; this table was re-verified against
> `internal/template/templates/.claude/agents/moai/*.md` (grep of `model|effort|skills|hooks`
> lines). FILE WINS. The 6 local-copy drifts are listed in §B.1.

| agent | model | effort | skills preload | permissionMode | memory | hooks | carries `Agent` | carries `mcp__moai__*` |
|---|---|---|---|---|---|---|---|---|
| builder-harness | inherit | medium | moai-foundation-cc | bypassPermissions | **user** | no | no | no |
| e2e-tester | inherit | low | moai-workflow-testing | default | project | no | no | no |
| manager-design | inherit | medium | moai-domain-frontend | **acceptEdits** | project | no | no | no |
| manager-develop | inherit | high | moai-foundation-core | bypassPermissions | project | **yes** | no | yes (3) |
| manager-docs | inherit | **low** | moai-foundation-core | bypassPermissions | project | **yes** | no | yes (2) |
| manager-git | **sonnet** | low | moai-foundation-core | bypassPermissions | project | no | no | no |
| manager-lead | inherit | xhigh | **moai-foundation-core, moai-workflow-project (2)** | bypassPermissions | project | no | **yes** | yes (2) |
| manager-spec | inherit | high | **moai-foundation-core, moai-workflow-spec (2)** | bypassPermissions | project | **yes** | no | yes (3) |
| plan-auditor | inherit | high | *(none — no skills field)* | default | project | no | no | yes (5) |
| super-advisor | inherit | high | moai-foundation-core | **plan** | project | no | no | yes (11) |
| sync-auditor | inherit | high | moai-foundation-quality | **plan** | project | **yes** | no | yes (5) |

Effort ladder in use: `low` (e2e-tester, manager-docs, manager-git), `medium` (builder-harness,
manager-design), `high` (manager-develop, manager-spec, plan-auditor, super-advisor,
sync-auditor), `xhigh` (manager-lead). All 11 carry `Bash`. 7 of 11 carry `mcp__moai__*` tools
(20 distinct of the 21-tool server — `goal_arm` is the only absent tool, matching its
orchestrator-only wiring per `.claude/rules/moai/core/moai-mcp-tools.md`).

`.md` structure contract (from full reads): YAML frontmatter in field order
`name, description (block scalar |), tools (CSV string), model, effort, color, permissionMode,
memory, skills (YAML array, optional), hooks (optional)` — then the body (the agent prompt,
plain markdown). The neutral parser must handle the block-scalar description and the CSV tools
line.

`Explore` (Anthropic built-in) has no `.md` in this tree — it appears only as a correspondence
NOTE in §A.3, never as an emitted file.

### §A.3 Tools→Codex mapping table (FIRST-CLASS DELIVERABLE)

Codex agent-TOML surfaces available (per M0 probe + optional-field list): required `name`,
`description`, `developer_instructions`; optional `model`, `model_reasoning_effort`,
`sandbox_mode`, `mcp_servers`, `skills.config`. Permission concepts are NOT 1:1. Per semantic
class:

| # | Semantic class (Claude tokens) | Codex surface | v1 decision | Enforcement mechanism | Status |
|---|---|---|---|---|---|
| 1 | Core file read (`Read, Grep, Glob`) | built-in; no field | no field emitted | `sandbox_mode` consequence | implied by t91 §2 (delegation ran) |
| 2 | Core file write (`Write, Edit`) | built-in; no field | Read-vs-Write distinction NOT mechanically preserved → **documented drop**: Codex sandbox is workspace-level, not tool-level | `sandbox_mode`; body prose already disciplines auditors | same |
| 3 | Shell (`Bash` — all 11 agents) | built-in exec | `sandbox_mode` value emitted per manifest default for all 11 (a read-only sandbox would break `go test`-style verification work the bodies mandate) | `sandbox_mode` | field works (M0); **value set unmeasured → probe P-01** |
| 4 | Web (`WebFetch, WebSearch` — manager-docs, manager-spec, super-advisor, builder-harness) | no per-agent field; web access is a global Codex feature/config | **documented drop** at per-agent level; note: depends on user's global codex config | none per-agent | unmeasured → optional probe |
| 5 | Task list (`TaskCreate, TaskUpdate, TaskList, TaskGet` — all 11) | no known equivalent | **documented drop** with rationale (Claude task-tool family; body prose degrades gracefully — workflows fall back to prose reporting) | none | unmeasured → optional probe |
| 6 | Skill loader (`Skill` + `skills:` preload) | `skills.config` (value set unmeasured) | **deferred to M1** — no field emitted in v1 (M1 owns skills canonicalization to `.agents/skills`; t91 §4 verified the scan path + symlinks) | M1 seam | unmeasured → optional probe P-05 |
| 7 | Subagent spawn (`Agent` — manager-lead only) | delegation exists on Codex (t91 §2: `collaborationspawn_agent`), but per-agent tool grant is not expressible in agent TOML | **documented drop** at field level; lead semantics survive in body prose; internal tool name `collaboration*` has low version stability (t91 §2) | body prose | unmeasured → optional probe |
| 8 | Design sync (`DesignSync` — manager-design only) | none | **documented drop** (Claude-specific MCP-backed tool) | none | — |
| 9 | moai MCP (`mcp__moai__*`, 20 distinct tools across 7 agents) | `mcp_servers` | `[mcp_servers.moai]` table (`command = "moai"`, `args = ["mcp-server"]`) on the 7 agents carrying any `mcp__moai__*` token — NOT the array form `mcp_servers = ["moai"]`, which codex-cli 0.147.0 rejects ("invalid type: sequence, expected a map"; run-phase measurement MS3b, commit e6c2239e5); **per-tool filtering inside one MCP server treated as unavailable → documented drop** (server exposes all 21; the body prose already names which tools the agent uses — prose-level discipline) | server-level grant; NOTE: non-interactive `codex exec` MCP calls require approval-policy handling (t91 §5) — M4 wiring concern | server + 21 tools **measured** (t91 §5); table shape **measured** (MS3b, e6c2239e5); per-agent filtering unmeasured → optional probe P-06 |
| 10 | Effort (`effort:` frontmatter) | `model_reasoning_effort` | identity mapping proposed (`low→low, medium→medium, high→high, xhigh→xhigh`) — preserves relative ordering; direction justified by token-identity where both ladders share names; **never silently downgrade**: unmapped value blocks emission | emitter enum (fail-closed) | field works (M0); **enumeration unmeasured → probe P-02** |
| 11 | Model (`model:` — `inherit` ×10, `sonnet` ×1 manager-git) | `model` | **omit for all 11 in v1** (inherit Codex default); manager-git's `sonnet` pin = documented drop (a Claude alias is not a Codex model id; its cost intent already travels via `effort: low`) | omission = default | **omission semantics unmeasured → probe P-03** |
| 12 | Frontmatter `hooks:` (manager-develop, manager-spec, manager-docs, sync-auditor) | no per-agent hooks surface; Codex hooks are project-level `.codex/hooks.json` | **documented drop** → M3 seam (M3 owns the adapter incl. the `PostToolUse`+`collaboration*` redesign per t91 §8) | M3 | hooks layer measured (t91 §6) |
| 13 | `memory:` / `color:` / `permissionMode:` | no equivalent | **documented drop** (permission semantics live in `sandbox_mode` + Codex approval policy; memory/color are UI/session state) | none | — |
| 14 | Correspondence NOTE: `Explore` (Anthropic built-in, no `.md`) | Codex built-in explorer | no file emitted; recorded in the mapping manifest as a correspondence note only | — | — |

**Ship-omitted fallback rule (card constraint 3)**: any optional field whose value set is not
probe-confirmed ships OMITTED (Codex default inheritance) rather than guessed — because an
unvalidated value is silently ignored (t91 §1) and a wrong-but-accepted value is worse than an
absent one.

### §A.4 Unmeasured semantics → run-phase probes + recorded decisions

| Probe | Question | Method (t91 §9 pattern) |
|---|---|---|
| P-01 | `sandbox_mode` allowed value set | emit probe TOMLs through the isolated-`CODEX_HOME` harness with candidate values; observe acceptance/delegation |
| P-02 | `model_reasoning_effort` enumeration | same, per candidate value |
| P-03 | `model` field omission semantics; arbitrary-string acceptance | omit on one probe, set a bogus string on another; compare behavior (silent-ignore hazard applies) |
| P-04 | does `.codex/agents/` scan subdirectories (`moai/` subdir) or flat only? | place identical probe in `.codex/agents/moai/x.toml` vs `.codex/agents/x.toml`; observe which loads |
| P-05 | `skills.config` value set (M1-deferred; optional) | probe only if cheap; otherwise leave to M1 |
| P-06 | per-agent tool filtering inside one MCP server (optional) | inspect whether agent TOML restricts exposed MCP tools |

**Recorded decisions (lead-ratified 2026-08-22; each probe above remains a run-phase
confirmation task, not an open question):**

- **DECIDED — sandbox_mode value set (confirms via P-01)**: probe P-01 first; if any emitted
  value is unconfirmed, omit the `sandbox_mode` field entirely in v1 — never guess (one bad
  key can kill the whole file — lead-cited measurement, 2026-08-22 ratification).
- **DECIDED — model_reasoning_effort enumeration (confirms via P-02)**: probe P-02 first;
  the identity mapping is emitted only if every emitted value confirms; otherwise omit the
  field.
- **DECIDED — model field semantics (confirms via P-03)**: omit `model` on all 11 agents
  (manager-git's `sonnet` pin = documented drop).
- **DECIDED — .codex/agents/ layout (confirms via P-04)**: subdirectory layout preferred;
  flat + `moai-` filename prefix fallback if P-04 shows no subdir scan.

### §A.5 Design decision — shape of the neutral layer (highest-change-likelihood decision)

**Option A (LEAD-APPROVED 2026-08-22) — `.md` IS the neutral core + Codex mapping manifest.**
The existing template `.md` files remain the single source for name/description/model/effort/
skills/tools/body; a new machine-readable manifest (`agents-codex.yaml`, embedded in the emitter
package) owns the Codex deltas (class-level mapping table, per-agent sandbox/effort overrides,
layout knob). `.md` publication is **identity** (the artifact is its own source); `.toml`
publication is a transform of (`.md` × manifest).

- Why: the regression ban is enforced **by construction** (the emitter never re-renders the
  `.md`); zero migration for 11 files; maintainers keep editing `.md` directly (existing
  muscle memory, Template-First rule, mirror-parity tests); 2 committed files per agent
  (source `.md` + generated `.toml`) instead of 3.
- Honest reading of "both are GENERATED from the neutral layer": the `.md` artifact is the
  neutral layer itself, so no second representation of it exists to drift. This satisfies the
  single-source intent (no independently hand-maintained duplicates); it does not introduce a
  symmetric re-render. **Lead-approved (2026-08-22)** — the literal symmetric-generation
  reading (Option B) is not required; the decision is closed for run-phase.

**Option B (REJECTED for v1) — new neutral namespace re-rendering both.**
Neutral YAML per agent under a new `.agents/` namespace; both `.md` and `.toml` generated and
committed. Rejected because: (a) byte-identity would depend on a custom frontmatter renderer
reproducing hand-authored formatting exactly — test-guarded rather than structural; (b) 3
committed files per agent; (c) maintainer workflow churn across the repo's tooling; (d) M1
will introduce the `.agents/` namespace for skills — agent-definition relocation is cheaper to
fold in AFTER M1 lands, and under Option A it is a single input-path constant in the emitter.

### §A.6 File-level change plan

| Path | Action | Notes |
|---|---|---|
| `internal/template/templates/.claude/agents/moai/*.md` (11) | **UNCHANGED** | byte-identity target; frozen inputs |
| `internal/template/templates/.codex/agents/moai/<name>.toml` (11) | NEW, generated, committed | emitted artifacts; embedded via existing `//go:embed all:templates`; template-neutrality guard applies |
| `internal/template/agentemit/` (package) | NEW | emitter: loader/parser (.md frontmatter contract §A.2), manifest types + embedded `agents-codex.yaml`, deterministic TOML writer, fail-closed validators |
| `internal/template/agentemit/agents-codex.yaml` | NEW | the Codex mapping manifest (class table §A.3 + enums + layout knob); repo-internal build input — NOT under `templates/`, not distributed |
| `internal/template/agentemit/agentemit_test.go` | NEW | golden byte-identity, determinism, fail-closed negatives, embed-FS presence, deploy fixture |
| `Makefile` | MOD (optional) | `agents-emit` regeneration target for maintainers; drift guard is the go test itself (regenerate in-memory, byte-compare committed) |
| deploy code | expected UNCHANGED | `.codex/` rides the existing template mirror deployment; verified by fixture test in MS3, not assumed |

Technical approach notes (implementation detail, owned by run phase): deterministic TOML writer
with fixed key order; body encoded as a TOML multi-line string (prefer literal `'''` strings —
no escape processing; validator fails closed on unrepresentable content, e.g. a body containing
the delimiter); no TOML library is currently a direct dependency in `go.mod` — either add a
spec-compliant TOML parser for test-side round-trip validation or validate against the
emitter's own grammar plus probe-confirmed parsing by codex-cli itself (decision left to MS1,
fail-safe either way because P-probes parse the real consumer).

## §B Known Issues

1. **Local vs template `.md` drift (pre-existing)**: 6 of 11 local `.claude/agents/moai/` copies
   differ from the template tree (builder-harness, e2e-tester, manager-develop, manager-spec,
   plan-auditor, super-advisor; verified `diff -rq` 2026-08-22). The neutral layer reads the
   **TEMPLATE** tree only. Maintaininess hazard: an editor picking the local copy silently
   targets the wrong source — documented here; the golden test pins the template tree.
2. **Docs-vs-file effort drift (pre-existing)**: `agent-authoring.md` § Effort-Level Calibration
   Matrix documents manager-develop and sync-auditor at `medium` (medium column), but the
   shipped template frontmatter says `high` for both. The FILE is ground truth for this SPEC
   (verified); the rule-doc drift is noted for a separate docs card — NOT fixed here (scope
   discipline).
3. **No TOML dependency in go.mod** (direct): see §A.6 technical notes.
4. **Silent-ignore hazard** (M0): the entire fail-closed validator design (R-008) exists
   because of it; residual risk is a value that Codex ACCEPTS but misinterprets — only probes
   (P-01..P-03) close that.
5. **codex-cli version drift**: measurements are pinned to 0.147.0. The manifest records the
   measured version; probes are re-runnable when codex-cli upgrades.

## §C Pre-flight (run-phase entry checks)

1. `go test ./internal/template/...` green on the base commit (baseline).
2. `make build` succeeds; embedded FS contains `.claude/agents/moai/*.md` (11 files).
3. Probe harness bootstrap: isolated `CODEX_HOME` scratch (t91 §9 pattern), codex-cli 0.147.0
   reachable; user's real `~/.codex` untouched (mtime/hash check per t91 §0).
4. Confirm no concurrent emitter writes (single-writer; emission is offline, committed).

## §D Constraints

Mirrors spec.md §D (regression ban; verbatim body; fail-closed vs silent-ignore;
probe-or-omit; Template-First placement; scope fence M1–M4/M6; template neutrality; t91
measurement baseline). Additional plan-level constraint: emission is offline and deterministic
— no network, no environment capture; the probe harness is the ONLY component that talks to
codex-cli, and it never writes into the template tree.

## §E Self-Verification (plan-phase)

Executed during authoring (evidence in progress.md §E.1):
- SPEC ID regex check via executed Bash — output `PASS`.
- ID uniqueness: only `SPEC-CODEX-PHASE2-001` exists in the CODEX area; no collision.
- Frontmatter validated against the 12-field canonical schema (`.claude/rules/moai/development/spec-frontmatter-schema.md`).
- Agent inventory re-verified against the TEMPLATE tree by grep + full reads (§A.2); 6
  delegation-table corrections recorded.
- M0 facts cross-read from `.moai/reports/t91/README.md` (primary checkout) — cited per
  section, not from memory.

## §F Milestones

**MS1 — Emitter core + neutral-layer contract** (highest decision density: manifest schema,
mapping semantics, frontmatter parse contract, TOML writer). Package `internal/template/agentemit`:
manifest types + embedded `agents-codex.yaml` encoding §A.3; `.md` loader (block-scalar
description, CSV tools, optional skills/hooks); deterministic TOML writer; fail-closed
validators (unknown tool token, unmapped effort, invalid sandbox, unrepresentable body); unit
tests on fixture `.md` files (not yet the real 11). Exit: fixture tests green.

**MS2 — Probes: lock the enums** (P-01 sandbox set, P-02 effort enumeration, P-03 model
omission, P-04 layout; P-05/P-06 optional). Uses the t91 harness pattern; probe TOMLs may be
hand-written (t91 did so) — no dependency on MS1. Updates the manifest to measured enums or
applies the ship-omitted fallback. Exit: every emitted field value is probe-confirmed or
omitted; results recorded for progress.md §E.2.

**MS3 — Mass emission + guards**. Run the emitter against the real 11 template `.md` files;
commit the 11 `.toml` under `templates/.codex/agents/moai/` (layout per P-04); golden
byte-identity test (11/11 sha256), determinism test (emit-twice), no-modification test
(`.md` tree untouched), embed-FS presence test, deploy fixture test (`moai update`-path deploy
into `t.TempDir()` lands `.codex/agents/moai/`), neutrality grep on emitted files. `make
build` + optional Makefile target. Exit: all AC-001..AC-011, AC-013 green.

**MS4 — Close-out**: M4/M1/M3 seam notes finalized (§H), §E.2 evidence assembled (commands +
verbatim outputs), any manifest doc polish. Mechanical; lowest reversibility risk.

## §G Anti-Patterns

- Re-rendering the `.md` from parsed fields (breaks byte-identity; forbidden by R-003).
- Emitting a guessed enum value because "Codex accepted it" — acceptance is NOT validation
  (t91 §1); unconfirmed value → omit.
- Hand-editing a generated `.toml` (drift; regeneration overwrites — the test enforces).
- Reading the local `.claude/agents/moai/` copies as the source (§B.1 — template tree only).
- Growing scope into M4 wiring, M1 skills, or M3 hooks inside M5.
- Adding per-agent overrides to the manifest "for later" with no consumer (YAGNI; the
  override key exists in the schema but v1 ships zero overrides).

## §H Cross-References and Seams

- **M4 seam**: M4 (`moai init --agent` wiring generator) consumes (a) the committed
  `templates/.codex/agents/moai/*.toml` as installable artifacts and (b) the emitter's
  fail-closed validators as its own output-checking layer (t91 §8 M4 row demands generator-side
  whitelist validation — M5's validator is the reusable piece). M5 exposes no CLI.
- **M1 seam**: when skills canonicalize to `.agents/skills`, the mapping manifest's Skill row
  (class 6) switches from "deferred" to an emission rule; no M5 artifact changes.
- **M3 seam**: the hooks drop (class 12) is the input inventory for the hook adapter card.
- t91 report (primary checkout): `.moai/reports/t91/README.md` — §1 silent-ignore, §2
  agents-TOML + collaboration* instability, §5 MCP + approval gate, §8 M-rows, §9 reproduction.
- SPEC artifacts: spec.md (this dir), acceptance.md, progress.md.
- Ownership: spec/plan/acceptance authored by manager-spec; progress.md §E.2/§E.3 belong to
  manager-develop, §E.4 to manager-docs.
