# progress.md — SPEC-HOOK-GUARD-POWERSHELL-FORMS-001

Card t1255 · factory lane worker-65 · branch `WT-powershell-guard-debt` · base `19b5321c1`.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-26

Plan-auditor iter-1 PASS-WITH-DEBT 0.85, commission fixes resolved; final PASS 0.95 (verdict `.moai/reports/plan-audit/SPEC-HOOK-GUARD-POWERSHELL-FORMS-001-review-1.md`, audited commit `2286d77c3`). Implementation Kickoff Approval: satisfied by the operator's standing 자율(autonomous) policy (lead dispatch 2026-09-26, card t1255, factory lane worker-65) — recorded in plan.md §A, not skipped.

## §E.2 Run-phase Evidence

### M0 — `-EncodedCommand` spelling measurement (REQ-HGF-010)

Measured 2026-09-26 in the **primary checkout** (the worktree session guard refuses `pwsh`; the worktree lane could not host it). The measurement is pwsh behavior and is independent of tree state; the card base on the lane is `19b5321c1`.

**Recipe** (acceptance.md §E, run verbatim): payload `B64=$(printf 'Write-Output ok' | iconv -t UTF-16LE | base64)`; per-candidate invocation `timeout 60 pwsh -NoProfile -NonInteractive $s "$B64" 2>&1 | head -1`; caps: 12 candidates (11 loop spellings + attached-colon `-enc:<B64>` as its own 12th invocation), one pass, wall-clock ≤ 10 min, ≤ 2 turns, foreground only, per-invocation `timeout 60`, no background load.

**Verbatim stdout rows** (12 candidates; scratch `/tmp/t1255-m0/` — this transcription is the durable record):

| # | Spelling | pwsh verdict (verbatim first line) | Exit |
|---|---|---|---|
| 1 | `-e` | `[ok]` | 0 |
| 2 | `-ec` | `[ok]` | 0 |
| 3 | `-en` | `[ok]` | 0 |
| 4 | `-enc` | `[ok]` | 0 |
| 5 | `-enco` | `[ok]` | 0 |
| 6 | `-encodedc` | `[ok]` | 0 |
| 7 | `-EncodedCommand` | `[ok]` | 0 |
| 8 | `-ENC` | `[ok]` | 0 |
| 9 | `--EncodedCommand` | `[ok]` | 0 |
| 10 | `/enc` | `[ok]` | 0 |
| 11 | `–enc` (U+2013) | `[ok]` | 0 |
| 12 | `-enc:<B64>` | `[The argument '-enc:VwByAGkAdABlAC0ATwB1AHQAcAB1AHQAIABvAGsA' is not recognized as the name of a script file. Check the spelling of the name, or if a path was included, verify that the path is correct and try again.]` | 0 |

**Deciding facts**:

1. **All 11 prefix spellings are accepted, INCLUDING the U+2013 unicode-dash form `–enc`** — the only accepted spelling the current detector misses (`powerShellParameterName` matches ASCII `-`/`--`/`/` prefixes only).
2. **The attached-colon form `-enc:<B64>` is NOT accepted** as an encoded command — pwsh parses it as a script-file argument. Rejection is observable by OUTPUT, not exit code (exit was 0 on every row; the exit code alone never classifies any row).

**Detector comparison (REQ-HGF-011 pin)**:

| Accepted spelling | Detector verdict today | Action |
|---|---|---|
| `-e`, `-ec` | matched (`isEncodedCommandParameter` literal arms) | none |
| `-en` … `-EncodedCommand`, `-ENC`, `--EncodedCommand` | matched (`en`-prefix arm) | none |
| `/enc` | matched (`/` prefix arm) | none |
| `–enc` (U+2013) | **missed** | M5: add unicode-dash prefix handling |
| `-enc:<B64>` (rejected by pwsh) | detector matches anyway (the `:value` Cut is prefix-agnostic) | not required to match; kept as harmless over-match — costs one audit line, never a deny (REQ-HGF-011 superset rationale recorded at the declaration in M5) |

