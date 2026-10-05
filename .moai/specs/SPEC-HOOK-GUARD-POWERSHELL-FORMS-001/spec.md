---
id: SPEC-HOOK-GUARD-POWERSHELL-FORMS-001
title: "Classify the PowerShell tool-call forms the shell guards pass without record (t1224 debts F1-F5)"
version: "0.2.0"
status: completed
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook (branch_guard.go, powershell_indirection.go, pre_tool.go, integration_lock_guard.go)"
lifecycle: spec-anchored
tags: "hooks, powershell, branch-guard, denylist, unclassifiable, windows, t1224, t1255"
era: V3R6
tier: M
related_specs: [SPEC-HOOK-MATCHER-POWERSHELL-001]
depends_on: [SPEC-HOOK-MATCHER-POWERSHELL-001]
---

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec | Initial draft for card t1255 (t1224 follow-up debts F1-F5 + denylist-wrapping side issue). Per-form detection decision table (§A.3) resolved from the probe evidence re-measured on this tree at `19b5321c1`; 14 requirements, 14 acceptance criteria. |
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iter-1 PASS-WITH-DEBT 0.85 (`.moai/reports/plan-audit/SPEC-HOOK-GUARD-POWERSHELL-FORMS-001-review-1.md`, audited commit `2286d77c3`) commission fixes: D1 git-word gate qualifier added to REQ-HGF-005; D2 AC-HGF-010/012 RED cells wired to four-element ledger rows E-18/E-19 (empty-sweep selector run + progress.md §E.2 placeholder read, both measured at `2286d77c3`); D3 AC-HGF-014 diff range switched to the merge-base recompute form + empty-sweep non-vacuous rule; D4 M0 candidate count reconciled to 11 loop spellings + 1 attached-colon `-enc:<B64>` = 12; D5 AC-HGF-001 E-09 citation corrected (wrapper-form control, not the plain form); D6 §E cross-references extended with the deny-pattern SSOT family; D8 pre-flight conflict scan widened to all referenced SPECs. D7 (commit trailer) is the lane's. REQ count (14) and AC count (14) unchanged. |

## §A Background

### A.1 Provenance

- Card t1255 (lead dispatch 2026-09-26, factory lane worker-65). Parent SPEC: SPEC-HOOK-MATCHER-POWERSHELL-001 (`status: completed`, landed in develop via merge `0119a8f2c`). This SPEC closes the follow-up debts that SPEC's verdict recorded as residual risk (`.moai/reports/t1224/verdict.md` § Residual-risk and § 후속 카드용 재현, sha256 `ef598289…`).
- **Depends-on satisfied**: SPEC-HOOK-MATCHER-POWERSHELL-001 is `completed` on develop; the card's precondition "t1224 착지 뒤 착수" holds.
- **Fresh RED evidence on this tree**: the t1224 overlay probe (`.moai/reports/t1224/followup-probe_test.go.txt`, sha256 `868c34db…`) was re-run on this worktree at `19b5321c1` (clean, card base) on 2026-09-26 — `go test -count=1 -overlay … ./internal/hook/ -run TestAuditProbe -v -o /dev/null`, exit 0, `ok … 6.197s`. The verdict's original rows (measured at `f4d62e5c3`) reproduce unchanged here; acceptance.md §B carries the rows measured on **this** tree as the pinned RED-now baseline.

### A.2 The measured forms (tree `19b5321c1`)

Every form below reaches the guards and passes **without a deny and without an audit line** (`decision="allow" unclassifiedLines=0 construct=""`), while the control `cmd /c git switch probe` denies:

| # | Form | Class | Current verdict at `19b5321c1` |
|---|---|---|---|
| F1a | `git.exe switch probe` | branch guard misses the `.exe`-suffixed executable spelling | allow, 0 audit lines |
| F1b | `& 'git' switch probe`, `& "git.exe" switch probe` | call operator + quoted call target — the quoted span collapses to a placeholder before the pattern scan | allow, 0 audit lines |
| F1c | ``git swi`tch probe`` | PowerShell backtick escape splits the subcommand token | allow, 0 audit lines |
| F2 | `pwsh -Command "git switch probe"`, `powershell -c "git switch probe"` | the quoted payload is executed code, but the quoted-span collapse treats it as data | allow, 0 audit lines |
| F3 | `& (Get-Command git) switch probe` | executable resolved at runtime by a subexpression | allow, 0 audit lines |
| — | `saps git -ArgumentList 'switch','probe'` | `saps` is a documented `Start-Process` alias; the indirection token switch matches only the full spelling | allow, 0 audit lines |
| — | `iex "git status"` / `$c = 'git switch probe'; iex $c` | **D2 baseline (working)** — allow + exactly one unclassifiable line | allow, 1 line |

Deny-list side issue (부수), same tree: bare `terraform destroy` denies; `iex "terraform destroy"`, `Start-Process terraform -ArgumentList 'destroy'`, and Bash `eval "terraform destroy"` all **allow** — the deny list does not look inside indirection operands, on either shell.

F4: the integration-lock audit log comment (`powershell_indirection.go:29-31`) describes a narrower scope than the measured behavior — under a live foreign hold, `iex "git status"` and `Start-Process git -ArgumentList 'log'` are **both logged although neither is merge-shaped**.

F5: the `-EncodedCommand` spellings pwsh accepts are unmeasured; the detector (`isEncodedCommandParameter`) matches documented prefixes only, and the Unicode-dash form (`–enc`, U+2013) is unmeasured.

### A.3 Per-form decision table (the design axes, resolved)

| Form | Decision | Rationale |
|---|---|---|
| F1a `.exe` suffix | **(a) DETECTION EXPANSION** — deny path via the existing query-vs-mutate verdicts | Statically decidable: the executable spelling carries `.exe`; the subcommand alone decides the verdict, so `git.exe status` keeps passing exactly as `git status` does. Parity with the plain spelling is the point of the guard. |
| F1b quoted call target | **(a) DETECTION EXPANSION** | The call operator `&` gives the following quoted span executable-name meaning — it is the one site where a quoted span is command, not data. The general quoted-argument collapse is untouched, so quoted-prose protection (the `moai todo add "… git switch …"` incident class) survives everywhere else. |
| F1c backtick escape | **(a) DETECTION EXPANSION, bounded** | PowerShell itself de-escapes backtick in command position; the guard reaches the decision the de-escaped command reaches. Bounded to command position: a backtick inside a single-quoted string is literal data and keeps the quoted-span treatment. |
| F2 `-Command` payload | **(a) DETECTION EXPANSION** | The `-Command` payload is executed code. The same branch-state scan — including its in-payload quoted-span and comment treatment — applies to the payload, so query payloads pass and mutate payloads deny. Parity anchor: `cmd /c git switch probe` already denies; leaving `pwsh -Command "…"` allowed is an undefended asymmetry between equivalent wrappers. |
| F3 dynamic resolution | **(b) UNCLASSIFIABLE DEMOTION** — allow + one D2 audit line | The resolved executable is decided at runtime: `& (<expression>) switch` does not statically name git, so a deny would fire without the positive evidence the fail-open norm requires. One audit line makes the gap visible — the same treatment REQ-HMP-010 chose for encoded commands, which are likewise undecidable without evaluation. |
| `saps` / `start` alias | **EXPAND the existing demotion** | Documented `Start-Process` aliases; the fully spelled form already logs one line. The alias joins the same construct so behavior is spelling-independent. Over-match is bounded by the existing git-word gate — no `git`, no line. |
| 부수 denylist wrapping | **FIX in this card — both shells, literal operands only** | A literal quoted operand of `eval` / `iex` / `Start-Process` is positive evidence the deny-list pattern will execute (unlike F3, nothing is resolved at runtime). Non-literal operands stay fail-open. Fixing only PowerShell would break the Bash-parity the threat model is built on (`eval` is the same measured hole); deferring leaves a one-quote bypass of a hard deny. This subset (M4) is separable — see §C.4. |

## §B Requirements (GEARS)

### B.1 Branch-state scan forms (F1, F2)

- **REQ-HGF-001** — **Where** the branch guard is enabled, **when** a tool command names the git executable with an `.exe` suffix, optionally behind a path (the measured form is `git.exe switch probe`), and the remaining command text matches a branch-state pattern in the primary checkout with no exemption, the pre-tool hook shall deny the call with the same `BRANCH_GUARD_VIOLATION:` reason family as the plain spelling, and a `.exe`-suffixed git invocation whose subcommand is a query (`git.exe status`) shall remain allowed.
- **REQ-HGF-002** — **Where** the branch guard is enabled, **when** a command invokes the PowerShell call operator `&` on a quoted call target whose content names git (`& 'git' switch probe`, `& "git.exe" switch probe`) followed by text matching a branch-state pattern, the pre-tool hook shall deny under the same conditions as REQ-HGF-001, and the quoted-argument collapse shall keep protecting quoted spans at every other site (a quoted `git switch` carried as data inside another command's argument shall stay allowed).
- **REQ-HGF-003** — **Where** the branch guard is enabled, **when** a git subcommand token in command position is split by a PowerShell backtick escape (the measured form is ``git swi`tch probe``), the pre-tool hook shall reach the same decision the de-escaped command reaches, and a backtick inside a single-quoted string shall keep the quoted-span (literal data) treatment.
- **REQ-HGF-004** — **Where** the branch guard is enabled, **when** a `pwsh` or `powershell` invocation carries a `-Command`-family parameter (`-Command`, `-c`, and their prefix abbreviations), the pre-tool hook shall apply the branch-state scan to the parameter's payload with the existing pipeline computed within the payload (quoted-span collapse and comment elision included), so `pwsh -Command "git switch probe"` denies and `pwsh -Command "git status"` stays allowed, and the control form `cmd /c git switch probe` shall keep its existing deny.

