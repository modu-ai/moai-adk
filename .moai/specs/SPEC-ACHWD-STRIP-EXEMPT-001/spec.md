---
id: SPEC-ACHWD-STRIP-EXEMPT-001
title: "Apply the AC-HWD-015 strip-aware mirror amendment to SPEC-HOOK-WIRING-DRIFT-001 (card t469 wrapper record)"
version: "0.1.0"
status: completed
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".moai/specs/SPEC-HOOK-WIRING-DRIFT-001"
lifecycle: spec-first
tags: "spec-amendment, card-t469, template-neutrality, strip-aware-mirror"
tier: S
depends_on: [SPEC-HOOK-WIRING-DRIFT-001]
---

# SPEC-ACHWD-STRIP-EXEMPT-001 — card t469 wrapper record

## §A What this SPEC is

A thin wrapper record for card t469. The card's entire deliverable is an
in-place amendment of the completed SPEC `SPEC-HOOK-WIRING-DRIFT-001`: its
acceptance criterion AC-HWD-015 demanded a byte-identical local↔template
mirror, which is unsatisfiable together with that SPEC's own REQ-HWD-014
(template-side neutrality stripping) for any file whose local copy carries a
forbidden-class token.

**Plan authority is NOT duplicated here.** The amendment design, the
re-measured evidence, the machine-check command, the mutant record, and the
exact payload texts A1–A5 all live in
`.moai/specs/SPEC-HOOK-WIRING-DRIFT-001/plan.md` §I (plan-audit
PASS-WITH-DEBT 0.80, `.moai/reports/t469/plan-audit.md`; iteration-2 patch
D1–D5 applied at `8eb5b9102`). This record exists so the card carries its own
SPEC ID, evidence path, lifecycle, and carried-debt ledger.

## §B Requirements (GEARS)

- **REQ-ASE-001** (Ubiquitous) — The amendment payloads A1–A5 defined in
  `SPEC-HOOK-WIRING-DRIFT-001` plan.md §I.4 shall be applied verbatim to that
  SPEC's `spec.md` and `acceptance.md`.

- **REQ-ASE-002** (Event) — **When** the strip-aware mirror check
  (SPEC-HOOK-WIRING-DRIFT-001 plan.md §I.4 A4 command, run from the project
  root) executes, this wrapper's verification record shall capture exit
  status 0.

- **REQ-ASE-003** (Event) — **When**
  `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md`
  executes, this wrapper's verification record shall capture 0 errors.

## §3 Acceptance criteria (Tier S — inline)

- **AC-ASE-001** — **Given** the run-phase commit on `WT-achwd-strip-exempt`,
  **When** SPEC-HOOK-WIRING-DRIFT-001's `spec.md` and `acceptance.md` are
  diffed against the pre-run state (`a122e7568`), **Then** the changes are
  exactly payloads A1–A5 — frontmatter `version: "0.4.0"`,
  `status: in-progress`, `amendment_of:` self-reference; the v0.4.0 HISTORY
  row; the `## Amendments` sub-section; the REQ-HWD-013 rewording; the §F
  `### Out of Scope` H3; and the AC-HWD-015 rewrite — and nothing else.
- **AC-ASE-002** — **Given** the amended tree, **When** the §I.4 A4 perl
  command runs from the project root, **Then** it exits 0 with no `MISMATCH`
  output on the three M3 files.
- **AC-ASE-003** — **Given** the amended tree, **When** the spec lint command
  runs, **Then** it reports 0 errors (the 18 pre-existing
  `CoverageIncomplete` warnings are expected to persist at the same count).

## §F Exclusions

### Out of Scope — everything beyond applying the planned amendment

- Any edit under `internal/template/templates/**` — the amendment edits SPEC
  artifacts only.
- Re-opening, re-verifying, or re-scoring any other AC of
  SPEC-HOOK-WIRING-DRIFT-001; t216's landing stands.
- The fleet-wide strip-aware mirror invariant across all template-managed
  pairs (owned by the neutrality doctrine or a future dedicated SPEC — see
  SPEC-HOOK-WIRING-DRIFT-001 plan.md §I.4 A3).

## §H Cross-references

- Plan authority: `.moai/specs/SPEC-HOOK-WIRING-DRIFT-001/plan.md` §I
- Amendment target: `.moai/specs/SPEC-HOOK-WIRING-DRIFT-001/` (spec.md,
  acceptance.md)
- Evidence: `.moai/reports/t469/` (plan-summary.md, plan-audit.md)
- Card: t469, branch `WT-achwd-strip-exempt`
