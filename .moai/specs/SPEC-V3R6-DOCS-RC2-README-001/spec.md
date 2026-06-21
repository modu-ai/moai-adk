---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "v3.0.0-rc2 README + CHANGELOG factual-alignment (repo-root docs)"
version: "0.3.0"
status: completed
created: 2026-06-19
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "repo-root docs (README.md, README.ko.md, CHANGELOG.md, CLAUDE.md)"
lifecycle: spec-anchored
tags: "docs, readme, changelog, factual-alignment, v3r6, rc2"
era: V3R6
tier: M
depends_on: []
related_specs:
  - SPEC-V3R6-LIFECYCLE-REDESIGN-001
  - SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
  - SPEC-V3R6-HARNESS-NAMESPACE-V2-001
  - SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
  - SPEC-V3R6-DOCS-V3-README-001
---

# SPEC-V3R6-DOCS-RC2-README-001

## §A. Problem Statement

### A.1 Context

The repo-root user-facing docs (`README.md`, `README.ko.md`, `CHANGELOG.md`, and the `CLAUDE.md` §4 Agent Catalog) have drifted from the authoritative V3R6 ground truth across three surfaces:

1. **Quantitative drift** — agent counts (24 / 26 / 38,700 LOC / 38 packages / 18 languages / 47–52 skills) contradict the current state (8 retained agents, 100K+ lines of Go [graceful-aging; precise figure not hardcoded], 100 packages, 16 languages, 31 moai-* template-managed skills [excluding 2 `harness-moaiadk-*` user-owned], 17 commands).
2. **Catalog drift** — the AI Agent Orchestration Mermaid diagram and the Design System implementer table reference archived `expert-*` agents that were retired in `SPEC-V3R6-AGENT-TEAM-REBUILD-001`.
3. **Version/lifecycle drift** — README.ko.md's "What's New" section is frozen at v2.17.0 (two major versions behind); CHANGELOG carries mis-labeled `[Unreleased] — v2.20.0-rc1` subsections; the 3-phase lifecycle (plan→run→sync, with the Mx-phase retired per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`) is not surfaced.

A read-only analysis + design pass already enumerated the exact drift points (file:line) for each surface. This SPEC converts those findings into GEARS-format requirements. This is a **documentation factual-alignment** SPEC — the deliverable is updated prose/diagrams, not Go code.

### A.2 Why This Matters

- **Unobserved-claim hazard** (per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1): every stale number in the README is an unverified claim about the codebase. "24 specialized agents" is not merely out-of-date prose; it is a factual assertion that a fresh reader will act on (look for 24 agent files, find 7, lose trust in the doc).
- **First-impression surface**: `README.md` is the repo's front door. The hero line and the Key Numbers block are the most-read lines in the project.
- **2-major-version freeze on KO What's New**: the Korean-locale entry point still advertises the v2.17.0 feature set; readers arriving from Korean-language channels see a project frozen in time.

### A.3 Scope Boundary

This SPEC covers the **repo-root project-owned docs** only:
- `README.md` (EN)
- `README.ko.md` (KO)
- `CHANGELOG.md`
- `CLAUDE.md` §4 (builder-harness path correction only)

It does NOT cover the `docs-site` (adk.mo.ai.kr), `.moai/docs/` doctrine files, or the `internal/template/templates/` SSOT — those are addressed in §H and the sibling SPEC.

---

## §B. Ground Truth (Authoritative Baseline)

The following values are the **authoritative baseline** for this SPEC. Every REQ below that asserts a numeric/categorical value MUST be attributed to one of these rows.