M5 additionally normalizes U+2014 (em dash) and U+2010 (hyphen) under the same err-wide over-match policy — not measured, never deny-capable.

### M1 — RED capture (E8)

**Command** (env-scrubbed, scoped selector): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/ -run TestBranchGuardPSForms -v`
**Exit code**: `1` — **18 FAIL / 22 PASS** (this run, this tree, HEAD `15e75fbcc` + the M1 test file uncommitted).
**Tree**: `WT-powershell-guard-debt` @ `15e75fbcc` (card base, clean) + the new `branch_guard_psforms_test.go`.

**Verbatim RED failing-test names** (one per disguise leg; full output captured at scratch `/tmp/t1255-run/red-m1.txt`, deciding lines transcribed here):

```
    --- FAIL: TestBranchGuardPSForms/f1a/deny/exe-switch (0.58s)
    --- FAIL: TestBranchGuardPSForms/f1a/deny/exe-reset-hard (0.47s)
    --- FAIL: TestBranchGuardPSForms/f1b/deny/single-quoted-target (0.49s)
    --- FAIL: TestBranchGuardPSForms/f1b/deny/double-quoted-exe-target (0.66s)
    --- FAIL: TestBranchGuardPSForms/f1c/deny/backtick-switch (0.84s)
    --- FAIL: TestBranchGuardPSForms/f1c/deny/backtick-rebase (0.70s)
    --- FAIL: TestBranchGuardPSForms/f2/deny/command-payload (0.57s)
    --- FAIL: TestBranchGuardPSForms/f2/deny/short-c-payload (0.60s)
    --- FAIL: TestBranchGuardPSForms/f2/deny/payload-with-comment (0.73s)
    --- FAIL: TestBranchGuardPSForms/f3/demote/dynamic-resolution (0.58s)
    --- FAIL: TestBranchGuardPSForms/alias/saps-logs (0.51s)
    --- FAIL: TestBranchGuardPSForms/alias/start-logs (0.50s)
    --- FAIL: TestBranchGuardPSForms/dl/deny/bash-eval (0.54s)
    --- FAIL: TestBranchGuardPSForms/dl/deny/ps-iex (0.48s)
    --- FAIL: TestBranchGuardPSForms/dl/deny/ps-start-process (0.58s)
    --- FAIL: TestBranchGuardPSForms/f5/enc/u2013-dash (0.46s)
    --- FAIL: TestBranchGuardPSForms/f5/enc/u2014-dash (0.45s)
    --- FAIL: TestBranchGuardPSForms/f5/enc/u2010-hyphen (0.47s)
```

Representative verbatim failure detail (form families F1a/F2/F3/dl; the remaining families share the same two shapes — `want deny` for deny legs, `audit lines = 0, want 1` for demotion legs):

```
=== RUN   TestBranchGuardPSForms/f1a/deny/exe-switch
    branch_guard_psforms_test.go:135: decision = "allow" (reason ""), want deny
=== RUN   TestBranchGuardPSForms/f2/deny/command-payload
    branch_guard_psforms_test.go:135: decision = "allow" (reason ""), want deny
=== RUN   TestBranchGuardPSForms/f3/demote/dynamic-resolution
    branch_guard_psforms_test.go:142: unclassified audit lines = 0 ([]), want 1
=== RUN   TestBranchGuardPSForms/dl/deny/bash-eval
    branch_guard_psforms_test.go:135: decision = "allow" (reason ""), want deny
