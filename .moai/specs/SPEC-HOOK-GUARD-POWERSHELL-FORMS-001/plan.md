---
id: SPEC-HOOK-GUARD-POWERSHELL-FORMS-001
title: "Plan — Classify the PowerShell tool-call forms the shell guards pass without record (t1224 debts F1-F5)"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
tier: M
---

# plan.md — SPEC-HOOK-GUARD-POWERSHELL-FORMS-001

## §A Context

- **Tree / branch / HEAD at authoring**: worktree `.claude/worktrees/t1255`, branch `WT-powershell-guard-debt`, HEAD `19b5321c1` (card base = local develop merge `19b5321c1`, clean). Lane commits with the card id `t1255` in every commit message; the SPEC agent does not commit.
- **SPEC artifacts**: `.moai/specs/SPEC-HOOK-GUARD-POWERSHELL-FORMS-001/{spec,plan,acceptance,progress}.md` (Tier M + progress).
- **Parent SPEC**: SPEC-HOOK-MATCHER-POWERSHELL-001, `status: completed`, landed via merge `0119a8f2c`. Its verdict (`.moai/reports/t1224/verdict.md`) § Residual-risk names F1-F3 and the alias/quoted-target forms; § 후속 카드용 재현 (lines 66-106) carries the probe recipe and the `-EncodedCommand` measurement recipe. Probe copy: `.moai/reports/t1224/followup-probe_test.go.txt` (sha256 `868c34db…`, local-only file — the acceptance ledger in-tree carries the deciding rows).
- **Evidence already measured on this tree** (2026-09-26, exit 0): the overlay probe re-run at `19b5321c1` — all F1/F2/F3/alias forms `allow, unclassifiedLines=0`; control `cmd /c` denies; D2 `iex` rows log exactly one line; DENYLIST rows confirm the 부수; ILOCK rows confirm F4. Verbatim rows: acceptance.md §B ledger.
- **Implementation Kickoff Approval**: satisfied by the operator's standing 자율(autonomous) policy for this run (lead dispatch 2026-09-26, card t1255). Recorded, not skipped — this line is the gate record.
- **Existing infrastructure**: guard pipeline in `branch_guard.go` (`substituteQuotedArguments` → `substituteHeredocBodies` → `substituteShellComments` → `branchStatePatterns`), D2 policy in `powershell_indirection.go` (`powerShellIndirection`, `appendUnclassifiedAudit`, `isEncodedCommandParameter`), deny-list application in `pre_tool.go` (~`:1024`), integration-lock unclassified append at `integration_lock_guard.go:109-112`. EXTEND these; do not restructure.

## §B Known Issues (relevant subset of the delegation template)

- **B1 Cross-platform**: the hook runs on Windows too; `GOOS=windows GOARCH=amd64 go build ./...` must stay exit 0. The F5 detector work touches byte-level prefix matching — U+2013 (`–`) is a multi-byte rune; any spelling normalization must be explicit about bytes vs runes.
- **B7 CWD resolution**: tests use per-test fixtures (`newBranchGuardRepoFixture`, `t.TempDir`) — never the repository's own `.moai/` (REQ-HGF-014).
- **B8 Working-tree hygiene**: runtime-managed paths (`.moai/state/`, `.moai/logs/`, `.moai/harness/`) are never written by tests or code changes.
- **B11 Blocker protocol**: any missing input (e.g. pwsh unavailable for M0) returns a structured blocker report / INCONCLUSIVE record — never a question in prose.

## §C Pre-flight (run before M1, after absorbing local develop)

```bash
# 1. branch + baseline
git branch --show-current && git rev-parse --short HEAD
# 2. cross-platform build feasibility
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
# 3. lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m internal/hook/... 2>&1 | tail -5
# 4. probe integrity (hash-verified evidence)
shasum -a 256 .moai/reports/t1224/followup-probe_test.go.txt
# 5. line-number re-measure after absorption
grep -n "integrationLockAuditRelPath is" internal/hook/powershell_indirection.go
grep -n "Dangerous command blocked" internal/hook/pre_tool.go | head -2
# 6. M0 availability probe (outside worktree!)
which pwsh || echo "pwsh absent — M0 records INCONCLUSIVE per REQ-HGF-010 Where gate"
# 7. referenced-SPEC status scan (every referenced SPEC must be non-superseded / non-archived / non-rejected)
for s in SPEC-HOOK-MATCHER-POWERSHELL-001 SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001 SPEC-GUARD-COMMENT-SCAN-001 SPEC-POWERSHELL-DENY-PARITY-001 SPEC-V3R6-TOOL-POLICY-SSOT-001; do
  printf '%s: ' "$s"; grep -m1 '^status:' ".moai/specs/$s/spec.md"
done
```