| Surface | Baseline value | Authority source | Notes |
|---------|----------------|------------------|-------|
| Version | `v3.0.0-rc2` | `.moai/config/sections/system.yaml` + git tag `v3.0.0-rc2` | NO `v3.0.0` stable tag exists. Docs MUST say `v3.0.0-rc2`. |
| Agents (retained) | 8 total | `.claude/agents/moai/` (7 files) + Anthropic built-in `Explore` | manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness + Explore. |
| Agents (archived) | 12 | `.moai/backups/agent-archive-2026-05-25/` | NOT to be surfaced in catalog/Mermaid. |
| builder-harness path | `.claude/agents/moai/builder-harness.md` | filesystem (verified) | The `.claude/agents/builder/` directory does NOT exist. |
| Commands | 17 | `internal/cli` command registry | brain clean codemaps coverage design e2e feedback fix gate harness loop mx plan project review run sync. NO `/moai db` slash command exists. |
| Skills | 31 (moai-* template-managed, excluding 2 `harness-moaiadk-*` user-owned) | `find .claude/skills -name SKILL.md -path "*moai*" \| grep -v "harness-moaiadk-" \| wc -l` → 31 | LIVE re-derive 2026-06-19 (iter-3): `*moai*` path match returns 33, of which 2 are `harness-moaiadk-best-practices` + `harness-moaiadk-patterns` (user-owned per `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`). `moai-design-system` removed by `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` (post-authoring), which is why the iter-1 count of 32 dropped to 31. README surfaces the **31** template-managed count and MUST state the exclusion explicitly so the number is reproducible. |
| Go LOC | graceful-aging ("100K+ lines"; precise figure NOT hardcoded) | `find . -name "*.go" -not -path "./vendor/*" -not -path "./internal/template/embedded.go" -not -name "*.pb.go" -not -name "zz_*.go" \| xargs wc -l \| tail -1` returns an inflated count (iter-3 LIVE re-measurement 2026-06-19 = 191,248; the pipeline over-counts because the exclusion set is incomplete) | Prefer graceful-aging phrasing ("100K+ lines of Go across 100+ packages"); do NOT hardcode the precise figure in README (LOC drifts on every commit — see §E.3). Earlier drafts cited precise figures (~193,616 / ~198,945) that the §B pipeline cannot reproduce deterministically; iter-3 (D3) drops the precise LOC claim from the Ground Truth to align the anti-unobserved-claim SPEC's own §B with the no-unobserved-claim invariant (per `verification-claim-integrity.md` §2 attribution). |
| Go packages | 100 | `go list ./...` count | Prefer "100+ packages" phrasing. |
| Languages | 16 | `CLAUDE.local.md §15` + README L80/514 | go python typescript javascript rust java kotlin csharp ruby php elixir cpp scala r flutter swift |
| GLM model | `glm-5.2[1m]` (high) | `internal/template/templates/.moai/config/sections/llm.yaml` (TEMPLATE SSOT) | Local llm.yaml is stale (glm-5.1); docs follow TEMPLATE SSOT. |
| Lifecycle | 3-phase (plan→run→sync) | `SPEC-V3R6-LIFECYCLE-REDESIGN-001` (code landed, spec draft) | Mx-phase (4th) RETIRED. `sync_commit_sha` lives in `§E.4`. |
| Era system | 5 buckets + grandfather | `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001` (completed) | V2.x / V3R2-R4 / V3R5 / V3R6 / unclassified. |
| Harness namespace | `moai-*` / `moai-harness-*` / `moai-meta-harness` = template-managed; `harness-*` = user-owned | `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (completed) | CITE THIS ONE. Do NOT cite superseded `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`. |
| Dynamic workflows | Committed | `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` (completed) | Claude Code v2.1.154+ primitive + `/effort ultracode`. |

---

## §C. Drift Inventory (Evidence — file:line)

Every REQ below cites its drift evidence from this inventory. Line numbers are from the analysis pass against the current repo-root files.

### C.1 README.md (EN) drift points

> **D6 caveat (LIVE re-derive 2026-06-19)**: `README.md:297` already reads `**Total: 8 retained agents** (7 MoAI-custom + 1 Anthropic built-in Explore)`. Any AC that greps the whole file for `8 retained` / `8 (retained )?agents` is therefore **vacuously satisfied today** (returns 3, not 0). Every "new value present" AC below MUST be line-anchored to the specific drift line via `sed -n '<line>p' | grep` so it cannot pass on the L297 token that already exists. This caveat is the root cause of the iter-1 D2 defect.

| ID | File:line | Current (stale) | Target (authoritative) |
|----|-----------|-----------------|------------------------|
| DRIFT-EN-01 | `README.md:40` | "24 specialized AI agents and 52 skills" | 8 retained agents, 31 moai-* skills (excluding 2 `harness-moaiadk-*`) |
| DRIFT-EN-02 | `README.md:62` | "38,700+ lines / 38 packages" | graceful-aging "100K+ lines / 100+ packages" (precise LOC NOT hardcoded; iter-3 §B pipeline over-counts, see §B Go LOC row) |
| DRIFT-EN-03 | `README.md:64` | "26 agents + 47 skills", "18 languages" | 8 agents, 31 skills, 16 languages |
| DRIFT-EN-04 | `README.md:262` | "delegates to 24 specialized agents" | "delegates to 8 retained agents" (NOTE: L297 already has a *different* "8 retained agents" token — the AC MUST anchor on L262 specifically, see D6 caveat above) |
| DRIFT-EN-05 | `README.md:264-286` | AI Agent Orchestration Mermaid (fence L264, close L286) uses bare category labels `Manager (8)`, `Expert (8)`, `Builder (3)`, `Evaluator (2)`, `Design System (4+1)` + node body `backend · frontend · security · devops<br/>performance · debug · testing · refactoring` referencing archived `expert-*` agents | Rewrite to show only the 8 retained agents (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness, Explore); bare category labels + archived-agent node bodies absent |
| DRIFT-EN-06 | `README.md:451-478` | Plan→Run→Sync pipeline (no lifecycle note) | Add 3-phase lifecycle note (Mx retired) |
| DRIFT-EN-07 | `README.md:921,939,969` | Design System implementer table cites `expert-frontend` | manager-develop / harness-generated specialist |
| DRIFT-EN-08 | `README.md:1111-1210` | Full `/moai db` section (slash command does not exist) | Remove; ≤1-line pointer to `moai hook db-schema-sync` |
| DRIFT-EN-09 | `README.md:1235,1245,1281,1289` | statusline FAQ illustrative versions stale | Opus 4.7+ / CC 2.1.17x+ OR mark explicitly illustrative |

### C.2 README.ko.md (KO) drift points

| ID | File:line | Current (stale) | Target |
|----|-----------|-----------------|--------|
| DRIFT-KO-01 | README.ko.md hero / Key Numbers equivalents | mirrors of DRIFT-EN-01..04 | 8 agents / 100K+ LOC / 16 langs / 31 skills |
| DRIFT-KO-02 | `README.ko.md:46-90` | "## v2.17.0의 새로운 기능" section | REWRITE to v3/V3R6 generation |
| DRIFT-KO-03 | `README.ko.md:58` | "my-harness-*" namespace | "moai-* (template-managed) vs harness-* (user-owned)" |
| DRIFT-KO-04 | README.ko.md Mermaid / Design System / `/moai db` equivalents | mirrors of DRIFT-EN-05,07,08 | mirror fixes |
| DRIFT-KO-05 | README.ko.md language count equivalent | 18 languages | 16 languages |

### C.3 CHANGELOG.md drift points

| ID | File:line | Current (stale) | Target |
|----|-----------|-----------------|--------|
| DRIFT-CL-01 | CHANGELOG.md top `## [Unreleased]` block | Generic unreleased | Promote to `## [v3.0.0-rc2] — <date>` reflecting rc2 cohort |
| DRIFT-CL-02 | CHANGELOG.md `## [Unreleased] — v2.20.0-rc1` subsection headers | LIVE 2026-06-19: `grep -cE '^##.*v2\.20\.0-rc1' CHANGELOG.md` returns **4** heading lines (L378, L487, L507, L527); total `v2.20.0-rc1` mentions = 28 (4 headings + 24 body references). Scope decision (D4): **remove the 4 mis-labeled HEADING lines** (consolidate their content under the correct `## [v3.0.0-rc2]` section or a fresh `## [Unreleased]`); body mentions inside consolidated content are rewritten as part of the merge. AC-CL-002 grep predicate matches the heading-only form `^##.*v2\.20\.0-rc1` (NOT bare `v2.20.0-rc1`, which would over-match body prose). |

