# Progress — SPEC-POWERSHELL-DENY-PARITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts (Tier M): spec.md, plan.md, acceptance.md, progress.md. Status: draft. Version 0.2.0.
- Requirements: 18 GEARS (REQ-PSD-001..018), split by branch (P = parity, N = no-rule). ACs: 14 matrix rows + 7 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-POWERSHELL-DENY-PARITY-001` → PASS; uniqueness: 0 existing matches.
- Plan-audit iter-1: FAIL 0.66 (`.moai/reports/plan-audit/SPEC-POWERSHELL-DENY-PARITY-001-review-1.md`); v0.2.0 addresses D1–D10 (blocking) and D11–D15 (optional). Awaiting iter-2 re-audit.
- Premise status: built-in PowerShell removal protection verified from vendor docs (spec §A.2); "Bash denies do not reach the PowerShell tool" is UNVERIFIED; run-phase M1 measures it on a residual-set command before any fix.
- Dependency: absorb local `develop` carrying t1207 (`baa054586`) in M0.
- Open decisions D1 (operator), D2 (lead/run-phase), D3 (operator card issuance), D4 (operator) — plan.md § Open Decisions. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Run-phase delivered by manager-develop (cycle_type=tdd) in worktree `.claude/worktrees/t1211`, branch `WT-powershell-deny`. Operator Kickoff approved in lane. Decisions: D1 = (a) residual set with `kill -9` excluded, TRUNCATE sub-choice = EXCLUDE (reason `case-fold over-block`) → 36 counterparts; D2 moot (no `Remove-Item` rows); D3 = separate card (hook matchers untouched); D4 accepted (moot, Branch P taken).

### M0 — develop absorption

- BASE (absorbed local `develop`) = `5ac030965`; branch HEAD after absorption = `8337fdbbd` (merge commit). t1207 is contained (Windows Bash rules use literal `C:/`; `grep -c 'PowerShell(.*\\\\:' TMPL` and the existing `\:` guard both green on the final tree).
- Pre-flight on `8337fdbbd`: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; template deny array: 55 entries, 47 `Bash(`, 0 `PowerShell(`.

### M1 — measurement (run by the lane outside the repository; evidence copied into `.moai/reports/t1211/m1/`)

Preconditions (AC-PSD-003), captured before arm A (lane-reported): `pwsh -NoProfile -Command '$PSVersionTable.PSVersion.Major'` → `7`; `claude --version` → `2.1.283 (Claude Code)` (re-observed by manager-develop: `2.1.283 (Claude Code)`); managed settings directory absent.

Model-free search (AC-PSD-002): `claude --help` (2.1.283) subcommands observed by manager-develop — `agents, attach, auth, auto-mode, doctor, gateway, import, install, logs, mcp, plugin|plugins, project, respawn, rm, setup-token, stop|kill, ultrareview, update|upgrade`; grep for `permission|rule|evaluat|check` finds only session flags (`--permission-mode`, `--permission-prompts`, `--dangerously-skip-permissions`). No rule-evaluation command exists → model arms used (REQ-PSD-003 satisfied by the search).

Invocation (every arm, own fresh scratch git project, arm settings committed, `victim.txt` + `victim-dir/keep.txt` untracked):
`CLAUDE_CODE_USE_POWERSHELL_TOOL=1 timeout 180 claude -p "<prompt>" --setting-sources project --safe-mode --strict-mcp-config --tools PowerShell --permission-mode bypassPermissions --max-turns 3 --output-format stream-json --verbose`

Evidence files (sha256 measured by manager-develop with `shasum -a 256`):

| File | sha256 |
|---|---|
| `m1/A.jsonl` | `3b300ad8d4f94dde4a434b873da4df34e76aabbc8fc015a36ed81dbfdab3d8d6` |
| `m1/B.jsonl` | `38efe665d0db4907fabc9f6776a905ec3b52dd0836b251ddac4c6523c8b11afd` |
| `m1/C.jsonl` | `5f842483bed38b51b4536c0e121155896e1f9b27b0e1fb759be02694ae7992c1` |
| `m1/D.jsonl` | `fe4cda50dabab8dcc6ee98ef0bef697f15a60818e449aa215cfb65210feade55` |
| `m1/VOID-stray-root.jsonl` | `7595eecf94dff77ae2345d162f3639fb99caeee4359bd7634bc33c923622d0ef` |
| `m1/*.err` (all five) | `e3b0c442…b855` (empty) |

Per-arm record (AC-PSD-001) — parsed from the jsonl by manager-develop (`system/init`, `tool_use`, `tool_result`, `permission_denied`, `result`):

| Arm | `permissions.deny` (`m1/<arm>.settings.json`) | exit | `system/init` tools | PowerShell `tool_use` | tool_result | `permission_denied` / `permission_denials` | target after run | turns | timeout |
|---|---|---|---|---|---|---|---|---|---|
| A | `["Bash(git clean -fdx:*)"]` | 0 (lane) | `['PowerShell']` | `git clean -fdx` | `Removing victim-dir/` `Removing victim.txt` (is_error False) | none / `[]` | `victim.txt` deleted | 2 | no |
| B | `["PowerShell(git clean -fdx:*)"]` | 0 (lane) | `['PowerShell']` | `git clean -fdx` | `Permission to use PowerShell with command git clean -fdx has been denied.` (is_error True) | 1 event, `decision_reason_type: rule` / `[{tool_name: PowerShell, command: git clean -fdx}]` | `victim.txt` survives | 2 | no |
| C | `[]` | 0 (lane) | `['PowerShell']` | `git clean -fdx .` | `Removing victim-dir/` `Removing victim.txt` (is_error False) | none / `[]` | `victim.txt` deleted | 2 | no |
| D | `[]` | 0 (lane) | `['PowerShell']` | none issued | — | none / `[]` | `victim-dir/keep.txt` survives | 2 | no |

All five runs: `result.subtype = success`, `terminal_reason = completed`, `stop_reason = end_turn`, `num_turns = 2` (< cap 3).

Validity (AC-PSD-004):

| # | Holds | Evidence line |
|---|---|---|
| V1 | yes | pwsh major `7` |
| V2 | yes | every arm `system/init` tools = `['PowerShell']` |
| V3 | yes | A/B `tool_use PowerShell 'git clean -fdx'`, C `'git clean -fdx .'` |
| V4 | yes | B `permission_denied` + `permission_denials` on that tool_use; victim survives |
| V5 | yes | C no denial; tool_result `Removing victim.txt` |
| V6 | yes | all `num_turns 2`, exit 0, no 124 |

Outcome: arm A shows no denial on the PowerShell call and `victim.txt` deleted → **GAP CONFIRMED → Branch P** (REQ-PSD-006).

Arm D (AC-PSD-005) — observation, selects no branch: the model issued no PowerShell call (stream carries a `model_refusal_fallback` system event, `api_refusal_category: cyber`, fallback `claude-opus-4-8`), no denial recorded, `victim-dir/keep.txt` survives. The built-in wildcard removal deny was therefore **not observed** either way; no REQ is marked satisfied by it.

Disclosure (verbatim from the lane): one stray invocation was run by mistake from the scratch root (not a git repo, no arm settings, backgrounded with `&`); `git clean` failed there; no observable in any arm changed; its output is kept as `VOID-stray-root.*` and it is not counted as any arm. (Parsed: cwd = scratch root, tool_result `fatal: not a git repository`, `permission_denials: []`.)

### M2 — RED (E8)

New guard `internal/template/settings_powershell_deny_test.go` (closed-world parity, over-block with controls, in-test mutants) and the `PowerShell(` extension of `TestSettingsTemplateDenyWildcardSyntax` written before any rule. Command: `go test ./internal/template/ -run TestSettingsTemplatePowerShellDeny -count=1 -v` → exit 1 (`.moai/reports/t1211/run/red.txt`):

```
--- FAIL: TestSettingsTemplatePowerShellDenyParity (0.00s)
    --- FAIL: TestSettingsTemplatePowerShellDenyParity/darwin (0.00s)
    --- FAIL: TestSettingsTemplatePowerShellDenyParity/linux (0.00s)
    --- FAIL: TestSettingsTemplatePowerShellDenyParity/windows (0.00s)
--- FAIL: TestSettingsTemplatePowerShellDenyNoOverBlock (0.00s)
    settings_powershell_deny_test.go:186: Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(Clear-Disk:*)
    settings_powershell_deny_test.go:186: Bash deny has neither a PowerShell counterpart nor an exclusion: Bash(DELETE FROM:*)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.365s
```

108 "neither" lines = 36 rows × 3 platforms.

### M3 — GREEN

- `.moai/config/sections/tool-policy.yaml`: new §D.1h with 36 `tool: "PowerShell"` deny entries (verbatim Bash patterns) and the exclusion comment. `grep -c 'tool: "PowerShell"'` → `36`.
- `moai tool-policy build` (via `go run ./cmd/moai tool-policy build`) → `.claude/settings.json [json]: allow=114 ask=0 deny=91 env_gated_skipped=5`; stderr: `Info: template ... has a render-time-conditional permissions block (git_mode gating); skipping template target to preserve the conditional.` The generator refuses the template by design, so the 36 rows were written into `TMPL` by a script from the same list (template order, appended after the Bash rows). Parity of the two surfaces is measured: template `PowerShell(` set == local `PowerShell(` set → `True 36`.
- `go test ./internal/template/ -run TestSettingsTemplate -count=1 -v` → exit 0 (`run/green.txt`); all PowerShellDeny subtests and `TestSettingsTemplateDenyWildcardSyntax/{darwin,linux,windows}` PASS.

### AC matrix (Branch P)

| AC | Status | Evidence |
|---|---|---|
| AC-PSD-001 | PASS | per-arm table above |
| AC-PSD-002 | PASS | `claude --help` search above |
| AC-PSD-003 | PASS | preconditions above |
| AC-PSD-004 | PASS | V1–V6 table; Branch P |
| AC-PSD-005 | PASS | arm D recorded as non-selecting observation (built-in not observed) |
| AC-PSD-006 | PASS | `setcmp.py` on final tree: `PowerShell deny count 36`; guard green (every Bash deny mapped or excluded: 36 + 11 = 47); D1 scope 37 − TRUNCATE = 36; yaml `tool: "PowerShell"` = 36 |
| AC-PSD-007 | PASS | `grep -c 'PowerShell(.*\\\\:' TMPL` → `0` |
| AC-PSD-008 | PASS | `TestSettingsTemplatePowerShellDenyNoOverBlock` PASS; always-false matcher mutant → exit 1, 6 × `destructive command … is not blocked by any PowerShell deny` (`run/mutant-matcher-false.txt`); benign control: shipping `PowerShell(TRUNCATE:*)` → `benign command truncate -s 0 app.log is blocked by PowerShell(TRUNCATE:*)` + `excluded Bash deny also has a PowerShell counterpart: Bash(TRUNCATE:*)` (`run/mutant-truncate-shipped.txt`) |
| AC-PSD-009 | PASS | `make tool-policy-drift-check` → exit 0 (`ok … internal/config/toolpolicy`); `make build` → exit 0 (catalog.yaml byte-unchanged) |
| AC-PSD-010 | PASS | out-of-band mutations of `TMPL`, each `go test ./internal/template/ -run TestSettingsTemplate` → exit 1 naming the rule: (a) `…neither a PowerShell counterpart nor an exclusion: Bash(git clean -fdx:*)`; (b) `…: Bash(wipefs:*)`; (c) `PowerShell deny escapes ':' …: PowerShell(git reset --hard C\:/:*)`; (d) `PowerShell deny mixes a non-trailing '*' with ':*': PowerShell(git * clean -fdx:*)` (`run/mutant-{a,b,c,d}.txt`). The same four mutants also run permanently as `TestSettingsTemplatePowerShellDenyGuardDetectsMutations` subtests. Files restored from backup; final tree re-tested green |
| AC-PSD-011 | PASS | `go test ./internal/template/... ./internal/config/toolpolicy/... -count=1` → exit 0 (`run/scoped-tests.txt`) |
| AC-PSD-012 | PASS | `setcmp.py <root> 5ac030965` → `base Bash deny 47 final Bash deny 47 identical True`, `PowerShell in allow 0 in ask 0`, `defaultMode base None final None`; positive control `--drop-one-bash` → `identical False  only base: ['Bash(git clean -fdx:*)']`. Local `.claude/settings.json` diff is 36 added lines only (`defaultMode` unchanged) |
| AC-PSD-013 | PASS | all four `docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md` in the run commit (see §E.3 `run_commit_sha`) |
| AC-PSD-014 | N/A (branch P) | — |

Other checks: `go vet ./internal/template/ ./internal/config/toolpolicy/` → exit 0; `gofmt -l` → empty; `golangci-lint run ./internal/template/` → `0 issues.`; `go test ./internal/cli/ -run ToolPolicy` → ok; `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

### Gaps

- Arm exit codes and post-run target existence are lane-reported; the jsonl corroborates A/C deletion via tool_result text and B via the denial, but not arm D's `keep.txt`.
- The built-in wildcard deny (arm D) was not observed; native `rm` from `pwsh` on a system path is unmeasured (spec §A.2).
- The matcher model does not cover alias canonicalization or compound-command splitting; samples depending on either are not asserted.
- `kill -9` alias-form behaviour is unmeasured (excluded).
- Full test suite not run locally (CI on the develop push).

### Residual risk

- Measured on macOS with the opt-in PowerShell tool; transfer to Windows is inferred.
- Template and YAML are kept in step by the guard + a measured set equality, not by the generator (the generator skips the git_mode-gated template).
- PowerShell hook matchers remain `Write|Edit|Bash` (D3, separate card), so hook-level guards still do not see PowerShell commands.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: d601647e0   # C1 evidence+status 49c0fe453; C2 guard+rules+docs d601647e0
run_status: complete
branch: P
ac_pass_count: 13
ac_fail_count: 0
ac_na: [AC-PSD-014]
preserve_list_post_run_count: 47   # Bash deny rows, identical to BASE 5ac030965
l44_pre_commit_fetch: not-run      # lane-local; push is the lead's
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 11
m1_to_mN_commit_strategy: "C1 M1 evidence + status; C2 guard + rules + docs + run evidence; C3 SHA backfill"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: 7017309f0   # backfilled; the sync commit could not cite its own hash
sync_status: complete
card: t1211
run_commits: [49c0fe453, d601647e0, 6b8279daf]
evidence:
  m1_measurement: .moai/reports/t1211/m1/     # arms A/B/C/D + VOID-stray-root (jsonl + settings + err)
  run: .moai/reports/t1211/run/               # red/green, mutants (a/b/c/d, matcher-false, truncate-shipped), drift, lint, vet, make-build, scoped-tests, cli-toolpolicy
b12_self_test_a: "grep -c 'SPEC-POWERSHELL-DENY-PARITY-001' CHANGELOG.md -> 0 (pre-emission)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 14 (AC-PSD-001..014; CHANGELOG entry cites 14: 13 PASS + AC-PSD-014 N/A branch P)"
b12_self_test_c: "ls on every path cited in the CHANGELOG entry -> all exist"
changelog_entry_position: "[Unreleased] ### Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (status only; updated already 2026-09-26)"
lead_remeasure_before_sync:
  - "go test ./internal/template/ -run TestSettingsTemplate -count=1 -> ok"
  - "make tool-policy-drift-check -> exit 0"
```

**Addendum (sync-audit follow-up).** `679cbbb26` applied sync-audit findings F1/F2 (YAML-to-template PowerShell deny parity guard); this commit applies F3/F4/F5 (docs-site built-in protection caveat in 4 locales, CHANGELOG row-count and guard wording, this `sync_commit_sha` backfill).

**Deviation recorded for sync-audit judgment (REQ-PSD-013).** REQ-PSD-013 expected `moai tool-policy build` to regenerate the template. That command deliberately skips the git_mode-conditional template (`settings.json.tmpl`), so the 36 template rows were written by script from the same list the YAML SSOT carries. Equivalence is shown by a measured set comparison plus the closed-world guard `internal/template/settings_powershell_deny_test.go`, not by the generator. The auditor judges whether this satisfies REQ-PSD-013 or needs a follow-up.

**Operator decisions carried into this close.**
- D1: mirror set = the 37 residual Bash denies minus TRUNCATE → **36 PowerShell rows**.
- TRUNCATE excluded from the mirror.
- D4: accepted.
- D3 (PowerShell hook matchers still `Write|Edit|Bash`) → deferred to card **t1224**.

**Gaps (not observed).**
- Arm D: the built-in wildcard deny was not observed (the model refused before the tool call).
- Native `rm` under `pwsh` on a system path: not measured.
- `kill -9` alias-form exclusion: not measured.
- All measurement ran on macOS with the opt-in PowerShell tool; transfer to Windows is extrapolated.
- Full test suite not run locally; the verdict is CI on the lead's develop push.

**Residual risk.** Template and YAML stay in step through the guard and the set equality, not through the generator; hook-level guards still do not see PowerShell commands until t1224 lands.