Absorption duty: `git merge develop` (local) before M1; re-measure every line number spec.md §C cites and update the SPEC citations if moved (report, don't silently renumber).

## §D Constraints

**PRESERVE (never modify)**:
- `internal/template/**` and `.claude/**` — REQ-HGF-012; the template-surface assessment is "Go code only" and AC-HGF-014 verifies it at close.
- The existing quoted-argument / heredoc / comment collapse pipeline semantics — the false-positive protections they encode (quoted prose, heredoc bodies, comments) are settled behavior from SPEC-GUARD-COMMENT-SCAN-001 and the branch-guard family.
- The D2 audit-line contract: event `powershell-unclassified`, reason carries `unclassifiable`, one line per call, newline-safe (REQ-HGF-013).
- `cmd /c` control parity: the existing deny stays byte-identical.
- Other SPEC directories, runtime-managed `.moai/` paths, CHANGELOG.md.

**Forbidden**:
- `go test ./...` (full suite locally), background load, `pwsh` inside any worktree, `--no-verify`, `--amend`, force-push.
- Decoding `-EncodedCommand` payloads; denying F3 dynamic-resolution forms; scanning non-literal operands on the deny path.

**Required**: Conventional Commits with card id (`fix(hook): … (t1255)` style), `🗿 MoAI` trailer per harness attribution; env-scrubbed scoped test runs.

## §E Self-Verification deliverables (run-phase)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §E, with the attribution triple (command + verbatim output + `(this run, this tree)` + HEAD SHA) on every item:

- **E1** AC binary PASS/FAIL matrix — all 14 ACs of acceptance.md, citing the ledger row or the committed test output.
- **E2** Cross-platform build — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`, both exit 0.
- **E4** Subagent boundary grep — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` → 0 matches.
- **E5** Lint — `golangci-lint run internal/hook/...`; NEW issues named, baseline separated.
- **E8** RED outputs — verbatim failing output of the M1 repro tests captured BEFORE the detector changes (one block per form family), plus the pinned probe baseline of acceptance.md §B.

## §F Milestones (priority-ordered; decisions most likely to change come first)

### M0 (Priority High — measurement; feeds M5, blocks nothing else)
F5 spelling measurement, **outside any worktree** (scratch dir under `/tmp`; the worktree session guard refuses `pwsh`). Recipe: acceptance.md §E verbatim (B64 of `Write-Output ok` as UTF-16LE, loop over the candidate spellings including `/enc` and U+2013 `–enc`). **Caps declared up front** (card constraint + t1152 lesson): one loop of 11 candidate spellings plus the attached-colon form `-enc:<B64>` as a separate 12th invocation (acceptance.md §E recipe — recipe/cap/plan counts identical), wall-clock ≤ 10 min, ≤ 2 turns, foreground only, `timeout 60` per pwsh invocation, no background load. Deliverable: accepted-set table + detector comparison recorded in progress.md §E.2 (committed) and `.moai/reports/t1255/` (local detail). pwsh absent → INCONCLUSIVE record, detector unchanged (REQ-HGF-010 `Where` gate is explicit).

### M1 (Priority High — RED before any detector change)
Port the probe into committed tests: new `internal/hook/branch_guard_psforms_test.go` following the `branch_guard_flagclass_test.go` family shape, reusing the `hmp*` fixtures. One table-driven case per mutant-matrix row (acceptance.md §D), asserting the **post-fix** expectation. Run the scoped selector on the absorbed base and capture the verbatim RED output per form family into progress.md (plan.md §E8). No detector edits in this milestone.

### M2 (Priority High — F1a/F1b/F1c + F2 branch-state scan expansion)
Extend the branch-state scan for the `.exe` suffix (REQ-HGF-001), the call-operator + quoted-target form (REQ-HGF-002), the command-position backtick de-escape (REQ-HGF-003), and the `-Command` payload scan (REQ-HGF-004). The in-payload scan reuses the existing collapse pipeline computed **within** the payload. Both-way mutants green (legit legs of the matrix unchanged).

### M3 (Priority High — F3 demotion + alias coverage)
Add the call-position parenthesized-subexpression construct to `powerShellIndirection` (REQ-HGF-005 — allow + one line) and the `saps`/`start` tokens to the start-process construct set (REQ-HGF-006). Git-word gate unchanged — no `git`, no line.

### M4 (Priority Medium — separable 부수 subset; see spec.md §C.4)
Deny-list literal-operand scan (REQ-HGF-007): extract the literal quoted operand of Bash `eval`, PowerShell `iex`/`Invoke-Expression`, and the joined program+arguments of `Start-Process … -ArgumentList …`; run the existing compiled deny-list patterns over the extracted text; deny with the bare-form reason. Non-literal operands untouched. If the operator splits this subset to a follow-up card, record the split + deferred ACs in progress.md and close the rest.

### M5 (Priority Medium — closeout, mechanical)
F4 comment alignment in `powershell_indirection.go` (declaration comment + any adjacent scope comment — REQ-HGF-009, doc-only). F5 detector conformance per M0 (REQ-HGF-011 — extend `isEncodedCommandParameter` only if the measured accepted set demands it; superset kept only with the over-match rationale recorded at the declaration). Full AC matrix re-measure, env-scrubbed scoped package runs, GOOS=windows build, lint, template-diff-empty check (AC-HGF-014), progress.md §E.2/§E.3 population.

## §G Anti-patterns (detection-specific pitfalls)

- **Do NOT weaken `substituteQuotedArguments` globally** to catch `& 'git'` — the collapse exists because of the measured quoted-prose false positive (`moai todo add "… git switch …"`). The call-target inspection reads the raw span at the call-operator site only.
- **Do NOT raw-text-match the `-Command` payload.** The payload's own quoting is real: `pwsh -Command "Write-Output 'git switch'"` must stay allowed — the in-payload scan collapses quoted spans within the payload before matching. This row is in the mutant matrix and is the trap the F2 work is most likely to fall into.
- **Do NOT decode or scan encoded payloads**; the encoded-command construct stays unclassifiable-whatever-the-payload (parent D2).
- **Do NOT deny F3 forms.** `& (Get-Command git) switch` looks decisive but the subexpression's result is a runtime fact; a deny here breaks the fail-open norm and false-positives non-git dynamic invocations.
- **Do NOT let backtick normalization touch single-quoted spans** — there the backtick is literal data.
- **Do NOT emit more than one audit line per call**; the INJECTION probe row (newline command → one line, command verbatim in its quoted field) is the regression guard.
- **Do NOT drop the git-word gate when adding `start`** — the alias is a common English word; the gate is what bounds its over-match to one audit line.

## §H Cross-references and plan-auditor focus questions

- Parent SPEC: SPEC-HOOK-MATCHER-POWERSHELL-001 (D2 policy, REQ-HMP-010 wording, matcher registration). Verdict family reused: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001. Collapse family: SPEC-GUARD-COMMENT-SCAN-001.
- **Plan-auditor focus**:
  1. Is the per-form decision table (spec.md §A.3) sound — specifically F3 demotion vs expansion (is the runtime-resolution argument stated strongly enough), and F2 expansion carrying the in-payload-collapse constraint?
  2. Do the RED-now cells carry all four elements with the tree pinned to `19b5321c1` (not a branch name), and is any criterion vacuously green?
  3. Is the 부수 judgment (fix, both shells, literal-only, separable M4) within the card's "같은 카드에서 판정" mandate, and is the defer-alternative rejection argued?
  4. Does REQ-HGF-012 (template assessment) give the audit a mechanically verifiable no-template-change claim?
  5. Are the M0 caps concrete enough to bound the LIVE measurement (turn cap, wall-clock cap, per-invocation timeout)?
