---
id: SPEC-V3R6-CLI-AUDIT-001
title: "moai CLI 40+ subcommand inventory + dead command identification + init/update/profile integration analysis (research baseline for Sprint 7 CLI-INTEGRATION-001)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit, research, sprint-2, sprint-7-baseline"
tier: M
related_specs:
  - SPEC-V3R6-LEGACY-CLEANUP-001
  - SPEC-V3R6-LEGACY-CLEANUP-002
---

# SPEC-V3R6-CLI-AUDIT-001 — moai CLI Inventory + Dead Command + Integration Analysis

## §1 Goal

Produce a research-only audit baseline of the moai CLI surface — subcommand inventory, dead-command identification, and `moai init`/`moai update -c`/profile integration analysis — that serves as the input baseline for **Sprint 7 FINAL SPEC-V3R6-CLI-INTEGRATION-001** (deferred per v3.0 roadmap Round 2 re-prioritization, 2026-05-23).

### Why now (Sprint 2 P2 position)

- Sprint 7 CLI-INTEGRATION-001 cannot be scoped without an authoritative subcommand inventory. The CLI has grown organically (110 `AddCommand` invocations across 106 `internal/cli/*.go` non-test files + `cmd/moai/`) and prior memory cites "40+ subcommand" as estimate without enumeration.
- Dead-command suspects (e.g., `harness-observe*` family — 4 hook scripts present, code paths unverified) must be confirmed or refuted before consolidation work begins.
- The `moai init` / `moai update -c` / `moai cc -p <profile>` triad needs integration mapping before Sprint 7 attempts unification — without the map, Sprint 7 will repeat the discovery work.

### Why research-only (no code changes)

- This SPEC is **Tier M research-only**. It produces a written research deliverable (audit report at `.moai/reports/cli-audit/audit-{ISO-DATE}.md`) but does NOT modify any `.go` file, hook script, template, or doc-site content.
- Any consolidation, dead-command retirement, or integration code change is deferred to Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 (the follow-up SPEC this audit feeds).
- The lifecycle distinction matters: research-only SPECs end after manager-develop produces the audit report; no `/moai mx` (annotation) step applies because no source code annotation is added.

### Scope discovery (preliminary, from plan-phase pre-flight)

| Discovery axis | Pre-flight finding |
|----------------|-------------------|
| Non-test Go files in `internal/cli/` | 106 |
| `AddCommand` invocations (`internal/cli/` + `cmd/moai/`) | 110 |
| Hook scripts in `.claude/hooks/moai/` | 32 (incl. 4 `harness-observe*` candidates) |
| `moai ` invocation references (local + template hooks) | 310 |
| Estimated subcommand surface (preliminary) | 40–55 (refined in M1) |

### Dead-command candidates (preliminary, requires M2 verification)

Pre-flight grep surfaced these candidates — **enumeration is NOT exhaustive** (full inventory in M2 deliverable):

1. `moai harness-observe` family (4 hook scripts: `handle-harness-observe.sh` / `handle-harness-observe-stop.sh` / `handle-harness-observe-subagent-stop.sh` / `handle-harness-observe-user-prompt-submit.sh`) — code paths exist (`internal/cli/doctor_harness.go` line 32-line, others) but production usage uncertain.
2. `moai db-schema-sync` (per memory `project_v3r6_sprint2_amr_run_ready` Round 2) — code presence unverified.
3. Other candidates TBD via cross-reference grep in M2.

### Out of scope (deferred to Sprint 7 CLI-INTEGRATION-001)

- Actual deprecation/retirement of dead commands (this SPEC only identifies; Sprint 7 retires)
- Implementation of `moai init` + `moai update -c` + profile unification (this SPEC only maps; Sprint 7 implements)
- Code refactoring of any `internal/cli/*.go` file
- Hook script consolidation or removal
- Template `.claude/hooks/moai/` script changes

---

## §2 EARS Requirements

### REQ-CLA-001 — Subcommand inventory (Ubiquitous)

**The system shall** produce a comprehensive subcommand inventory enumerating every moai CLI subcommand (root + sub-subcommand + flag matrix) reachable through `cmd/moai/main.go` cobra command tree, captured in the audit report at `.moai/reports/cli-audit/audit-{ISO-DATE}.md` under `## §1 Subcommand Inventory`.

**Acceptance**: AC-CLA-001

### REQ-CLA-002 — Dead-command identification (Event-driven)

**When** the subcommand inventory (REQ-CLA-001) is complete, **the system shall** cross-reference each subcommand against (a) `.claude/hooks/moai/*.sh` references, (b) `internal/template/templates/.claude/hooks/*.sh` references, (c) `.claude/skills/moai/workflows/*.md` invocations, (d) `cmd/moai/main.go` registrations, (e) test file invocations under `internal/cli/*_test.go`, and (f) docs-site (`docs-site/content/{ko,en,ja,zh}/**/*.md`) references — classifying each subcommand as **active**, **dead-suspect** (zero non-self references), or **internal-only** (referenced only by hooks/templates, not user-facing).

**Acceptance**: AC-CLA-002