```

**RED reason (right reason, per verification-completeness §2)**: every FAIL is the under-match the SPEC names — the disguise leg's command reached the guard and was allowed unrecorded (decision="allow", 0 audit lines). It is NOT red because the guard is inert: the 22 PASS include the parity anchors `cmd /c git switch probe` → deny, bare `terraform destroy` → deny, `pwsh -Command "git status" ; git switch probe` → deny (command-position git outside the payload already matched), and every D2 leg (`iex "git status"` → allow + 1 line) — the machinery is alive; the specific forms are what it cannot see.

M2-M5 flip these 18 to their target verdicts; the 22 stay green through every change.

### M2-M5 — implementation and GREEN close (2026-09-26)

Commits (branch `WT-powershell-guard-debt`, each carries `Authored-By-Agent: manager-develop`):

| Commit | Milestone | Content |
|---|---|---|
| `f25c3f84e` | M1 | RED battery `branch_guard_psforms_test.go` + SPEC `draft → in-progress` + M0 record |
| `d557f71a0` | M2 | REQ-HGF-001..004: `normalizeGitExeSuffix`, `substituteCallOperatorTargets` (git-naming targets only), `substituteCommandBackticks`, `extractPowerShellCommandPayload` (full pipeline within the payload); `powerShellParameterName` now returns the lower-cased name its doc comment promised |
| `b1700a18e` | M3 | REQ-HGF-005/006: `constructDynamicResolution` (`& (…) ` call-position subexpression → allow + 1 line); `saps`/`start` join the start-process construct set behind the existing git-word gate |
| `d70d2ad47` | M4 | REQ-HGF-007: `extractLiteralIndirectionOperand` (eval / iex / Invoke-Expression / Start-Process incl. aliases) scans the literal quoted operand against the existing compiled deny list with the bare-form reason; `$`-carrying operands are found-but-empty (fail open) |
| `e03260744` | M2-REFACTOR | staticcheck QF1002: tagged switch in `splitPSTokens` |
| `72c2cdc4d` | M5 | REQ-HGF-009: integration-lock audit comment states the measured scope. REQ-HGF-011: `powerShellParameterName` accepts U+2013 (measured) + U+2014/U+2010 (documented superset, over-match rationale at the declaration); `-enc:<B64>` stays a harmless over-match, not required (M0 row 12) |

**Final GREEN (battery)**: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/ -run TestBranchGuardPSForms` → exit 0, `ok github.com/modu-ai/moai-adk/internal/hook` — **41 subtests, 0 FAIL** (this run, HEAD `72c2cdc4d`): 18 former-RED legs flipped, 23 controls/legit legs never moved.

**Full hook package suite** (same env-scrub, `go test -count=1 ./internal/hook/`): exit 0, `ok … 264.553s` — pre-existing `TestBranchGuard*`/`TestHMP*` families unbroken (run under slot lease `internal-hook-suite`).

**F4 comment verbatim (AC-HGF-011 evidence)** (`sed -n '29,34p' internal/hook/powershell_indirection.go` at `72c2cdc4d`):

```
// integrationLockAuditRelPath is the integration lock's audit log, relative to
// the handler's project root. While a live foreign hold exists it records
// EVERY unclassifiable PowerShell command observed, whether or not it is
// merge-shaped (measured forms: `iex "git status"` and
// `Start-Process git -ArgumentList 'log'` are both logged and both allowed —
// SPEC-HOOK-GUARD-POWERSHELL-FORMS-001 REQ-HGF-009); the guard's other
// fail-open paths keep writing their stderr advisories.
```

**E1 — AC matrix (14/14 PASS)**:

