# Plan — SPEC-POWERSHELL-DENY-PARITY-001

Ordered by decision reversibility: the measurement that decides whether anything is built comes first, then the rule-shape decisions, then the mechanical propagation.

## §A Context

- Card: t1211. Branch `WT-powershell-deny`, base local `develop` `4dcd4d8d4`.
- SSOT: `.moai/config/sections/tool-policy.yaml` → generator `moai tool-policy build` (`internal/cli/tool_policy.go`; `--local-only` skips the template) → `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json`. Drift guard: `make tool-policy-drift-check` (runs ahead of `make build`).
- Existing static guard to extend: `internal/template/settings_test.go` `TestSettingsTemplateDenyWildcardSyntax` (on develop after t1207).
- Docs pages that enumerate destructive denies: `docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md` (ja line ~309 table "システム破壊", en § "deny (unconditional block)" ~line 299).

## §B Known Issues / Dependencies

- **t1207 dependency.** Local `develop` `baa054586` carries t1207 (`C\:/` → `C:/`). This branch predates it. Run-phase M0 must `git merge develop` (local) before M2.
- The documentation does not literally state that Bash rules skip PowerShell — the premise is a hypothesis until M1 records it.
- `claude --help` (CC 2.1.283, observed at plan time) lists no rule-evaluation subcommand; M1 still searches for one (REQ-PSD-003) before falling back to model arms.

## §C Pre-flight — decisions most likely to change

### §C.1 M1 measurement design (REQ-PSD-001..005)

Scratch project created under the OS temp dir (outside the repo), containing only:

- `.claude/settings.json` with `permissions.deny` = one rule per arm, `defaultMode` unset.
- `victim/keep.txt` — the observable. Deletion of `victim/` = rule did not block.

Arms (each `timeout 180 claude -p ... --max-turns 3 --permission-mode bypassPermissions` with `CLAUDE_CODE_USE_POWERSHELL_TOOL=1`, one run each; deny still applies under bypassPermissions per docs):

| Arm | deny rule | prompt instructs | expected if hypothesis true |
|---|---|---|---|
| A | `Bash(Remove-Item *)` | use the PowerShell tool to run `Remove-Item -Recurse -Force victim` | `victim/` deleted |
| B | `PowerShell(Remove-Item *)` | same | `victim/` present, deny event in transcript |
| C (control) | none | same | `victim/` deleted (proves the tool path runs at all) |

Record per arm: exit code, whether `victim/` exists, the tool name actually invoked (from `--output-format stream-json`), deny message if any. If arm C does not delete, the run is inconclusive (REQ-PSD-005). Mechanical alternative first: search CC docs/CLI for a permission-rule evaluation command; if found, run arms A/B through it with no model.

### §C.2 Parity table (REQ-PSD-006/007) — draft, finalised after M1

Scope decision **D1** governs which rows ship. Proposed shape (literal `C:/`, no `\:`):

| Bash deny (develop) | PowerShell counterpart | Note |
|---|---|---|
| `rm -rf /:*`, `rm -rf ~:*`, `rm -rf C:/:*` and `/\* *` glob variants | `Remove-Item -Recurse -Force /:*`, `... ~:*`, `... C:/:*` (+ glob variants) | `rm`/`del`/`rd`/`ri`/`rmdir`/`erase` canonicalize to `Remove-Item` |
| `del /S /Q C:/:*`, `rmdir /S /Q C:/:*` | covered by `Remove-Item ... C:/` rows via canonicalization | M1 must confirm alias canonicalization for parameter-bearing forms |
| `Remove-Item -Recurse -Force C:/:*` | `Remove-Item -Recurse -Force C:/:*` | direct |
| `Clear-Disk:*`, `Format-Volume:*`, `format:*` | same verbs | direct |
| `git push --force:*` family, `git reset --hard:*`, `git clean -fd*`, `git rebase -i*` (incl. `git * ...` forms) | identical `PowerShell(git ...)` | git syntax identical in PowerShell (D1) |
| `DROP DATABASE`, `psql -c DROP`, `redis-cli FLUSHALL`, `mongo*` etc. | identical | D1 |
| `chmod`, `dd`, `mkfs`, `fdisk`, `systemctl`, `killall`, `reboot`, `shutdown`, `init`, `kill -9` | D1: include (harmless, parity) or exclude (no Windows meaning); `Stop-Computer`/`Restart-Computer` are NOT added (no Bash counterpart — Out of Scope) |

