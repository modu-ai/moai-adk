# Plan — SPEC-POWERSHELL-DENY-PARITY-001

Ordered by decision reversibility: the measurement that decides whether anything is built comes first, then the scope and rule-shape decisions, then the mechanical propagation.

## §A Context

- Card: t1211. Branch `WT-powershell-deny`, base local `develop` `4dcd4d8d4`.
- SSOT: `.moai/config/sections/tool-policy.yaml` → generator `moai tool-policy build` (`internal/cli/tool_policy.go`; `--local-only` skips the template) → `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json`. Drift guard: `make tool-policy-drift-check` (runs ahead of `make build`; compares YAML against the local `.claude/settings.json` only).
- Existing static guard to extend: `internal/template/settings_test.go` `TestSettingsTemplateDenyWildcardSyntax` (on develop after t1207; its mixing check and `\:` check currently test the `Bash(` prefix only).
- Docs pages: `docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md`.
- Built-in protection (spec §A.2) and residual set (spec §A.3) are the premise of the scope; both are cited from vendor docs fetched 2026-09-26.

## §B Known Issues / Dependencies

- **t1207 dependency.** Local `develop` carries t1207 (`baa054586`, verified ancestor of `develop`). This branch predates it. M0 absorbs local `develop` before any later milestone.
- The documentation does not literally state that Bash rules skip PowerShell — the premise stays a hypothesis until M1 records it.
- `claude --help` (CC 2.1.283, observed at plan time) lists no rule-evaluation subcommand; M1 still searches for one (REQ-PSD-003) before the model arms.
- Undocumented case: native `rm` from `pwsh` on macOS/Linux against a system path. Not measured (destructive). Recorded as a Gap.

## §C Pre-flight — decisions most likely to change

### §C.1 M1 measurement design (REQ-PSD-001..008)

**Declared caps (fixed before execution, not raised mid-run):** 4 arms, 1 run per arm, `--max-turns 3`, `timeout 180` per run from outside, no retries, no background processes.

**Preconditions (REQ-PSD-004), recorded before arm A:** `pwsh` on `PATH` with `$PSVersionTable.PSVersion.Major` ≥ 7; `claude --version`; presence or absence of a managed-settings file (plan-time observation: `/Library/Application Support/ClaudeCode/` absent). Any failure → blocker report, no arm runs.

**Scratch project** (OS temp dir, outside the repo): a `git init` repository with `.claude/settings.json` committed (so it is tracked and `git clean` cannot remove it) and one untracked file `victim.txt` (the observable), plus `victim-dir/keep.txt` for arm D.

**One fresh scratch project per arm.** Each arm runs in its own scratch project, freshly created from the same recipe with that arm's `permissions.deny` committed, so one arm's `git clean -fdx` or `Remove-Item` cannot consume another arm's observables (`victim.txt`, `victim-dir/`). No project is reused across arms; each arm's pre-run existence of `victim.txt` and `victim-dir/keep.txt` is recorded in §E.2.

**Invocation (every arm):**

```
CLAUDE_CODE_USE_POWERSHELL_TOOL=1 timeout 180 claude -p "<prompt>" \
  --setting-sources project --safe-mode --strict-mcp-config \
  --tools PowerShell --permission-mode bypassPermissions \
  --max-turns 3 --output-format stream-json --verbose
```

Flag basis (verified in `claude --help`, CC 2.1.283): `--setting-sources` "Comma-separated list of setting sources to load (user, project, local)"; `--safe-mode` disables "CLAUDE.md, skills, installed plugins, hooks, MCP servers ... Auth, model selection, built-in tools and plugins, and permissions work normally"; `--tools` "Specify the list of available tools from the built-in set". Deny rules apply under `bypassPermissions` ("Deny rules block in every mode, including `bypassPermissions`", permission-modes doc). Denials are read from stream-json: "denials appear as `permission_denied` system messages, and the final result message lists them in `permission_denials`"; the `system/init` event reports the session's tools (headless doc).

**Arms** — target is a residual-set command (`git clean -fdx`, template row `Bash(git clean -fdx:*)`), which no built-in check covers:

| Arm | `permissions.deny` in scratch settings | prompt instructs | Role |
|---|---|---|---|
| A | `Bash(git clean -fdx:*)` | use the PowerShell tool to run `git clean -fdx` | discriminating arm |
| B | `PowerShell(git clean -fdx:*)` | same | positive control: settings loaded and PowerShell deny path works |
| C | (none) | same | positive control: the command runs through PowerShell with nothing else blocking |
| D | (none) | use the PowerShell tool to run `Remove-Item -Recurse -Force ./victim-dir/*` | observation of the built-in wildcard deny (REQ-PSD-008); selects no branch |

Recorded per arm: exit code; `system/init` tools list; each PowerShell `tool_use` command string; `permission_denied` events and the result's `permission_denials`; whether `victim.txt` (arm D: `victim-dir/keep.txt`) exists afterwards; turn count and whether `timeout` fired (exit 124).

