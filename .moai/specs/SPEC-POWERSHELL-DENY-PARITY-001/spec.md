---
id: SPEC-POWERSHELL-DENY-PARITY-001
title: "PowerShell deny-rule parity for Bash denies outside the built-in removal protection"
version: "0.2.1"
status: in-progress
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
| 0.2.1 | 2026-09-26 | manager-spec | Plan-audit iter-2 conditional PASS 0.89 follow-up (N1-N5): one turn-cap rule (hitting `--max-turns` or the 180 s timeout is INCONCLUSIVE), fresh scratch project per arm, `TRUNCATE` sub-choice under D1, glued trailing `*` in the matcher model, built-in alias coverage marked as inference. |
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iter-1 FAIL 0.66 revision (D1-D15). Premise corrected with Claude Code's built-in PowerShell removal protection; scope reframed to the residual Bash denies it does not cover; measurement re-aimed at a residual command with tool restriction, source isolation, and a full validity/outcome table; REQs and ACs split by branch; parity guard made closed-world. |

## §A Background

### A.1 The PowerShell tool and its rule namespace

Claude Code can run shell commands through a dedicated **PowerShell tool** instead of the Bash tool. Per `https://code.claude.com/docs/en/tools-reference` § PowerShell tool (fetched 2026-09-26): on Windows without Git Bash "the tool is enabled automatically"; on Windows with Git Bash "the tool is on by default for claude.ai and Console accounts"; on Linux, macOS, and WSL "the tool is opt-in" and "requires PowerShell 7 or later: install `pwsh` and ensure it is on your `PATH`" (toggle `CLAUDE_CODE_USE_POWERSHELL_TOOL`).

Per `https://code.claude.com/docs/en/permissions` § PowerShell (fetched 2026-09-26): "PowerShell permission rules use the same shape as Bash rules. Wildcards with `*` match at any position, the `:*` suffix is equivalent to a trailing ` *`"; "Common aliases are canonicalized before matching. A rule written for the cmdlet name also matches its aliases ... Matching is case-insensitive"; each subcommand of a compound command is checked independently.

### A.2 Built-in removal protection already covers the filesystem-root denies

Per `https://code.claude.com/docs/en/permission-modes` § Critical paths › Remove-Item in PowerShell (fetched 2026-09-26), Claude Code gives `Remove-Item` and the `cmd` built-ins `rd`, `rmdir`, `del`, and `erase` their own checks:

- "**System paths**: the filesystem root and its top-level directories, drive roots and their top-level directories, and your home directory. Claude Code denies the command in every mode, without asking you."
- "**Wildcards**: a bare `*`, or any target ending in `/*` or `\*` ... Claude Code denies the command in every mode, without asking you."
- "The system-paths case also applies to `rd`, `rmdir`, `del`, and `erase` when Claude runs them through `cmd` ... This `cmd` check requires Claude Code v2.1.283 or later." It can be turned off with `CLAUDE_CODE_DISABLE_POWERSHELL_CMD_RM_DENY=1`; "`Remove-Item` on a system path stays denied either way."
- The same page, § bypassPermissions: "The Remove-Item in PowerShell denies also apply in this mode."

Consequence: the 9 filesystem-root `Bash(...)` denies of the template (`rm -rf /`, `rm -rf /*`, `rm -rf ~`, `rm -rf ~/*`, `rm -rf C:/`, `rm -rf C:/*`, `del /S /Q C:/`, `rmdir /S /Q C:/`, `Remove-Item -Recurse -Force C:/`) are already enforced on the PowerShell path for `Remove-Item` (and — inferred, not documented for the built-in check: its canonicalized aliases; recorded as a Gap alongside native `rm`) and for `cmd`-run `rd`/`rmdir`/`del`/`erase`. A PowerShell mirror of those 9 rows adds no enforcement for the documented cases. One case is not documented either way: a native `rm` binary run from `pwsh` on macOS/Linux (where `rm` need not be a `Remove-Item` alias). That case is recorded as a Gap, not measured (measuring it would require a destructive root removal).

### A.3 The residual gap

The distributed template `internal/template/templates/.claude/settings.json.tmpl`, generated from the SSOT `.moai/config/sections/tool-policy.yaml` (SPEC-V3R6-TOOL-POLICY-SSOT-001), carries **47** `Bash(...)` deny rules on local `develop` after t1207 and **0** `PowerShell(...)` rules (measured at plan time by extracting the template deny array). Of the 47, **38** fall outside the built-in protection above:

| Category | Rows | Examples |
|---|---|---|
| git history-destructive | 14 | `git push --force:*`, `git reset --hard:*`, `git clean -fdx:*`, `git * push --force*` |
| disk formatting | 3 | `Clear-Disk:*`, `Format-Volume:*`, `format:*` |
| system / process | 11 | `chmod -R 777:*`, `dd:*`, `mkfs:*`, `shutdown:*`, `systemctl:*`, `kill -9:*`, `killall:*` |
| database | 10 | `DROP DATABASE:*`, `TRUNCATE:*`, `redis-cli FLUSHALL:*`, `psql -c DROP:*` |

**Hypothesis (UNVERIFIED — measured in run-phase M1):** `Bash(...)` deny rules are not consulted for commands issued through the PowerShell tool, so these 38 denies are absent on the PowerShell path. The documentation presents `PowerShell(...)` as a separate namespace but does not literally state that Bash rules are skipped for PowerShell commands.

### A.4 Dependency — t1207

Card t1207 (merged into local `develop` at `baa054586`; this branch's base `4dcd4d8d4` predates it) replaced the defective `C\:/` escape in the Windows Bash denies with a literal `C:/`. Run-phase starts only after this branch absorbs a local `develop` that carries t1207, and no rule authored here may reintroduce the `\:` escape.

## §B Requirements (GEARS)

Two delivery branches exist. **Branch P** (parity) applies when M1 confirms the gap; **Branch N** (no-rule) applies when M1 shows Bash denies already reach the PowerShell tool. Each requirement below names the branch it binds; unmarked requirements bind both.

### B.1 Measurement gate

- **REQ-PSD-001** — The run-phase shall measure, before any deny rule is added, whether a `Bash(...)` deny rule from the residual set (§A.3) blocks the same command issued through the PowerShell tool, and shall record the observable outcome of each measurement arm in `progress.md` §E.2.
- **REQ-PSD-002** — The measurement shall run in a scratch project outside the repository, with only the scratch project's setting source loaded, with hooks and other customizations disabled, with the PowerShell tool as the only available built-in tool, under caps declared before execution (at most one run per arm, at most 3 model turns, at most 180 s wall clock per arm, no retries), and with every spawned process bounded by an external timeout.
- **REQ-PSD-003** — Where Claude Code offers a mechanical (model-free) way to evaluate a permission rule against a command, the measurement shall use it in preference to a model-driven session.
- **REQ-PSD-004** — When PowerShell 7 or later is not on `PATH` at measurement time, the run-phase shall stop and return a blocker report without running any arm.
- **REQ-PSD-005** — When any validity condition of the measurement does not hold (the deny-path control arm does not show a denial with the target surviving, the no-rule control arm does not show the command executing through the PowerShell tool, the session's tool list is not the PowerShell tool alone, the recorded tool call is not a PowerShell call, or an arm hits the declared `--max-turns` cap or the 180 s wall-clock timeout), the run-phase shall stop, return a blocker report, and select neither branch.
- **REQ-PSD-006** — When the measurement is valid and the residual `Bash(...)` deny does not block the PowerShell command, the run-phase shall take Branch P.
- **REQ-PSD-007** — When the measurement is valid and the residual `Bash(...)` deny blocks the PowerShell command with a recorded denial, the run-phase shall take Branch N, shall not add any `PowerShell(...)` rule or rule-parity test, and shall deliver the recorded §E.2 evidence together with the documentation note of REQ-PSD-016.
- **REQ-PSD-008** — The run-phase shall record, as an observation that selects no branch, whether the built-in wildcard removal deny of §A.2 fires for a `Remove-Item` whose target is a wildcard inside the scratch project.

### B.2 Rule-set parity (Branch P)

- **REQ-PSD-009** — Where Branch P is taken, the tool policy shall declare a `PowerShell(...)` deny counterpart with the same deny intent for every `Bash(...)` deny inside the scope chosen at decision D1, and every other `Bash(...)` deny shall appear on the declared exclusion list with a stated reason.
- **REQ-PSD-010** — Where Branch P is taken, a `Bash(...)` deny whose command head is a PowerShell alias of a differently named cmdlet on any supported platform shall be placed on the exclusion list with the reason "alias head, unmeasured", unless run-phase evidence shows that the alias-form rule matches.
- **REQ-PSD-011** — Where Branch P is taken, the `PowerShell(...)` deny rules shall not contain the `\:` escape and shall not combine a non-trailing `*` with the `:*` suffix.
- **REQ-PSD-012** — Where Branch P is taken, the `PowerShell(...)` deny rules shall not match any command of the declared benign sample (plan.md §C.3), evaluated case-insensitively as the documentation specifies.

### B.3 Delivery path

- **REQ-PSD-013** — Where Branch P is taken, the PowerShell rules shall be authored in `.moai/config/sections/tool-policy.yaml` first and propagated to `internal/template/templates/.claude/settings.json.tmpl` and the local `.claude/settings.json` by the tool-policy generator, and `make tool-policy-drift-check` and `make build` shall pass afterwards.
- **REQ-PSD-014** — Where Branch P is taken, the repository shall carry a static Go test over the rendered template that fails when any `Bash(...)` deny rule has neither a declared `PowerShell(...)` counterpart present in the deny list nor an entry on the exclusion list, fails when any `PowerShell(...)` deny rule maps to no `Bash(...)` deny, and fails when any `PowerShell(...)` deny rule violates REQ-PSD-011.
- **REQ-PSD-015** — The template change shall remain neutral across the 16 supported programming languages and free of internal content (no SPEC IDs, card IDs, dates, or commit SHAs in `internal/template/templates/**`), as enforced by the existing template-neutrality and internal-content-leak tests.
- **REQ-PSD-016** — The docs-site page `advanced/settings-json.md` shall carry, in all four locales (ko, en, ja, zh) in the same change, the branch's note: under Branch P, which deny rules are declared for both the Bash and the PowerShell tools and that filesystem-root removals rely on Claude Code's built-in protection; under Branch N, that the Bash deny rules were measured to apply to PowerShell tool commands on the recorded Claude Code version.

### B.4 Unwanted behaviour

- **REQ-PSD-017** — The run-phase shall not modify the semantics of the existing `Bash(...)` deny rules (t1207 owns their form).
- **REQ-PSD-018** — The run-phase shall not add `PowerShell(...)` rules to the `allow` or `ask` lists and shall not change `defaultMode`.

## §C Constraints

- Template-First: SSOT → template → local, then `make build` (project-local development guide §2).
- The tool-policy loader and codegen accept any non-empty tool name (verified by plan-audit iter-1: `types.go` `Tool string`, `loader.go` non-empty check only, codegen emits `SettingsSpecifier()` for any tool); no generator change is expected.
- No test may spawn Claude Code or pwsh inside `go test`; every Go guard of this SPEC is a static assertion over the rendered template JSON.
- The measurement result is taken on macOS with the opt-in PowerShell tool; carrying it over to Windows is an inference recorded as residual risk.

## §D Out of Scope

### Out of Scope — Hook matcher coverage of the PowerShell tool

- The PreToolUse hook matcher `Write|Edit|Bash` in the template does not name `PowerShell`, so PowerShell commands bypass `handle-pre-tool.sh` guards. The vendor documentation states the hazard directly (`https://code.claude.com/docs/en/tools-reference` § PowerShell tool: "Match `Bash|PowerShell` in hooks that inspect shell commands; ... matching `Bash` alone is not enough"; "The Bash tool remains available for POSIX scripts when Git Bash is installed"). This is a separate axis (which guards can parse PowerShell syntax) and is proposed as a separate card (open decision D3).

### Out of Scope — Redesign of the Bash deny rules

- The `Bash(...)` rule forms, including the Git Bash `/c/` path gap recorded in `.moai/reports/t1207/`, belong to t1207 and its follow-ups.

### Out of Scope — Re-implementing the built-in removal protection

- PowerShell mirrors of the 9 filesystem-root denies are not in the default scope, because the built-in protection of §A.2 already denies them; including them is the operator's explicit choice at D1.

### Out of Scope — New deny intents

- Destructive patterns without a current Bash counterpart (for example `Stop-Computer`, `Restart-Computer`, or `Remove-Item` on arbitrary non-root paths) are not added; this SPEC mirrors, it does not widen.

### Out of Scope — `allow` / `ask` parity

- `Bash(...)` allow and ask rules are not mirrored to PowerShell.
