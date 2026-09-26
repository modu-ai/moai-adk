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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
