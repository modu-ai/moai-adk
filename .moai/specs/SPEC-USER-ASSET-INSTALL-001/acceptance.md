---
id: SPEC-USER-ASSET-INSTALL-001
title: "acceptance.md — acceptance criteria matrix"
version: "0.1.0"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Acceptance — SPEC-USER-ASSET-INSTALL-001

Verification layer: every criterion is Given-When-Then, binary-testable.
Requirement layer (GEARS obligations) lives in spec.md §2; the REQ↔AC map is
spec.md §3. All user-folder tests run against isolated temp HOMEs (plan §D).

## D. AC Matrix

| AC | Verifies | Binary test |
|---|---|---|
| AC-001 | REQ-001 | L0 skill dirs exist under temp `$HOME/.claude/skills/` with bytes matching catalog hashes |
| AC-002 | REQ-001, REQ-022 | `$HOME/.agents/skills/<skill>/SKILL.md` + `$HOME/.codex/agents/<name>.toml` exist for the L0 set |
| AC-003 | REQ-006 | Per-user manifest written; every installed path carries sha256 + bundle + moai version |
| AC-004 | REQ-012 | Second install run: zero file writes (mtime/hash proof), zero-delta report |
| AC-005 | REQ-008 | Shipped-byte change → update rewrites the tracked file; hash in manifest refreshed |
| AC-006 | REQ-009, REQ-021 | Manifest-tracked file dropped from all bundles → removed by update; unknown schema_version → removal refused |
| AC-007 | REQ-010 | Untracked file at an install target → left byte-identical, collision reported |
| AC-008 | REQ-010 | Untracked file at a refresh/removal target → untouched, collision reported |
| AC-009 | REQ-014 | Doctor user-install check: detects manifest-tracked file deleted / hash-modified / untracked entry in the four roots |
| AC-010 | REQ-015 | Doctor project-vs-lock check: detects project file absent from lock AND lock entry absent from project |
| AC-011 | REQ-005 | Post-init project tree contains NO `moai-*` skill or agent asset |
| AC-012 | REQ-001, REQ-005 | Post-init project tree contains settings, AGENTS.md, lock file, hooks, `.mcp.json` (with moai MCP entry) |
| AC-013 | REQ-016 | Repo tree carries no `plugins/moai/`, no marketplace manifests; `make build` green without plugin-emit targets |
| AC-014 | REQ-017 | Boundary grep: init/update paths hold zero plugin marketplace/install invocations |
| AC-015 | REQ-018 | Boundary grep + build: zero `DeployModePlugin`/`PluginMirrorPolicy` references; single deploy payload shape |
| AC-016 | REQ-019 | Doctor output carries no "Plugin Deployment"/"Plugin Version" carrier rows; replacement rows present per D-Q6 |
| AC-017 | REQ-003 | Installed L0 set equals the resolved gate answer (5 agents by name, plan/run/sync surface, hooks, factory) |
| AC-018 | REQ-004 | Bundle install adds exactly the bundle's catalog entries; bundle removal takes exactly them |
| AC-019 | REQ-002 | Profile dirs (`~/.moai/claude-profiles/<name>`) byte-unchanged by install/update (settings isolation) |
| AC-020 | REQ-020 | Migration: template-managed project skills removed; user-modified preserved + reported; user-created untouched |
| AC-021 | REQ-021 | Manifest with foreign schema_version → install/refresh proceed, removal refuses with named error |
| AC-022 | REQ-011 | Mixed update run summary: installed/refreshed/removed/collision counts each match a seeded fixture |
| AC-023 | REQ-013 | Read-only target dir → remaining files still processed; summary lists the failure with path + reason |
| AC-024 | REQ-007 | Install executes with network disabled (no dial in trace); same output as online run |

## D.1 Severity

- Blocker: AC-001, AC-002, AC-003, AC-007, AC-011, AC-013, AC-020
- Major: AC-004, AC-005, AC-006, AC-008, AC-009, AC-010, AC-012, AC-014,
  AC-015, AC-016, AC-017, AC-018, AC-021, AC-022
- Minor: AC-019, AC-023, AC-024

## D.2 Traceability

Every REQ-001..REQ-022 maps to ≥1 AC (spec.md §3); every AC maps to exactly
one primary REQ (table above). No orphan AC, no uncovered REQ.

## D.3 Indirect Verification

- AC-013/AC-015 verify ABSENCE: accepted evidence is a boundary grep with a
  POSITIVE CONTROL (the grep pattern matched at baseline, pre-retirement) —
  a zero-hit grep without a positive control proves nothing.
- AC-024's "no dial in trace" uses the binary's own offline operation; a
  network-less CI lane run is acceptable evidence.

## D.4 Edge Cases (covered inside the ACs above)

- Colliding DIRECTORY (not file) at a skill path → AC-007 semantics apply.
- Manifest present but corrupt JSON → refuse removal, report corruption,
  offer rebuild-from-scan (doctor informational), never auto-delete.
- Partial prior install (manifest newer/older than tree) → AC-009 reports the
  divergence class.
- Simultaneous project lock drift and user-install drift → doctor reports
  both independently (AC-009 + AC-010 compose).

## D.5 Quality Gate Criteria

- TRUST 5: 85%+ coverage on new packages; `golangci-lint` clean; gofmt clean;
  conventional commits referencing card t1509.
- Full `make build` (with the post-M6 prerequisite list) green at M6/M7 exit.
- No test writes outside temp dirs (verified by the suite's existing
  HOME-seam tests staying green).

## D.6 Definition of Done

- All Blocker + Major ACs PASS with observed evidence; Minor ACs PASS or carry
  an explicit accepted-debt note in progress.md §E.2.
- decision-index D-Q1..D-Q6 all carry operator verdicts (none open at close).
- REQ-016..019 boundary greps recorded with positive controls in §E.2.
- CHANGELOG entry; retiring SPEC notes SPEC-PLUGIN-MARKETPLACE-001's carrier
  as retired (relationship recorded, history preserved).

## D.7 Forward-Looking Checks

- The manifest schema carries room for future per-file metadata (no breaking
  change when a field is added — consumers ignore unknown fields).
- The bundle taxonomy tolerates a future pack→bundle rename without a
  manifest migration (bundle is a string label, not an enum).
- D-Q3's declared limitation (if chosen) is doc-visible so a later SPEC can
  lift it without re-deriving the analysis.