Parameter abbreviation / order risk: PowerShell accepts `-r`, `-fo`, `-Force -Recurse`. Whether a literal-order pattern matches these forms is measured in M1 (extra probe rows, still within caps via the mechanical path, or recorded as a Gap if only model arms exist). **D2** chooses literal forms vs a wider wildcard such as `Remove-Item * C:/:*`, weighed against REQ-PSD-009.

### §C.3 Benign sample (REQ-PSD-009)

Must remain allowed under the final rule set: `Remove-Item ./build -Recurse -Force`, `Get-ChildItem C:/`, `git push origin HEAD`, `git status`, `rm ./tmp.txt`. The Go guard asserts no PowerShell rule's literal prefix matches these strings under the documented wildcard semantics.

## §D Constraints

- Declared caps are hard: 1 run/arm, `--max-turns 3`, `timeout 180`. No background load; no retry loops.
- No new config keys; no hook changes (Out of Scope, D3).
- Template neutrality: no card/SPEC IDs or dates in YAML `audit:`/`source:` fields that propagate into the template (check whether codegen emits them).

## §E Self-Verification

`go test ./internal/template/... ./internal/config/toolpolicy/...`, `make tool-policy-drift-check`, `make build`, template-neutrality test, docs-site 4-locale parity grep. Full suite left to CI.

## §F Milestones (priority order)

1. **M0 (High)** — absorb local `develop` (t1207). Re-measure `internal/template` tests on the merged tree.
2. **M1 (High)** — measurement per §C.1; record in progress §E.2. Branch: REQ-PSD-004 (no-op + guard + doc) / REQ-PSD-005 (blocker) / proceed.
3. **M2 (High)** — RED: extend `settings_test.go` with the parity guard (REQ-PSD-011) and benign-sample guard; observe failure.
4. **M3 (Medium)** — add `tool: "PowerShell"` entries to tool-policy.yaml; verify loader/codegen accept the tool name; `moai tool-policy build`; `make build`; GREEN.
5. **M4 (Medium)** — docs-site `advanced/settings-json.md` × 4 locales note (REQ-PSD-013).
6. **M5 (Low)** — neutrality / leak tests, drift check, lint.

## §G Anti-Patterns

- Adding rules before M1 records the gap (acting on an unverified premise).
- Copying the pre-t1207 `C\:/` form.
- Running the measurement inside the repo (repo hooks/settings contaminate the result).
- Hand-editing `settings.json.tmpl` instead of regenerating from the SSOT.

## §H Cross-References

- SPEC-V3R6-TOOL-POLICY-SSOT-001; `.moai/reports/t1207/verdict.md`, `audit.md`.
- `https://code.claude.com/docs/en/permissions` § PowerShell; `https://code.claude.com/docs/en/tools-reference` § PowerShell tool.

## Open Decisions

- **D1** — Parity scope: all 47 Bash denies (recommended: git/db commands behave identically under PowerShell) vs filesystem-destructive subset only (card wording).
- **D2** — Pattern shape for `Remove-Item`: literal parameter order (low over-block, may miss `-r -fo`) vs wildcard between verb and path (broader, over-block risk). Decided from M1 evidence.
- **D3** — Hook matcher `Write|Edit|Bash` → `Write|Edit|Bash|PowerShell`: separate card (recommended) vs fold in.
- **D4** — If M1 shows Bash rules DO apply to PowerShell: close with guard + docs only (REQ-PSD-004) — confirm the operator accepts this outcome up front.