| AC | Verdict | Evidence |
|---|---|---|
| AC-HGF-001 | PASS | E-01 RED → battery `f1a/deny/*` green; `f1a/allow/exe-status` still allow |
| AC-HGF-002 | PASS | E-02/E-03 RED → `f1b/deny/*` green; quoted-prose leg stays allow |
| AC-HGF-003 | PASS | E-04 RED → `f1c/deny/*` green; single-quoted literal stays allow |
| AC-HGF-004 | PASS | E-05/E-06 RED → `f2/deny/*` green; nested-quote mutant stays allow; `cmd /c` control unchanged |
| AC-HGF-005 | PASS | E-07 (0 lines) → `f3/demote/*`: allow + exactly 1 `dynamic-resolution` line; no-git-word leg 0 lines |
| AC-HGF-006 | PASS | E-08 (0 lines) → `alias/saps-logs` +1 `start-process` line; `saps notepad` 0 lines |
| AC-HGF-007 | PASS | E-12 RED → `dl/deny/ps-iex` deny (E-13 reason family); `iex "git status"` + `iex $c` D2 legs unchanged |
| AC-HGF-008 | PASS | E-14 RED → `dl/deny/ps-start-process` deny; `Start-Process notepad` allow |
| AC-HGF-009 | PASS | E-11 RED → `dl/deny/bash-eval` deny; `eval "echo hi"` / `eval "$(printf …)"` / `eval "$ENV:X"` allow |
| AC-HGF-010 | PASS | Battery green with non-zero swept count (41 subtests); REQ-HGF-014 isolation carried (hmpIsolateHome per-test, no t.Parallel, t.TempDir only) |
| AC-HGF-011 | PASS | Comment verbatim above; E-15/E-16 behavior unchanged |
| AC-HGF-012 | PASS | M0 table in §E.2; detector conformance per REQ-HGF-011 pin (U+2013 measured-accepted → detected; `-enc:<B64>` measured-rejected → not required) |
| AC-HGF-013 | PASS | `injection/newline-keeps-one-line` subtest green (1 line, newline escaped inside the quoted field) |
| AC-HGF-014 | PASS | CARD_BASE recompute `git merge-base develop HEAD` = `19b5321c1`; pathspec diff `internal/template/ .claude/` = **0 files**; non-vacuous control same range without pathspec = **8 files** (4 SPEC artifacts + 3 Go sources + 1 test file) |

