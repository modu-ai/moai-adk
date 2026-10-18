---
id: SPEC-AGENT-CORE4-001
title: "Research: piece 1 baselines"
version: "0.5.0"
created: 2026-10-18
---

# research.md — SPEC-AGENT-CORE4-001 (piece 1)

Measured on tree `8260bbaa0b` (code identical to `b4ea2d6c93`) unless stated; commands are in
`acceptance.md` §2 and the audit files. Full census and the original scope's evidence: commit
`b4ea2d6c93`, `research.md`.

- Copies: C1 vs C2 differ in 7 of 12 files (builder-harness, e2e-tester, manager-design,
  manager-develop, manager-spec, plan-auditor, super-advisor); C3p has 11 files, C3 has 12
  (manager-todo missing in C3p); C3 vs C3p differ in 8 files plus the missing one.
- Carried four: C1 and C2 differ in exactly 8 lines (builder-harness 38, 155; e2e-tester 92, 143;
  manager-design 50, 203, 206; super-advisor 129). C2 line 143 of e2e-tester holds the shipped
  defect `~/~/.claude/skills/...`. `CLAUDE.md` cited by C2 manager-design:206 does not exist in this
  repository; `.moai/config/sections/delegation.yaml` does.
- Size: the 12 C2 files total 2,969 lines, 252,623 bytes. Highest distinct roster-name count in a
  body: 9 (manager-lead) vs `SweepThreshold = 10` (`internal/harness/rosterguard/check.go:367`).
- Numeral layer: registry rows for the manager-design citation exist for C1, C2, C3
  (`registry.go` around lines 496-512); a copy at a new path produced one undischarged hit in the
  audit's probe, and the same bytes at the registered path produced none.
- Emitter: `emit.go` reads C2 `.md` files, passes bodies through unchanged and renders TOML per
  file; `make agents-emit-check` prints only `ok ...` today. `agents-codex.yaml` classes skill
  loading and sandbox as documented-drop / UNSUPPORTED on Codex (unchanged by piece 1).
- `AGENTS.local.md` §2.0 table: C1 hand-edited local, C2 hand-edited and "the neutral source layer",
  C3 machine-emitted from C2; "C1 <-> C2 is not byte-identical (an intentional divergence)".
- `.gitignore:232-248`: `.moai/reports/*` is local-only; the verdict-file exception was withdrawn.

Gaps: no live Claude Code or Codex run; the exact-byte capability expansion is unproven until M1;
the replacement text of the 8 lines is a plan-phase choice, not a runtime measurement.

Added in iter3: shipped frontmatter keys of all twelve C2 files are `tools`, `color`,
`permissionMode` (some `isolation`, `memory`, `skills`, `hooks`, `model`, `effort`); none has
`capabilities` or `sandbox`. Local main already carries foreign commits on the 16 paths
(`0242d5ef75`, `baecbf74c6`, card t1827, manager-todo twin surfaces). `golden_test.go` Markdown
identity block is lines 106-116 (assertion at 114). `agents-reference.md:46` holds the Selection
Decision Tree, where row 6 is manager-git, not manager-design; `delegation.yaml` has "13 retained
agents" (line 14) and manager-design (line 97) but no Selection Decision Tree. Registry
`CountPattern` for the manager-design citation rows matches `(?:§ 4|delegation\.yaml`) (N retained
agents`, so changing the citation to another file would break three rows.
