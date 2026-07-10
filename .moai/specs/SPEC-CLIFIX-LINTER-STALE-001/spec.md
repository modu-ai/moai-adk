---
id: SPEC-CLIFIX-LINTER-STALE-001
title: "CLI Linter/Doctor Staleness Remediation — agentlint yaml.v3, 10-agent roster, catalog-derived allowlists (P3)"
version: "0.1.0"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P3
phase: "v3.0.0 target"
module: "internal/cli/agentlint"
lifecycle: spec-anchored
tags: "cli, audit-remediation, agentlint, doctor, staleness, p3"
era: V3R6
tier: M
dependencies: [SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001]
---

# SPEC-CLIFIX-LINTER-STALE-001 — CLI Linter/Doctor Staleness Remediation (P3)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §3 clusters 4/5 (staleness rows) + §5 P3 roadmap row |

## §A Context

The quality gates themselves have gone stale against the live system (audit §5 P3: "품질 게이트가 실제 시스템을 다시 감사하게 됨"). agentlint's hand-rolled frontmatter parser never populates Hooks/Skills/Sandbox, permanently disabling rule LR-04; its effort matrix references the retired 17-agent roster so none of the current 10 agents is covered; live+template-mirror double-scan makes LR-07 structurally false-positive; the doctor skills allowlist is hand-maintained and badly stale (live skills WARN, retired skills PASS); root help advertises the nonexistent `moai brain`; and ClaimTask reports success for nonexistent/completed tasks.

Findings SSOT: audit §3 cluster 4 Major rows (help.go, doctor_skills.go) + cluster 5 Major rows (agentlint ×3) + cluster 2 Major row (ClaimTask validation). Re-verify anchors at run time.

## §B Requirements (GEARS)

- REQ-LINT-001-001: The agentlint frontmatter parser shall parse agent frontmatter via yaml.v3 (agentlint/agent_lint.go:530-570), populating the Hooks, Skills, and Sandbox fields, so rule LR-04 (dead-hook detection) can produce findings.
- REQ-LINT-001-002: The agentlint canonical effort matrix and write-heavy agent list (agent_lint.go:581-599) shall be regenerated from the current 10-agent retained catalog, and entries referencing retired agents shall be removed, so every current agent is covered by effort/model policy checks.
- REQ-LINT-001-003: When agentlint scans both live agent files and their template mirrors (agent_lint.go:184,835), rule LR-07 duplicate detection shall deduplicate by path pairing or content hash, so a live/template mirror pair is not reported as a duplicate-agent finding.
- REQ-LINT-001-004: The doctor skills allowlist (doctor_skills.go:10-27) shall be derived from the template catalog instead of a hand-maintained static list, so live skills are not warned and retired skills are not passed.
- REQ-LINT-001-005: The root help text (help.go:57) shall not advertise the nonexistent `moai brain` command, and help command entries shall be gated on actually-registered commands.
- REQ-LINT-001-006: When `ClaimTask` is invoked for a task that does not exist or is not pending (team_spawn.go:345-352), the CLI shall report a claim failure instead of success, and only a pending task match shall produce a successful claim.
- REQ-LINT-001-007: The run-phase implementation shall add regression tests proving each previously-dead check now fires: an LR-04 finding on a dead-hook fixture, an effort-matrix finding on a current-roster violation fixture, zero LR-07 findings on a live/mirror fixture pair, and a doctor WARN on a retired-skill fixture.

## §C Scope

### In Scope

- agentlint parser (yaml.v3), effort matrix/write-heavy roster regeneration, LR-07 dedupe, doctor_skills catalog derivation, help.go phantom command removal, ClaimTask task-state validation, and the regression tests proving the gates fire.

### Out of Scope — Ledger write mechanics

- The ClaimTask O_APPEND ledger fix is SPEC-CLIFIX-CRITICAL-001 (REQ-CRIT-001-002); this SPEC only adds task-existence/pending-state validation on top of the corrected write path.

### Out of Scope — New lint rules

- No new agentlint rule codes are introduced; only existing rules are un-broken. Extending the rule set (e.g., dead-wire detection proposed in audit §4) is future work.

### Out of Scope — agentlint minor parser hygiene

- The cluster-5 minor items (frontmatter `---` anchoring, code-fence column-0 limitation, "specialist" substring over-match, errors package shadowing, first-broken-file abort) are deferred to SPEC-CLIFIX-HYGIENE-001 or later follow-ups; only the three Major agentlint defects are fixed here.

## §D Acceptance Criteria

- AC-LINT-001-001: Given an agent fixture whose frontmatter declares hooks that reference nonexistent scripts, When agentlint runs, Then an LR-04 finding is emitted (maps REQ-LINT-001-001)
- AC-LINT-001-002: Given the current 10-agent catalog, When agentlint's effort matrix is checked, Then every retained agent has a matrix entry and no retired-roster name remains (maps REQ-LINT-001-002)
- AC-LINT-001-003: Given a live agent file and its byte-identical template mirror, When agentlint scans both, Then no LR-07 duplicate finding is emitted for the pair while true duplicates still produce findings (maps REQ-LINT-001-003)
- AC-LINT-001-004: Given the template catalog state, When doctor skills runs, Then every catalog-live skill passes and a retired-skill fixture warns (maps REQ-LINT-001-004)
- AC-LINT-001-005: Given `moai --help` output, When searched for brain, Then no `moai brain` entry appears and every listed command resolves to a registered command (maps REQ-LINT-001-005)
- AC-LINT-001-006: Given a claim request for a nonexistent task ID and for an already-completed task, When ClaimTask runs, Then both return a claim failure and the ledger gains no claim line (maps REQ-LINT-001-006)
- AC-LINT-001-007: Given the regression test suite, When it runs, Then the four previously-dead checks (LR-04, effort matrix, LR-07 dedupe, doctor allowlist) each demonstrate a firing and a non-firing case (maps REQ-LINT-001-007)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001 (team_spawn.go base), SPEC-CLIFIX-CONTRACT-001 (agentlint/workflow_lint.go os.Exit removal precedes parser work in the same package), SPEC-CLIFIX-CONCURRENCY-001 (series order P0→P1→P2→P3).
- Non-goal: changing agent authoring conventions or the agent catalog itself — the linter is realigned to the existing 10-agent SSOT, not vice versa.