### REQ-CLA-003 — `moai init` / `moai update -c` / profile integration map (Ubiquitous)

**The system shall** produce an integration map at `## §3 Integration Map` of the audit report, documenting (a) `moai init <project>` template deployment + profile selection flow (config files read/written, template variables, environment exports), (b) `moai update` 10 flags interaction matrix (`check`, `shell-env`, `config -c`, `force`, `yes`, `templates-only`, `binary`, `dry-run`, `no-hooks`, `verbose`), (c) `moai cc -p <profile-name>` profile system (storage location, naming convention, switching mechanism, interaction with `moai cg` + `moai glm`), and (d) cross-cutting concerns — how profile selection in `moai cc -p` interacts with `moai init` template selection and `moai update -c` config-only flag.

**Acceptance**: AC-CLA-003

### REQ-CLA-004 — Sprint 7 CLI-INTEGRATION-001 baseline scope (Ubiquitous)

**The system shall** synthesize REQ-CLA-001/002/003 findings into a "Sprint 7 candidate scope" section at `## §4 Sprint 7 Baseline` of the audit report, recommending (a) subcommands to unify (with rationale), (b) dead commands to retire (with deprecation path proposal), (c) integration gaps requiring code-level bridging — formatted such that Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 manager-spec can directly consume the section as input requirements scope.

**Acceptance**: AC-CLA-004

### REQ-CLA-005 — Research-only constraint (Unwanted)

**The system shall not** modify any `.go` file, hook script (`.claude/hooks/**/*.sh` or `internal/template/templates/.claude/hooks/**/*.sh`), template content, docs-site content, or `.moai/config/sections/*.yaml` file during research execution. The only writable output path is `.moai/reports/cli-audit/` (the research deliverable directory) plus the 3 SPEC artifacts already created at plan-phase (`spec.md`, `plan.md`, `acceptance.md`) and milestone updates to `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (when manager-develop runs).

**Acceptance**: AC-CLA-005

---

## §3 Acceptance Criteria (high-level — full matrix in `acceptance.md`)

| AC | Verification path | REQ link |
|----|-------------------|----------|
| AC-CLA-001 | Audit report `## §1` contains ≥40 subcommand entries with name + parent + file:line | REQ-CLA-001 |
| AC-CLA-002 | Audit report `## §2` classifies every inventoried subcommand as active/dead-suspect/internal-only with grep evidence per class | REQ-CLA-002 |
| AC-CLA-003 | Audit report `## §3` integration map covers init/update/profile triad with dataflow diagram + flag matrix | REQ-CLA-003 |
| AC-CLA-004 | Audit report `## §4` Sprint 7 baseline section directly consumable by future manager-spec invocation | REQ-CLA-004 |
| AC-CLA-005 | `git diff --name-only main -- '*.go' '*.sh' '*.yaml' 'docs-site/' 'internal/template/templates/'` returns empty after run-phase | REQ-CLA-005 |

---

## §A Context

### §A.1 Sprint position

Sprint 2 P2 SPEC. Prior Sprint 2 SPECs (lifecycle-complete): CHANGELOG-CLEANUP-001 / SESSION-HANDOFF-AUTO-001 / LEGACY-CLEANUP-001 / LEGACY-CLEANUP-002 (per memory `project_sprint2_legacy_cleanup_001_run_complete` + earlier entries). After this SPEC, Sprint 2 may proceed to AGENT-FOLDER-NEST-001 (Tier S, flat→nested) or other P3 candidates per v3.0 roadmap Scenario B v2.

### §A.2 Sprint 7 link (deferred consumer)

This SPEC's audit report is the input baseline for Sprint 7 FINAL **SPEC-V3R6-CLI-INTEGRATION-001** (planned but not yet authored). Sprint 7 CLI-INTEGRATION-001 will:
- Implement subcommand unification per `## §4 Sprint 7 Baseline` recommendations
- Retire dead commands with deprecation path per `## §4`
- Bridge integration gaps in init/update/profile triad

The audit report MUST be written such that Sprint 7's manager-spec can consume `## §4` as scope without re-deriving inventory.

### §A.3 Hybrid Trunk execution mode

**Tier M** SPEC. Per CLAUDE.local.md §23 [HARD] Hybrid Trunk policy with Tier M classification, the **default flow** for Tier M is feat-branch + auto PR with 4 CI status checks. However, since this SPEC produces **no code changes** (REQ-CLA-005 [Unwanted]), the CI status checks (Test ubuntu / Lint / Build linux/amd64 / CodeQL) will trivially pass on .md-only diffs.

**Per-SPEC override candidate**: User MAY elect main-direct push for this SPEC at run-phase (analogous to LEGACY-CLEANUP-001 per-SPEC override pattern, lesson L40). Decision deferred to run-phase AskUserQuestion at manager-develop completion. Default (no override): feat-branch + auto PR. If override granted: main-direct push with `🗿 MoAI` trailer + commit body annotation noting per-SPEC override per L40 pattern.

### §A.4 Lifecycle: spec-anchored

