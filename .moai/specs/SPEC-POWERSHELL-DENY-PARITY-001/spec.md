---
id: SPEC-POWERSHELL-DENY-PARITY-001
title: "PowerShell deny-rule parity with the Bash destructive-command denies"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/config/sections/tool-policy.yaml, internal/config/toolpolicy, internal/template/templates/.claude/settings.json.tmpl, docs-site"
lifecycle: spec-anchored
tags: "permissions, powershell, windows, deny, tool-policy, template, security, t1211"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-TOOL-POLICY-SSOT-001]
---

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec | Initial draft for card t1211. Measurement-first design: run-phase M1 decides whether the gap exists before any rule is added. |

## §A Background

Claude Code on Windows can execute shell commands through a dedicated **PowerShell tool** instead of the Bash tool. Per the official permissions documentation (fetched 2026-09-26, `https://code.claude.com/docs/en/permissions` § PowerShell), PowerShell permission rules live in their own `PowerShell(...)` namespace, use the same shape as Bash rules (`*` wildcard at any position, `:*` suffix equivalent to a trailing ` *`), canonicalize common aliases before matching (`PowerShell(Get-ChildItem *)` matches `gci`, `ls`, `dir`), match case-insensitively, and check each sub-command of a compound command (`|`, `;`, `&&`, `||`) independently via the PowerShell AST. Per `https://code.claude.com/docs/en/tools-reference`, the PowerShell tool is enabled automatically on Windows without Git Bash, on by default for claude.ai/Console accounts on Windows with Git Bash (toggle `CLAUDE_CODE_USE_POWERSHELL_TOOL`), and opt-in on Linux/macOS/WSL with pwsh 7+.

The distributed template `.claude/settings.json.tmpl`, generated from the SSOT `.moai/config/sections/tool-policy.yaml` (SPEC-V3R6-TOOL-POLICY-SSOT-001), carries **47** `Bash(...)` deny rules (counted on local `develop` after t1207) and **0** `PowerShell(...)` rules.

**Hypothesis (UNVERIFIED — to be measured in run-phase M1):** `Bash(...)` deny rules do not apply to commands issued through the PowerShell tool, so every destructive-command deny is absent on the PowerShell path. The documentation presents `PowerShell(...)` as a separate namespace but does not literally state that Bash rules are not consulted for PowerShell. This SPEC therefore makes the fix conditional on a recorded measurement.