**E2 — cross-platform build** (this run, HEAD `72c2cdc4d`): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
**E4 — subagent boundary grep**: 1 match, `internal/hook/pre_tool.go:695` (`observeQuestionChannel` observer) — present verbatim in the card base `15e75fbcc`; **NEW matches: 0**.
**E5 — lint**: `golangci-lint run --timeout=2m internal/hook/...` → `0 issues.` (the one NEW issue found mid-run, staticcheck QF1002 on `splitPSTokens`, fixed in `e03260744`). `go vet ./internal/hook/...` exit 0.
**E3 — coverage**: `go test -count=1 -cover ./internal/hook/` → exit 0, `coverage: 86.1% of statements` (quality.yaml target 85).
**E6 — push state**: 6 commits on `WT-powershell-guard-debt`, NOT pushed (lane rule: develop push is the lead's batch act). Worktree kept — this branch is the only copy until the lead merges.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: "72c2cdc4d"
run_status: "complete — all 14 ACs PASS; release-blocking matrix green"
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0   # internal/template/ and .claude/ diff = 0 over the card range (AC-HGF-014)
l44_pre_commit_fetch: "not-run — worktree-isolated card branch; develop absorption belongs to the integration-window holder"
l44_post_push_fetch: "not-applicable — lane does not push; lead batch-pushes develop"
new_warnings_or_lints_introduced: 0   # one transient QF1002 was introduced and fixed within the run (e03260744)
cross_platform_build:
  native: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 4   # branch_guard.go, powershell_indirection.go, pre_tool.go, branch_guard_psforms_test.go
m1_to_mN_commit_strategy: "one commit per milestone (M1 RED / M2 / M3 / M4 / M5) + one REFACTOR style commit"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: "pending-backfill-sync"   # D3 backfill window — placeholder in the sync commit, real SHA backfilled in the following commit
sync_status: "complete — docs surfaces assessed, 3-phase close committed"
b12_self_test_a: "CHANGELOG pre-emission grep '<SPEC-ID>' = 0 matches (exit 1) — safe to emit"
b12_self_test_b: "AC identifiers in acceptance.md = 14 distinct (AC-HGF-001..014); CHANGELOG entry cites 14"
b12_self_test_c: "all claimed paths verified by ls: internal/hook/{branch_guard.go, powershell_indirection.go, pre_tool.go, branch_guard_psforms_test.go}"
changelog_entry_position: "Unreleased → Added, first bullet (above SPEC-CODEX-PARSER-SHAPE-001)"
frontmatter_status_transitions: "spec.md in-progress → completed (merged 3-phase close, single sync commit); updated: 2026-09-26 unchanged"
canary_compliance_check: "template tree UNTOUCHED — AC-HGF-014 recompute (merge-base develop HEAD = 19b5321c1; pathspec internal/template/ + .claude/ diff = 0 files; bare-range control 8 files)"
mx_tag_validation: "pre-existing @MX:ANCHOR/@MX:NOTE/@MX:REASON tags in branch_guard.go / pre_tool.go intact (grep observed this run); new helpers are unexported with 1-2 call sites each (fan_in < 3, no ANCHOR threshold crossed); no tag debt introduced"
```

### Docs-surface assessment (sync-phase)

| Surface | Verdict | Reason (measured this run) |
|---|---|---|
| CHANGELOG.md | **changed** | new entry under Unreleased → Added — user-visible guard behavior change; follows the t1224/t1211 sibling shape and the file's Keep-a-Changelog structure |
| README.md / README.ko.md | no change needed | the 4 PowerShell mentions per file are install/platform-support rows (lines 268/270/335/710 in README.md); guard behavior is not documented there |
| docs-site (ko/en/ja/zh) | no change needed | `hooks-guide.md` documents the t1224 PowerShell matcher registration only; no sentence it carries becomes false — this SPEC adds form classification the guide never enumerated. No new pages invented (dispatch constraint) |
| internal/template/ + .claude/ | UNTOUCHED | AC-HGF-014 (above) |

### E-17 sync-phase confirmation (acceptance.md E-17 line-range drift resolved)

acceptance.md E-17 cites `sed -n '29,32p' internal/hook/powershell_indirection.go` against tree `19b5321c1` (the pre-fix 3-line comment). After the F4 fix the comment occupies lines **29-34**. Re-verified verbatim at HEAD `2cb4511b2` (this run):

```
// integrationLockAuditRelPath is the integration lock's audit log, relative to
// the handler's project root. While a live foreign hold exists it records
// EVERY unclassifiable PowerShell command observed, whether or not it is
// merge-shaped (measured forms: `iex "git status"` and
// `Start-Process git -ArgumentList 'log'` are both logged and both allowed —
// SPEC-HOOK-GUARD-POWERSHELL-FORMS-001 REQ-HGF-009); the guard's other
// fail-open paths keep writing their stderr advisories.
```

The scope claim E-17 intended — "non-merge commands are also logged" — is stated verbatim by lines 31-32 ("EVERY unclassifiable PowerShell command observed, whether or not it is merge-shaped"). This fulfills AC-HGF-011's green path ("the comment states the measured scope"). acceptance.md body NOT modified (forbidden surface); this §E.4 record is the confirmation of record.

### Run-agent reported gaps — sync-phase disposition (sync-auditor reading list)

1. **U+2014/U+2010 documented superset** — stays a gap by design: REQ-HGF-011 is satisfied via its documented-superset branch (over-match costs one audit line, never a deny; rationale recorded at the `isEncodedCommandParameter` declaration). No sync action.
2. **Boundary grep baseline row** — §E.2 E4: 1 match at `internal/hook/pre_tool.go:695`, present verbatim in the card base `15e75fbcc`; NEW matches 0. The baseline row is that pre-existing `observeQuestionChannel` observer. No sync action.
3. **E-17 drift** — resolved above (line range 29→34, text re-verified verbatim).