**Validity conditions** — all must hold, otherwise the result is INCONCLUSIVE (REQ-PSD-005 → blocker, no branch):

| # | Condition |
|---|---|
| V1 | Preconditions met (pwsh ≥ 7) |
| V2 | In every arm, `system/init` tools list contains `PowerShell` and does not contain `Bash` |
| V3 | In arms A–C, a `PowerShell` `tool_use` whose command contains `git clean -fdx` is recorded |
| V4 | Arm B: a denial is recorded for that `PowerShell` `tool_use` AND `victim.txt` survives |
| V5 | Arm C: no denial for that `tool_use` AND `victim.txt` is deleted |
| V6 | No arm hit the declared `--max-turns 3` cap or the 180 s `timeout` (exit 124); hitting either is INCONCLUSIVE (REQ-PSD-005), even when a denial or execution was recorded before the cap |

**Outcome table (arm A, given V1–V6):**

| Arm A observation | Result |
|---|---|
| no denial on the PowerShell `tool_use`, `victim.txt` deleted | GAP CONFIRMED → Branch P (REQ-PSD-006) |
| denial on the PowerShell `tool_use`, `victim.txt` survives | GAP ABSENT → Branch N (REQ-PSD-007) |
| any other combination (denied but deleted, not denied but survives, no PowerShell call) | INCONCLUSIVE → blocker (REQ-PSD-005) |

Directory state corroborates; the stream-json denial record decides. Mechanical alternative first: if a model-free rule-evaluation command is found (REQ-PSD-003), arms A and B run through it and the model arms C/D still run for the tool-path control and the built-in observation.

### §C.2 Scope and exclusion list (REQ-PSD-009/010) — D1

The 47 develop `Bash(...)` denies split as follows. D1 chooses which categories ship; everything not shipped goes on the exclusion list with its reason.

| Category | Rows | Default treatment | Reason if excluded |
|---|---|---|---|
| filesystem-root removal | 9 | excluded | `builtin`: denied by Claude Code's Remove-Item / `cmd` system-path and wildcard checks (spec §A.2) |
| git history-destructive | 14 | counterpart | — |
| disk formatting | 3 | counterpart | — |
| system / process | 11 | counterpart, except `kill -9:*` | `alias-head, unmeasured`: `kill` is a PowerShell alias of `Stop-Process` on Windows, so a rule written as `kill -9` may never match after canonicalization (REQ-PSD-010) |
| database | 10 | counterpart | — |

Counterparts copy the Bash rule text verbatim into the `PowerShell(...)` namespace (git, database, and native-binary syntax is identical under PowerShell). Existing shapes carry over unchanged: `git * push --force*` uses space-suffix wildcards only, never `:*` after a middle `*`.

**D2 (lead / run-phase, not operator):** only relevant if D1 option (d) brings the `Remove-Item` rows in. The earlier "wildcard between verb and path" option `Remove-Item * C:/:*` is withdrawn: it mixes a middle `*` with `:*`, which `TestSettingsTemplateDenyWildcardSyntax` rejects for Bash and which would treat the middle `*` literally. Any wider form must use space-suffix syntax only.

### §C.3 Benign sample and matcher model (REQ-PSD-012)

The over-block test models only the documented semantics: `:*` ≡ trailing ` *`; a trailing ` *` matches the bare prefix or the prefix followed by a space and anything; a non-trailing `*` matches any character sequence; a trailing `*` not preceded by a space (glued to text, e.g. `--hard*`, `--force*`) matches any character sequence, including empty; comparison is case-insensitive (PowerShell). It does **not** model alias canonicalization or compound-command splitting; samples that depend on either are listed as a Gap, not asserted.

- **Benign (must NOT match any `PowerShell(...)` deny):** `Remove-Item ./build -Recurse -Force`, `Get-ChildItem C:/`, `git push origin HEAD`, `git clean -n`, `git status`, `Format-Table`, `truncate -s 0 app.log`, `redis-cli GET key`.
- **Known-deny controls (MUST match):** `git push --force origin main`, `git clean -fdx`, `git -C repo reset --hard HEAD`, `Format-Volume -DriveLetter D`, `redis-cli FLUSHALL`, `psql -c DROP TABLE t`.

`truncate -s 0 app.log` is deliberate: case-insensitive matching makes a `PowerShell(TRUNCATE:*)` counterpart match the lowercase `truncate` utility, which the Bash rule was not written to block. The operator resolves it at Kickoff through the D1 `TRUNCATE` sub-choice (ship with accepted over-block, or exclude with reason `case-fold over-block`); it may not silently pass.

The test proves it can fail: a mutation that replaces the matcher with an always-false function must turn the known-deny controls red (observed once, recorded in §E.2).

### §C.4 Closed-world guard (REQ-PSD-014)