### C.4 CLAUDE.md drift points

| ID | File:line | Current (stale) | Target |
|----|-----------|-----------------|--------|
| DRIFT-CLAUDE-01 | `CLAUDE.md` §4 Agent Catalog | ".claude/agents/builder/builder-harness.md" | ".claude/agents/moai/builder-harness.md" |

---

## §D. Requirements (GEARS)

### D.1 Agent-count reconciliation (EN)

**REQ-EN-001** — The `README.md` hero line (DRIFT-EN-01, `README.md:40`) **shall** state the retained agent count as 8 (7 MoAI-custom at `.claude/agents/moai/` + 1 Anthropic built-in `Explore`) and the skill count as **31 moai-* template-managed skills** (explicitly excluding the 2 `harness-moaiadk-*` user-owned skills per `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`). The bare phrase "31 skills" without the moai-* qualification is insufficient because `find .claude/skills -name SKILL.md -path "*moai*"` returns 33; the exclusion MUST be stated so the number is reproducible.

**REQ-EN-002** — The `README.md` Key Numbers block (DRIFT-EN-03, `README.md:64`) **shall** state 8 agents, 31 moai-* skills (with the harness-moaiadk-* exclusion noted per REQ-EN-001), and 16 languages; the previous "26 agents + 47 skills" and "18 languages" tokens **shall** be absent.