### B.2 Unclassifiable forms (F3, aliases)

- **REQ-HGF-005** — **When** a PowerShell command that names `git` in its command text resolves its executable dynamically through a parenthesized subexpression in call position (the measured form is `& (Get-Command git) switch probe`), the branch guard shall allow the call and append exactly one unclassifiable audit line whose reason contains the literal word `unclassifiable` — the resolved executable is decided at runtime and a deny would fire without positive evidence, mirroring the encoded-command treatment of REQ-HMP-010 — and a call-position parenthesized subexpression form carrying no `git` word in the command text (e.g. `& (Get-Command node) serve`) shall earn no audit line (REQ-HGF-008).
- **REQ-HGF-006** — **When** a PowerShell command names `Start-Process` through a documented alias (`saps`, `start`) while `git` is named as the program or inside its argument list, the unclassifiable policy shall produce exactly the same audit line (event `powershell-unclassified`, construct `start-process`) that the fully spelled form already produces.

### B.3 Deny-list literal payloads (부수)

- **REQ-HGF-007** — **When** a deny-list pattern (the compiled set that denies bare `terraform destroy`) matches a literal quoted operand of an indirection construct — Bash `eval "…"`, PowerShell `iex "…"` / `Invoke-Expression "…"`, or the joined program-and-argument text of `Start-Process … -ArgumentList …` — the pre-tool hook shall deny the call with the same deny reason as the bare form, and an operand that is not a literal (variable reference, subexpression, command substitution) shall keep the existing fail-open path with the audit line REQ-HMP-010 already produces where applicable.
- **REQ-HGF-008** — The detection added by REQ-HGF-001 through REQ-HGF-004 and REQ-HGF-007 shall not change the decision of, and shall not append an audit line to, any command that names no branch-state pattern in a scanned position and carries no deny-list pattern in any scanned position; the healthy battery enumerated in acceptance.md §D is the fixed control set.