The audit report itself is the deliverable. SPEC + audit report co-evolve: if Sprint 7 CLI-INTEGRATION-001 discovers the audit was incomplete, this SPEC's status transitions `implemented → superseded by SPEC-V3R6-CLI-AUDIT-002` (or similar refresh). Until Sprint 7, status flow: `draft → planned → in-progress → implemented`.

### §A.5 Research deliverable path

`.moai/reports/cli-audit/audit-{ISO-DATE}.md` (e.g., `audit-2026-05-23.md` if run on plan-phase day; date reflects actual run-phase execution day).

Report sections:
- `## §1 Subcommand Inventory` (REQ-CLA-001, AC-CLA-001)
- `## §2 Dead-Command Classification` (REQ-CLA-002, AC-CLA-002)
- `## §3 Integration Map` (REQ-CLA-003, AC-CLA-003)
- `## §4 Sprint 7 Baseline Scope` (REQ-CLA-004, AC-CLA-004)
- `## §5 Methodology Appendix` (grep commands used, files scanned, exclusions applied)

---

## §B PRESERVE List (DO NOT MODIFY during run-phase)

[HARD] The following paths are **read-only** during research execution. Any modification violates REQ-CLA-005 [Unwanted]:

- `internal/cli/**/*.go` (all 106 non-test + ~100 test files)
- `cmd/moai/**/*.go`
- `internal/template/templates/**` (entire template tree)
- `.claude/hooks/**/*.sh` (local hooks)
- `.claude/skills/**/*.md` (skill bodies)
- `.claude/rules/**/*.md` (rule bodies)
- `.claude/agents/**/*.md` (agent bodies)
- `docs-site/content/**/*.md` (4-locale user docs)
- `.moai/config/sections/*.yaml` (config files)
- `.moai/specs/**` except this SPEC's own directory (sibling SPEC isolation)
- All ambient runtime files (`.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml`, `.moai/research/v3.0-redesign-2026-05-23.md`)

**Writable** during research:
- `.moai/reports/cli-audit/audit-{ISO-DATE}.md` (NEW deliverable)
- `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/progress.md` (milestone tracker, NEW)
- `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md` (frontmatter status field updates only — `draft → planned → in-progress → implemented`)

---

## §C Exclusions (What NOT to Build)

[HARD] Per SPEC-Builder doctrine, every spec.md MUST include explicit exclusions:

1. **No subcommand retirement** — this SPEC identifies dead-suspect commands but does NOT remove them. Sprint 7 CLI-INTEGRATION-001 retires.
2. **No code refactoring** — no `internal/cli/*.go` consolidation, no helper extraction, no AddCommand registration changes.
3. **No hook script consolidation** — `harness-observe*` family (or other internal-only commands) remain as-is.
4. **No template changes** — `internal/template/templates/.claude/hooks/` is untouched.
5. **No `moai init` / `moai update -c` / profile unification implementation** — only the integration MAP is produced.
6. **No docs-site update** — even if the audit reveals docs-site references stale subcommands, no `.md` edits.
7. **No SPEC frontmatter migration to canonical schema** — `created`/`updated`/`tags` retained per current snake_case alias convention used by sibling Sprint 2 SPECs (canonical `created_at`/`updated_at`/`labels` migration is a separate FUTURE SPEC, not in scope here).
8. **No `/moai mx` annotation step** — research-only SPECs do not add `@MX:NOTE`/`@MX:ANCHOR` annotations to Go code (no Go code change; nothing to annotate).

---

## §D Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Audit incompleteness (missed subcommands) | Medium | M1 cross-reference 3 sources: cobra command tree walk + grep `AddCommand` + grep `cmd.AddCommand` |
| Dead-command false positive (command USED but reference invisible to grep) | High | M2 require ≥2 negative evidence (grep + ast-grep cross-check) before classifying dead-suspect; final dead-list requires user review at run-phase |
| Sprint 7 scope drift (audit too broad to consume) | Medium | REQ-CLA-004 explicitly constrains `## §4 Sprint 7 Baseline` to a directly-consumable scope section (not aspirational wish-list) |
| Audit report becomes stale post-Sprint 2 (more SPECs add subcommands) | Low | Audit report includes `## §5 Methodology Appendix` documenting grep commands used; future audit refresh re-runs same commands |
| Subcommand count estimate (40-55) wrong | Low | Pre-flight discovery already captured 110 AddCommand invocations; true subcommand count may be lower (many AddCommand are sub-subcommand registrations) — M1 produces authoritative count |

---

## §E Related Lessons (apply at run-phase)

- **L9** (parallel session race) — full git fetch before manager-develop spawn
- **L40** (§23 [HARD] per-SPEC explicit override pattern) — applicable for run-phase main-direct push decision
- **L44** (pre-spawn fetch) — verified clean 0 0 at plan-phase entry; reverify at run-phase
- **L26/L27** (Tier S minimal Section A-E + TaskList state isolation) — manager-develop spawn prompt MUST use minimal form for Tier M research-only (no per-file pre-conditions; deliverable IS the audit report)

---

Version: 0.1.0 (plan-phase draft)
Status: draft (transitions to planned after plan-auditor verdict if Tier M ≥0.80 threshold elected, else direct to in-progress at run-phase)