The test reads the rendered template deny list and a mapping declared in the test (Bash rule → PowerShell counterpart) plus an exclusion list (Bash rule → reason). It fails when: a Bash deny is neither mapped (with its counterpart present) nor excluded; a PowerShell deny maps to no Bash deny; a PowerShell deny contains `\:`; a PowerShell deny combines a non-trailing `*` with `:*`. The existing mixing and `\:` checks in `TestSettingsTemplateDenyWildcardSyntax` are extended from the `Bash(` prefix to the `PowerShell(` prefix.

## §D Constraints

- Declared caps are hard (§C.1). No background load; no retry loops.
- No new config keys; no hook changes (Out of Scope, D3).
- `audit:`/`source:` fields of tool-policy.yaml do not reach the template (verified by plan-audit iter-1: no template mirror of tool-policy.yaml; codegen writes specifiers only).

## §E Self-Verification

`go test ./internal/template/... ./internal/config/toolpolicy/...`, `make tool-policy-drift-check`, `make build`, template-neutrality and internal-content-leak tests, docs-site 4-locale grep. Full suite left to CI. Branch N runs only the neutrality/leak tests, the AC-PSD-012 set comparison, and the docs grep.

## §F Milestones (priority order)

1. **M0 (High)** — absorb local `develop` (t1207). Record the absorbed SHA (the base for AC-PSD-012). Re-measure `internal/template` tests on the merged tree.
2. **M1 (High)** — preconditions + mechanical search + arms per §C.1; record in progress §E.2; name the branch from the outcome table. Inconclusive → blocker, stop.
3. **Branch N only — M2N (High)** — 4-locale docs note (REQ-PSD-016); AC-PSD-012 set comparison; neutrality tests. Close.
4. **Branch P — M2 (High)** — RED: closed-world guard + over-block test with controls; observe failure (including the mutations of AC-PSD-010).
5. **Branch P — M3 (Medium)** — `tool: "PowerShell"` entries in tool-policy.yaml per D1 scope; `moai tool-policy build`; `make build`; GREEN.
6. **Branch P — M4 (Medium)** — 4-locale docs note (REQ-PSD-016).
7. **M5 (Low)** — neutrality / leak tests, drift check, set comparison, lint.

## §G Anti-Patterns

- Adding rules before M1 records the gap.
- Reading "directory unchanged" as "rule blocked" without the stream-json denial record.
- Running a measurement arm with the Bash tool available, or with user/managed hooks loaded.
- Aiming the measurement at a system-path removal (the built-in deny would mask the result, and failure would be destructive).
- Copying the pre-t1207 `C\:/` form; mixing a middle `*` with `:*`.
- Hand-editing `settings.json.tmpl` instead of regenerating from the SSOT.

## §H Cross-References

- SPEC-V3R6-TOOL-POLICY-SSOT-001; `.moai/reports/t1207/verdict.md`, `audit.md`; `.moai/reports/plan-audit/SPEC-POWERSHELL-DENY-PARITY-001-review-1.md`.
- `https://code.claude.com/docs/en/permission-modes` § Critical paths, § Remove-Item in PowerShell, § bypassPermissions.
- `https://code.claude.com/docs/en/permissions` § PowerShell; `https://code.claude.com/docs/en/tools-reference` § PowerShell tool; `https://code.claude.com/docs/en/headless` § Read session metadata, § Turn off permission prompts.

## Open Decisions

- **D1 (operator, at Kickoff)** — Parity scope. Each option has a concrete enforcement effect only under Branch P:
  - (a) Residual set, 38 rows minus alias-head exclusions (37 counterparts; `kill -9` excluded; filesystem 9 excluded as `builtin`). Covers every Bash deny the built-in does not.
  - (b) Windows-relevant residual: git 14 + disk 3 + database 10 = 27 counterparts. Excludes the 11 system/process rows, which leaves macOS/Linux opt-in `pwsh` sessions without `dd`/`mkfs`/`chmod`/`shutdown` denies.
  - (c) Git only: 14 counterparts. Smallest change; leaves disk-format and database denies absent on PowerShell.
  - (d) (a) plus the 9 filesystem rows as defense-in-depth. For the documented cases it adds nothing beyond the built-in; its only possible effect is on the undocumented native-`rm`-from-`pwsh` case (§B) and on a future change to the built-in, and neither effect is measured.
  - Sub-choice (applies to (a), (b), and (d)): **ship** the `PowerShell(TRUNCATE:*)` counterpart and accept the case-fold over-block of the lowercase `truncate` utility (§C.3), or **exclude** it with reason `case-fold over-block` (the chosen option's counterpart count drops by one, e.g. (a) 37 → 36, (b) 27 → 26).
- **D2 (lead / run-phase)** — Pattern shape for any `Remove-Item` rows, only under D1 (d); space-suffix syntax only (§C.2).
- **D3 (operator issues the card; lead proposes)** — Hook matcher `Write|Edit|Bash` → `Write|Edit|Bash|PowerShell` as a separate card, citing tools-reference ("matching `Bash` alone is not enough").
- **D4 (operator, at Kickoff)** — Accept Branch N as a valid close: no rule change, no Go guard, deliverable = §E.2 evidence + 4-locale docs note.