### B.4 Documentation and measurement (F4, F5)

- **REQ-HGF-009** — The comments governing the integration-lock audit log (the `integrationLockAuditRelPath` declaration and any adjacent comment describing the log's scope) shall state the measured scope — every unclassifiable PowerShell command observed while a live foreign hold exists is recorded, whether or not it is merge-shaped (measured forms: `iex "git status"`, `Start-Process git -ArgumentList 'log'`, both logged and both allowed) — and no comment shall claim a narrower scope than that behavior. Doc-only; no behavior change.
- **REQ-HGF-010** — **Where** pwsh is available outside any worktree, the run phase shall measure the set of `-EncodedCommand` parameter spellings pwsh accepts — including `/`-prefixed and Unicode-dash forms (U+2013 `–enc`) — over the candidate list of acceptance.md §E, under declared turn and wall-clock caps, and shall record the accepted set in progress.md §E.2.
- **REQ-HGF-011** — The encoded-command detector shall accept every spelling the M0 measurement shows pwsh accepting, either directly or through a documented superset whose rationale cites the over-match policy (an over-match costs one audit line and never a deny), and spellings the measurement shows pwsh rejecting shall not be required of the detector.

### B.5 Regression guards

- **REQ-HGF-012** — The delivery shall not modify any file under `internal/template/` or `.claude/`; the `PowerShell` PreToolUse matcher registration delivered by SPEC-HOOK-MATCHER-POWERSHELL-001 (`settings.json.tmpl` line 69) and the `PowerShell(...)` permission deny rules remain byte-unchanged. Template-surface assessment (recorded per the card's 로컬·템플릿 동시 requirement): this SPEC's changes are Go source and tests only — the matcher registration that routes PowerShell calls to the guard already exists, so no template change and no template rebuild step are required; acceptance.md AC-HGF-014 makes the assessment verifiable at close.
- **REQ-HGF-013** — The audit lines added for the new forms shall preserve the measured one-line injection resistance: a command containing a newline shall produce exactly one audit line with the command recorded verbatim inside its quoted field.
- **REQ-HGF-014** — Every hook test this SPEC adds or modifies shall set `MOAI_HOME` to a per-test temporary directory, shall not call `t.Parallel`, and shall not read or write the real MoAI home, its lease database, or the repository's `.moai/state`.

## §C Constraints and collision points

- **Base and absorption.** Authored on `19b5321c1` (branch `WT-powershell-guard-debt`, clean). Run-phase M1 absorbs local `develop` first and re-measures every line number this SPEC cites (`powershell_indirection.go:29-31`, `:75` token switch area, `:109`/`:135` helpers; `integration_lock_guard.go:109-112`; `pre_tool.go:1024` deny-list site) before editing — the parent SPEC's §C pattern.
- **Development mode**: `tdd` (quality.yaml) — RED-GREEN-REFACTOR. The repro tests of M1 are the RED layer and precede every detector change; plan.md §E.8 carries the verbatim RED outputs.
- **Implementation Kickoff Approval**: satisfied by the operator's standing 자율(autonomous) policy for this run — lead dispatch 2026-09-26, card t1255, factory lane worker-65. The gate is **recorded here, not skipped**; the lane proceeds under the operator's standing delegation.
- **Repro-first evidence discipline**: every acceptance criterion adopts the two-cell pattern (RED-now with command / verbatim stdout / exit code / tree SHA, plus a green-path cell) per `.claude/rules/moai/development/verification-completeness.md` §2; the evidence carrier is the ledger in acceptance.md §B.
- **Test environment**: env-scrubbed scoped runs only — `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/ -run '<new tests>'`. No full-suite local run; no background load; `pwsh` is never invoked inside a worktree (the worktree session guard refuses it) — the F5 measurement runs in a scratch directory outside any worktree.
- **Regions this SPEC edits**: `branch_guard.go` (pattern/token handling for F1a/F1b/F1c and the F2 payload scan entry), `powershell_indirection.go` (F3 construct, `saps`/`start` tokens, F4 comment, F5 conformance), `pre_tool.go` (deny-list application over extracted literal operands for REQ-HGF-007), plus new test file(s) in `internal/hook/`. No template or `.claude/` file (REQ-HGF-012).
- **Known helpers for the new tests** (all present and compiled by the probe run at `19b5321c1`): `hmpHandler`, `hmpCfg`, `hmpInput`, `hmpAuditLines`, `hmpSeedForeignLock` (`hmp_powershell_guard_test.go` family) and `newBranchGuardRepoFixture` (branch-guard test family). New tests follow the `branch_guard_{flagclass,comment,quoted,heredoc}_test.go` family shape.

### C.4 Separability of the 부수 subset

M4 (REQ-HGF-007/008 + AC-HGF-007/008/009) is a self-contained subset: if the operator splits it to a follow-up card, M0-M3 and M5 close this SPEC with the deny-list forms demoted to a recorded open item and the affected ACs marked deferred — the rest of the matrix does not depend on it. The split is an operator decision recorded in progress.md, never a lane-side scope cut.

## §D Out of Scope

### Out of Scope — t1224 debts F6, F7, F8

- F6 (case-C live evidence details not persisted to files), F7 (`pre_tool.go` ANCHOR wording and the `LogBashEvidence`-family naming), and F8 (the source guard sees string literals only) belong to other cards and are not absorbed here.

### Out of Scope — Windows live verification

- Measurement host is macOS; Windows-only behavior beyond `GOOS=windows go build` exit 0 remains inference, exactly as in the parent SPEC.

### Out of Scope — Full PowerShell parsing and payload decoding

- No PowerShell parser is added and encoded payloads are never decoded; detection stays structural token analysis over the command text.

### Out of Scope — Non-literal indirection operands on the deny path

- Variable, subexpression, and command-substitution operands stay fail-open (allow, plus the existing D2 audit line where applicable). Only literal quoted operands join the deny scan of REQ-HGF-007.

### Out of Scope — D2 policy for unclassifiable `-Command` payloads

- Whether a non-literal `-Command` payload (e.g. `pwsh -Command "$c"`) should also earn an audit line is left as-is; this SPEC adds no audit lines for that form. The deny gap is closed by payload scanning; the audit-volume question is separate.

### Out of Scope — Slot lease and PostToolUse surfaces

- Slot-lease patterns keep their operator-configured behavior (parent SPEC §A.3 site 4); the PostToolUse matcher and the `PowerShell(...)`/`Bash(...)` permission rules are untouched.

## §E Verification overview and cross-references

- 14 requirements (B.1-B.5), 14 acceptance criteria in `acceptance.md` (AC-HGF-001..014; 12 release-blocking two-cell, 2 regression-guard). Mutant matrix (both-way: legitimate-path no-change + disguise detection) is acceptance.md §D — mandatory per the card for every detection this SPEC extends.
- plan-auditor focus questions are listed in plan.md §H.
- Cross-references: SPEC-HOOK-MATCHER-POWERSHELL-001 (parent — D2 policy, matcher registration, REQ-HMP-010); SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001 (query-vs-mutate verdict family the expanded forms reuse); SPEC-GUARD-COMMENT-SCAN-001 (data-vs-command collapse family); SPEC-POWERSHELL-DENY-PARITY-001 (t1211 — the Claude Code **permission-rule** namespace, `PowerShell(...)` deny rules; a different layer from this SPEC's hook-side Go deny list — no scope overlap); SPEC-V3R6-TOOL-POLICY-SSOT-001 (the tool-policy SSOT — REQ-HGF-007 reuses the hook-side deny-list pattern set whose provenance traces through the parent SPEC's §A.3 classification, not the permission namespace); `.claude/rules/moai/development/verification-completeness.md` (two-cell adoption); `.moai/reports/t1224/verdict.md` (baseline evidence).