**REQ-EN-003** — The `README.md` delegation prose (DRIFT-EN-04, `README.md:262`) **shall** state "delegates to 8 retained agents" (or semantically equivalent). The "24 specialized agents" token at L262 **shall** be absent. **Because `README.md:297` already contains an unrelated "8 retained agents" token, the AC for this REQ MUST line-anchor on L262** (via `sed -n '262p' | grep`) — a whole-file `grep "8 retained"` would vacuously pass today (D6 caveat).

### D.2 Go-scale reconciliation (EN)

**REQ-EN-004** — **Where** the `README.md` Key Numbers block cites Go scale (DRIFT-EN-02/03, `README.md:62-64`), the docs **shall** prefer graceful-aging phrasing ("100K+ lines of Go across 100+ packages") over a precise re-hardcode of the moment-in-time figure. The precise figure is **NOT** to be hardcoded: the §B pipeline over-counts and is not deterministic (iter-3 LIVE 2026-06-19 = 191,248 via the spec's own pipeline; earlier drafts cited ~193,616 / ~198,945), and LOC drifts on every commit. Dropping the precise LOC claim aligns the anti-unobserved-claim SPEC's own §B with the no-unobserved-claim invariant.

### D.3 Language-count reconciliation (EN)

**REQ-EN-005** — The `README.md` language count **shall** state 16 languages, consistent with `CLAUDE.local.md §15` and `README.md:80/514`; the "18 languages" token (DRIFT-EN-03) **shall** be absent.

### D.4 Mermaid rewrite (EN)

**REQ-EN-006** — **When** the AI Agent Orchestration Mermaid diagram (DRIFT-EN-05, fence `README.md:264`, close `README.md:286`) is rendered, the diagram **shall** show only the 8 retained agents (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness, Explore). The stale bare category labels `Manager (8)`, `Expert (8)`, `Builder (3)`, `Evaluator (2)`, `Design System (4+1)` and the archived-agent node body `backend · frontend · security · devops<br/>performance · debug · testing · refactoring` **shall** be absent from the L264-286 block. (The hyphenated `expert-(frontend|backend|...)` form does NOT appear in the Mermaid block today — it appears only in the Design System table — so the AC MUST grep the ACTUAL stale bare-label form inside the line-anchored block, not the hyphenated form.)

### D.5 /moai db removal (EN)

**REQ-EN-007** — **Where** the `README.md` documents slash commands, the `/moai db` section (DRIFT-EN-08, `README.md:1111-1210`) **shall** be removed; **When** database-schema tooling is mentioned, the docs **shall** provide at most a one-line pointer to the CLI `moai hook db-schema-sync`.

### D.6 expert-frontend replacement (EN)

**REQ-EN-008** — **When** the Design System implementer table (DRIFT-EN-07, `README.md:921/939/969`) names a frontend implementer, the docs **shall** cite `manager-develop` (with `cycle_type` selection) or a harness-generated specialist (per `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`); the archived `expert-frontend` token **shall** be absent.

### D.7 Lifecycle note (EN)

**REQ-EN-009** — **Where** the `README.md` documents the Plan→Run→Sync pipeline (DRIFT-EN-06, `README.md:451-478`), the docs **shall** include a short note that the lifecycle is 3-phase (plan→run→sync) with the former Mx-phase retired per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`. The docs MAY optionally mention dynamic workflows and `/effort ultracode` per `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`.

### D.8 Statusline FAQ refresh (EN)

**REQ-EN-010** — **When** the `README.md` statusline FAQ (DRIFT-EN-09, `README.md:1235/1245/1281/1289`) cites illustrative model/CC versions, the docs **shall** either (a) refresh to current (Opus 4.7+, CC 2.1.17x+) or (b) mark the versions as explicitly illustrative ("example only — substitute your installed version"). The AC for this REQ MUST pin a concrete grep predicate (D7): `grep -cE "explicitly illustrative|example only" README.md` returns ≥1 (option b) OR `grep -cE "Opus 4\.[78]|CC 2\.1\.17" README.md` returns ≥1 (option a) — a bare "M6 captures grep OR prose note" is NOT acceptable.

### D.9 README.ko.md mirror + What's New rewrite (KO)

**REQ-KO-001** — The `README.ko.md` **shall** mirror all quantitative, Mermaid, `/moai db`, and `expert-frontend` fixes from REQ-EN-001 through REQ-EN-008 at the equivalent line locations (DRIFT-KO-01, DRIFT-KO-04, DRIFT-KO-05).

**REQ-KO-002** — The `README.ko.md` "## v2.17.0의 새로운 기능" section (DRIFT-KO-02, `README.ko.md:46-90`) **shall** be rewritten to the v3/V3R6 generation, covering at minimum: 8-agent retained catalog, `glm-5.2[1m]` model, 3-phase lifecycle, CG mode default, dynamic workflows, and `/effort ultracode`.

**REQ-KO-003** — The `README.ko.md` harness namespace line (DRIFT-KO-03, `README.ko.md:58`) **shall** state "moai-* (template-managed) vs harness-* (user-owned)" per `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`; the superseded "my-harness-*" token **shall** be absent.

### D.10 CHANGELOG reconciliation

**REQ-CL-001** — The `CHANGELOG.md` top `## [Unreleased]` block (DRIFT-CL-01) **shall** be promoted to a dated `## [v3.0.0-rc2] — <date>` section reflecting the rc2 V3R6 cohort, covering at minimum: 3-phase lifecycle, harness namespace V2, runtime recovery doctrine, orchestrator interrupt ledger, and `glm-5.2[1m]`.

**REQ-CL-002** — **When** the `CHANGELOG.md` carries mis-labeled `## [Unreleased] — v2.20.0-rc1` subsection headers (DRIFT-CL-02; LIVE count = **4** heading lines at L378/L487/L507/L527), the docs **shall** consolidate those 4 heading lines (and their body content) so that **no `^##.*v2\.20\.0-rc1` heading line** remains. Scope is heading-line removal (NOT bare-token purge): body prose that legitimately references the historical v2.20.0-rc1 milestone inside consolidated content is acceptable; the AC greps the anchored `^##.*v2\.20\.0-rc1` form, not bare `v2.20.0-rc1`.

**REQ-CL-003** — The `CHANGELOG.md` **shall not** contain a `## [v3.0.0]` stable-release section; only `## [v3.0.0-rc2]` exists because no `v3.0.0` stable tag exists (unobserved-claim prevention per `verification-claim-integrity.md` §1.1 surface 3).

### D.11 CLAUDE.md path correction

**REQ-CLAUDE-001** — The `CLAUDE.md` §4 Agent Catalog builder-harness reference (DRIFT-CLAUDE-01) **shall** state the real path `.claude/agents/moai/builder-harness.md`; the stale `.claude/agents/builder/builder-harness.md` token **shall** be absent.

### D.12 Version-wording invariant (cross-cutting)

**REQ-X-001** — **While** any doc touched by this SPEC mentions a version, the docs **shall** use the wording `v3.0.0-rc2` (never `v3.0.0 stable` or bare `v3.0.0` without the `-rc2` suffix), because the `v3.0.0` stable tag does not exist.

### D.13 Superseded-SPEC citation invariant (cross-cutting)

**REQ-X-002** — **When** any doc touched by this SPEC cites the harness-namespace work, the docs **shall** cite `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` and **shall not** cite the superseded `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`.

### D.14 Draft-SPEC phrasing invariant (cross-cutting)

**REQ-X-003** — **Where** any doc touched by this SPEC mentions `SPEC-V3R6-LIFECYCLE-REDESIGN-001` lifecycle mechanics, the docs **shall** describe the 3-phase lifecycle as landed (the run-phase code is merged); **Where** the docs mention `SPEC-V3R6-RULES-SSOT-DEDUP-001` or `SPEC-V3R6-RULES-VERSION-FORMAT-001`, the docs **shall** phrase that work as in-flight (draft), not finalized.

---

## §E. Constraints

- **E.1 Project-owned files**: `README.md`, `README.ko.md`, `CHANGELOG.md`, `CLAUDE.md` are project-owned (not template-managed), so the template-internal-content isolation doctrine (`.moai/docs/template-internal-isolation-doctrine.md` §25) does NOT bind these files. The docs MAY carry project-specific version/SPEC references.
- **E.2 CLAUDE.md template counterpart**: `CLAUDE.md` has a template source at `internal/template/templates/CLAUDE.md`. If `CLAUDE.md` is edited (REQ-CLAUDE-001), the plan.md MUST flag whether the template source needs the same path fix. The plan-phase does NOT modify the template — that decision is deferred to run-phase (see plan.md §G).
- **E.3 Graceful-aging preference**: Where a precise moment-in-time figure could be replaced by a range that ages better ("100K+ lines" vs a precise LOC count), the docs SHALL prefer the range. Rationale: the next drift sweep should not re-fire on every LOC delta, and the §B Go LOC pipeline cannot reproduce a deterministic precise figure (iter-3 D3 finding).
- **E.4 No stable-version claim**: REQ-X-001 is a hard constraint. Asserting `v3.0.0 stable` while only the `rc2` tag exists is an unobserved-claim violation.
- **E.5 Reproduction-first evidence**: Every drift claim in §C cites a `file:line`. Every AC in `acceptance.md` MUST be independently verifiable via grep (`grep -c <stale-token>` returns 0; `grep -c <new-value>` returns ≥1). **Line-anchor rule (D6)**: when a "new value" already exists elsewhere in the same file (e.g., `README.md:297` already says "8 retained agents"), a whole-file `grep` is vacuously satisfied — every such AC MUST either (a) line-anchor via `sed -n '<line>p' | grep` to the specific drift line, OR (b) carry a companion "stale token absent at <line>" AC. The acceptance.md suite has been audited (iter-1 fix) to line-anchor every "new value present" AC where the token pre-exists elsewhere.
- **E.6 Language parity**: EN and KO README files MUST carry the same quantitative values. The KO file is NOT a translation of stale EN — both must reflect the §B baseline.

---

## §F. Non-Functional Requirements

- **F.1 No Go code changes**: This SPEC produces zero `*.go` edits. All deliverables are markdown prose/diagrams in repo-root docs.
- **F.2 No template edits in plan-phase**: `internal/template/templates/**` is NOT modified by this SPEC's plan-phase. Run-phase decision flagged in plan.md §G.
- **F.3 Mermaid validity**: Any rewritten Mermaid diagram MUST render without syntax errors in the GitHub markdown renderer.
- **F.4 Lint cleanliness**: After run-phase, `moai spec lint` on this SPEC MUST return 0 findings.

---

## §G. Success Criteria

- Every stale token enumerated in §C is grep-absent (0 hits) after run-phase.
- Every authoritative value in §B is grep-present (≥1 hit) in the corresponding file.
- The KO "What's New" section reflects the v3/V3R6 generation (no v2.17.0 prose remains as the active section).
- The `v3.0.0-rc2` wording is consistent across all touched files; no `v3.0.0 stable` claim appears.
- `moai spec audit --json` for this SPEC reports era V3R6 with no MUST-FIX drift findings post-close.

---

## §H. Exclusions

### Out of Scope — docs-site (adk.mo.ai.kr)

- The `docs-site` 4-locale documentation (en/ko/ja/zh at adk.mo.ai.kr) is owned by a sibling SPEC. The repo-root `README.md` / `README.ko.md` are distinct from the docs-site sources. This SPEC does NOT touch `docs/`, the docs-site Vercel project, or any docs-site i18n parity work.
- Rationale: docs-site has its own 4-locale parity requirements and Vercel build pipeline; bundling it here would couple two independent release surfaces.

### Out of Scope — v3.0.0 stable release

- This SPEC documents `v3.0.0-rc2` only. Creating a `## [v3.0.0]` stable CHANGELOG section, cutting a `v3.0.0` git tag, or promoting the release candidate to stable belongs to a future release-cut SPEC.
- Rationale: the `v3.0.0` stable tag does not exist (verified against `git tag --list "v3*"`). Asserting stable status would violate the no-unobserved-claim invariant (`verification-claim-integrity.md` §1.1 surface 3).

### Out of Scope — internal/template/templates/** edits

- The template SSOT (`internal/template/templates/`) — including the template counterpart of `CLAUDE.md`, the template `README.md`/`README.ko.md` (if any), and the template `llm.yaml` — is NOT modified by this SPEC. Whether the template source needs the same builder-harness path fix is a run-phase decision documented in plan.md §G.
- Rationale: template-internal-content isolation is a separate doctrine with its own CI guard (`template-neutrality-check.yaml`); conflating project-owned doc fixes with template edits risks leaking project-specific state into the 16-language distribution.

### Out of Scope — .moai/docs/** doctrine files

- The `.moai/docs/` doctrine files (e.g., `git-workflow-doctrine.md`, `harness-namespace-doctrine.md`, `template-internal-isolation-doctrine.md`) are NOT updated by this SPEC. Those files are cross-referenced but not edited here.
- Rationale: doctrine files are owned by their originating SPECs; this SPEC consumes their authority, it does not modify them.

### Out of Scope — CLAUDE.local.md

- `CLAUDE.local.md` is the local dev guide and is explicitly out of template scope (it is listed as a local-only file in `CLAUDE.local.md §2`). It is NOT modified by this SPEC even though it carries the 16-language list that serves as the §B baseline authority.
- Rationale: CLAUDE.local.md is developer-private; the README is the public surface that needs alignment.

### Out of Scope — Go code / test changes

- This SPEC produces zero `*.go` edits and zero test changes. The deliverable is documentation prose and diagrams only.
- Rationale: the drift is factual/doc-level, not behavioral/code-level. Touching Go code would expand the blast radius without addressing the documented drift.

### Out of Scope — superseded SPEC citations

- This SPEC does NOT cite `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (superseded). The harness-namespace authority is `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` only. REQ-X-002 binds this.
- Rationale: citing a superseded SPEC propagates stale doctrine into freshly-aligned docs.

---

## §I. Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 + §5 — no-unobserved-claim invariant + worked example (the §5 defect-claim hazard is the direct rationale for REQ-X-001 and REQ-CL-003).
- `.moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/` — 3-phase lifecycle authority (code landed; spec draft). REQ-EN-009 cites this.
- `.moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/` — era system + grandfather clause SSOT.
- `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-V2-001/` — harness namespace final doctrine. REQ-KO-003 + REQ-X-002 cite this.
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/` — dynamic workflows + `/effort ultracode`. REQ-EN-009 references this.
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/` — 8-agent catalog consolidation (archived 12 agents). Underpins REQ-EN-001/003/006.
- `.moai/specs/SPEC-V3R6-DOCS-V3-README-001/` — sibling/cohort SPEC (docs-site README work); boundary documented in §H.
- `.claude/rules/moai/development/sprint-round-naming.md` — Sprint vs Round terminology (this SPEC is a single-SPEC milestone set, no Sprint/Round split).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era classification SSOT (this SPEC sets `era: V3R6` in frontmatter).

---

## §J. History

- **2026-06-19** — SPEC created (plan-phase, status: draft). Source: read-only analysis + design pass that enumerated the §C drift inventory. Decomposition self-check: `SPEC ✓ | V3R6 ✓ | DOCS ✓ | RC2 ✓ | README ✓ | 001 ✓ → PASS`.
- **2026-06-19 (iter-1 audit fix, v0.2.0)** — Independent plan-audit (iter-1 FAIL 0.74) re-derived all §B numbers LIVE and found 3 BLOCKING + 4 SHOULD-FIX defects. Fixes: D1 skills count 34→32 (exclude 2 `harness-moaiadk-*`, exclusion stated explicitly); D2 AC-EN-003b line-anchored on L262 (L297 already had "8 retained agents" → whole-file grep was vacuous); D3 AC-EN-006a re-anchored on the actual Mermaid bare labels `Manager (8)`/`Expert (8)`/`Builder (3)`/`Evaluator (2)`/`Design System (4+1)` inside L264-286 (the hyphenated `expert-(frontend|...)` form does not appear in the block); D4 v2.20.0-rc1 count corrected ×3→4 headings, AC-CL-002 scoped to `^##.*v2\.20\.0-rc1` heading-line removal; D5 Go LOC ~198,945→~193,616 (LIVE), precise figure dropped in favor of "100K+" graceful-aging phrasing; D6 line-anchor rule codified in §E.5 + §C.1 caveat; D7 AC-EN-010 pinned to concrete grep predicates (`explicitly illustrative|example only` ≥1 OR `Opus 4\.[78]|CC 2\.1\.17` ≥1). REQ-X-002 bans (HARNESS-NAMESPACE-V2 retained, HARNESS-MOAI-NAMESPACE superseded) confirmed intact.
- **2026-06-19 (iter-3 plan-audit fix, v0.3.0)** — Independent plan-audit iter-2 re-audit scored FAIL 0.74 with 2 BLOCKING + 2 SHOULD-FIX defects. Fixes: **D1'** [BLOCKING] skills count 32→31 (LIVE re-derive 2026-06-19 = 31 template-managed / 33 total; root cause = `moai-design-system` removed by `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` post-authoring → the iter-1 count of 32 became stale carry-over in §B Ground Truth, a `verification-claim-integrity.md` §2 attribution violation); all 9+ spec.md occurrences updated (§A.1, §B Skills row, §C.1 DRIFT-EN-01/03, §C.2 DRIFT-KO-01, REQ-EN-001/002). **D2'** [BLOCKING] AC-KO-001b language-count predicate line-anchored on `README.ko.md:111` (the iter-2 D6 discipline was applied to EN but missed on KO — L58/126/560/1215 carry the CORRECT `16개` token, so the KO whole-file grep was vacuously satisfied; the actual drift is `18개` at L111). **D3** [SHOULD-FIX] precise Go LOC figure dropped from §B Ground Truth (the §B pipeline over-counts non-deterministically; iter-3 LIVE = 191,248 via the spec's own pipeline, earlier drafts cited ~193,616 / ~198,945; the README prefers "100K+" graceful-aging phrasing per §E.3 anyway — so the unverifiable precise claim is removed from the anti-unobserved-claim SPEC's own §B). **D5** [SHOULD-FIX] AC-CL-001 internal contradiction resolved (the prior AC simultaneously required `grep → 0` and permitted `grep → ≥1`; split into two binary checks). plan.md + acceptance.md updated in lockstep. Frontmatter status unchanged (draft); version 0.2.0 → 0.3.0.