**Dependency — t1207.** Card t1207 (merged into local `develop` at `baa054586`, not yet present on this branch's base `4dcd4d8d4`) replaced the defective `C\:/` escape in the Windows Bash deny rules with a literal `C:/` (the escaped form matched a literal backslash and never blocked a real `C:/` command). PowerShell rules authored here must use t1207's landed form and must not reintroduce the `\:` escape.

## §B Requirements (GEARS)

### B.1 Measurement gate

- **REQ-PSD-001** — The run-phase shall measure, before any deny rule is added, whether a `Bash(...)` deny rule blocks the equivalent command issued through the PowerShell tool, and shall record the observable outcome of each measurement arm in `progress.md` §E.2.
- **REQ-PSD-002** — The measurement shall run in a scratch project outside the repository, with no repository settings or hooks loaded, under declared caps (one run per arm, at most 3 model turns, at most 180 s wall clock per arm), and with every spawned process bounded by an external timeout.
- **REQ-PSD-003** — Where Claude Code offers a mechanical (model-free) way to evaluate a permission rule against a command, the measurement shall use it in preference to a model-driven session.
- **REQ-PSD-004** — When the measurement shows that the `Bash(...)` deny rule already blocks the PowerShell command, the run-phase shall not add `PowerShell(...)` rules, and shall instead land only the regression guard (REQ-PSD-011) asserting the measured behaviour plus a documentation note.
- **REQ-PSD-005** — When the measurement is inconclusive (cap reached, no deletion and no deny event observable), the run-phase shall stop and return a blocker report rather than proceed on the hypothesis.

### B.2 Rule-set parity

- **REQ-PSD-006** — When the measurement confirms that `Bash(...)` denies do not reach the PowerShell tool, the tool policy shall declare a `PowerShell(...)` deny counterpart for every `Bash(...)` deny rule in scope per the parity table (plan.md §C), with the same deny intent.
- **REQ-PSD-007** — The PowerShell deny set shall cover the Windows-destructive intents of `rm -rf`, `del /S /Q`, `rmdir /S /Q`, and `Remove-Item -Recurse -Force` on the drive root, on the Unix root, and on the home directory, such that the aliases `rm`, `del`, `rd`, `ri`, `rmdir`, and `erase` of `Remove-Item` are covered through Claude Code's alias canonicalization.
- **REQ-PSD-008** — The PowerShell deny rules shall not contain the `\:` escape; drive paths shall be written in the literal form adopted by t1207.
- **REQ-PSD-009** — The PowerShell deny rules shall not block commands outside their stated deny intent: a declared benign sample (plan.md §C.3) shall remain allowed.

### B.3 Delivery path

- **REQ-PSD-010** — The PowerShell rules shall be authored in `.moai/config/sections/tool-policy.yaml` first and propagated to the template `.claude/settings.json.tmpl` and to the local `.claude/settings.json` by the tool-policy generator, and `make build` together with `make tool-policy-drift-check` shall pass afterwards.
- **REQ-PSD-011** — The repository shall carry a Go test that fails when any in-scope destructive `Bash(...)` deny rule of the template lacks its declared `PowerShell(...)` counterpart, and that fails when any `PowerShell(...)` deny rule contains the `\:` escape.
- **REQ-PSD-012** — The template change shall remain neutral across the 16 supported programming languages and free of internal content (no SPEC IDs, card IDs, dates, or commit SHAs in `internal/template/templates/**`), as enforced by the existing template-neutrality and internal-content-leak tests.
- **REQ-PSD-013** — Where a docs-site page enumerates the template's destructive deny rules, that page shall state, in all four locales (ko, en, ja, zh) in the same change, that the denies are declared for both the Bash and the PowerShell tools.

### B.4 Unwanted behaviour

- **REQ-PSD-014** — The run-phase shall not modify the semantics of the existing `Bash(...)` deny rules (t1207 owns their form).
- **REQ-PSD-015** — The PowerShell rules shall not be added to the `allow` or `ask` lists, and shall not change `defaultMode`.

## §C Constraints

- Template-First: SSOT → template → local, then `make build` (CLAUDE.local.md §2).
- The tool-policy loader and codegen must accept `tool: "PowerShell"` entries; if they hard-code `Bash`, that is in scope as the minimum generator change required by REQ-PSD-010.
- No test may spawn Claude Code or pwsh inside `go test`; the Go guard (REQ-PSD-011) is a static assertion over the template JSON.
- Run-phase starts only after this branch absorbs local `develop` carrying t1207 (`baa054586` or later).

## §D Out of Scope

### Out of Scope — Hook matcher coverage of the PowerShell tool

- The PreToolUse hook matcher `Write|Edit|Bash` (template line 58) does not name `PowerShell`, so PowerShell commands may bypass `handle-pre-tool.sh` guards (branch guard, integration lock, etc.). This is a sibling axis with its own design questions (which guards parse PowerShell syntax); it is recorded as open decision D3 and proposed as a separate card, not changed here.

### Out of Scope — Redesign of the Bash deny rules

- The `Bash(...)` rule forms, including the Git Bash `/c/` path gap recorded in `.moai/reports/t1207/`, belong to t1207 and its follow-ups.

### Out of Scope — New deny intents

- Adding destructive patterns that have no current Bash counterpart (e.g. `Remove-Item` on arbitrary non-root paths) is out of scope; this SPEC mirrors, it does not widen.

### Out of Scope — `allow` / `ask` parity

- `Bash(...)` allow and ask rules are not mirrored to PowerShell.
