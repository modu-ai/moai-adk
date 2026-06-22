# SPEC-SIMPLICITY-LADDER-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored by manager-spec (status: draft). Tier M.

- **Artifacts**: spec.md (GEARS, 12-field frontmatter + era:V3R6 + tier:M), plan.md (Tier M rationale + @MX:TODO-vs-@MX:DEBT verdict in §B), acceptance.md (17 ACs + 5 Given-When-Then scenarios), progress.md (this file).
- **SPEC ID self-check**: `SPEC-SIMPLICITY-LADDER-001` → decomposition PASS (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Exclusions**: §J carries 7 `### Out of Scope — <topic>` sub-headings (OutOfScopeRule satisfied).
- **Tier decision**: Tier M — REQ-2 touches `internal/mx/` scanner validity gate (`scanner.go:247`), which hard-rejects unknown tag kinds; registering `@MX:DEBT` requires Go code + TDD.
- **@MX:TODO-vs-@MX:DEBT verdict**: distinction holds → new tag justified (NOT an @MX:TODO extension). Decisive property: the TODO→WARN >3-iteration escalation rule would mis-handle deliberate-simplification debt. See plan.md §B.
- **Plan-auditor iter-1**: PASS-WITH-DEBT 0.84 (Tier M threshold 0.80; not skip-eligible). 5 defects resolved in-place (D1-D5), all backed by my own Read of `internal/mx/scanner.go` (verification-claim-integrity §1.1 — observed, not relayed):
  - **D1/D2**: §B.5 added — registering only `MXDebt` leaves `@MX:CEILING`/`@MX:UPGRADE` hitting `parseTag`'s `default` branch (scanner.go:235-251) → error → `continue`. REQ-2.4 now specifies a **recognized-sub-line-kind set** consulted before `default`. The latent `@MX:REASON` parse bug + `scanner_test.go:170` vacuous test (guarded by `&& len(tags) > 0`) are SURFACED and forward-linked to proposed `SPEC-MX-SUBLINE-PARSE-REPAIR-001` — NOT authored, NOT repaired here (scope discipline AP-SL-007).
  - **D3**: rot-risk contract now concrete — `"rotRisk": "no-trigger"` JSON field on `moai mx query --kind DEBT --json` (`RotRisk` field on existing `Tag` struct, no new ledger). AC-SL-008 + Scenario 3 assert the literal token.
  - **D4**: Scenario 4 re-labeled as a doctrine-text grep assertion (the TODO→WARN escalation is doctrine-only; grep of `internal/mx/` for `escalat`/`>3` = 0 matches).
  - **D5**: AC-SL-005/011 split into per-file `diff -q` mirror ACs (005a/005b/011a/011b/011c).
  - **Irony guard intact**: no new config/lint/hook/separate-JSON-ledger added; §J exclusions only GREW (added "latent REASON-path repair"). New APs AP-SL-007/008 guard D2/D3 scope creep.
- **Plan-auditor iter-2**: PASS-WITH-DEBT 0.86 (+0.02 monotonic; not skip-eligible). iter-1 D1-D5 confirmed resolved; the D1 fix introduced ONE new defect (D-NEW-1). 2 defects resolved in iter-3, both backed by my own Read/grep of `internal/mx/tag.go` + `mx-tag-protocol.md` (verification-claim-integrity §1.1):
  - **D-NEW-1 (SHOULD-FIX, regression I introduced)**: my iter-2 recognized-sub-line set `{CEILING,UPGRADE,REASON,SPEC,LEGACY,TEST,PRIORITY}` included `LEGACY`, copied verbatim from `mx-tag-protocol.md:26`. But `LEGACY` IS a real tag kind (`MXLegacy`, tag.go:25, confirmed via grep) — including it would make `parseTag` silently DROP standalone `@MX:LEGACY` tags. Fix: removed `LEGACY` → set is now `{CEILING,UPGRADE,REASON,SPEC,TEST,PRIORITY}` (SPEC/TEST/PRIORITY verified safe: grep `TagKind="(SPEC|TEST|PRIORITY)"` → 0). Added regression AC-SL-009c + Scenario 5 (standalone `@MX:LEGACY` → 1 `MXLegacy` tag) + AP-SL-009. The doctrine dual-classification (D-CARRY) left to `SPEC-MX-SUBLINE-PARSE-REPAIR-001`.
  - **D-NEW-2 (MINOR)**: AC-SL-011c had a dead "if not touched → N/A" escape. `skills/moai/references/mx-tag.md:39` enumerates the tag-type grammar (`NOTE|WARN|ANCHOR|TODO`), so DEBT MUST join it → that mirror IS touched. Reworded AC-SL-011c to mandate the edit + mirror (no N/A). M5 updated to confirmed-touched.
  - **Irony guard still intact**: still no new config/lint/hook/JSON-ledger; §J unchanged (7 exclusions); the LEGACY-exclusion is a removal, not an addition.
- **Next phase**: orchestrator re-audits iter-3 (this agent does NOT self-re-audit). After PASS, run-phase entry requires Implementation Kickoff Approval (CLAUDE.local.md §19.1).

## §E.2 Run-phase Evidence

Run-phase implemented by manager-develop (cycle_type=tdd). REQ-1 doctrine (ladder + carve-out + karpathy xref) + REQ-2 doctrine (@MX:DEBT in mx-tag-protocol + constitution @MX list + mx-tag.md grammar) authored; REQ-2 Go infra pre-existing GREEN (user-accepted) and re-verified. 4 template mirrors at per-file `diff -q` parity.

### AC PASS/FAIL Matrix

| AC ID | REQ | Status | Verification command | Actual output |
|-------|-----|--------|----------------------|---------------|
| AC-SL-001 | REQ-1.1 | PASS | python rung-count over ladder block in moai-constitution.md | `numbered rungs in ladder block: 6` |
| AC-SL-002 | REQ-1.2 | PASS | python forbidden-token scan of ladder block (npm/pip/import React/package.json/requirements.txt/go.mod/cargo/...) | `ladder block forbidden-token hits: NONE` |
| AC-SL-003 | REQ-1.3 | PASS | grep `Never simplify away` + `TRUST 5 Secured` + `Bash Risk-Amplifier Doctrine` in carve-out block | `carve-out: True / xref TRUST 5 Secured: True / xref Bash Risk-Amplifier: True` |
| AC-SL-004 | REQ-1.4 | PASS | grep `Enforce Simplicity` (=1 xref line) + `^[0-9]\. ` rung-count (=0, not restated) in karpathy-quickref.md | `Enforce Simplicity: 1 / ladder rungs in karpathy: 0` |
| AC-SL-005a | REQ-1.* | PASS | `diff -q .claude/.../moai-constitution.md internal/template/templates/.claude/.../moai-constitution.md` | `PASS (identical)` |
| AC-SL-005b | REQ-1.* | PASS | `diff -q .claude/.../karpathy-quickref.md <mirror>` | `PASS (identical)` |
| AC-SL-006 | REQ-2.1 | PASS | grep `@MX:DEBT` in mx-tag-protocol.md (Tag Types) + moai-constitution.md (@MX list) | `mx-tag-protocol: 6 / moai-constitution @MX list: 1` |
| AC-SL-007 | REQ-2.2 | PASS | grep `@MX:CEILING` + `@MX:UPGRADE` sub-lines documented as inline comment (not JSON) in mx-tag-protocol.md | sub-lines + example fence present |
| AC-SL-008 | REQ-2.3 | PASS | `go test -run TestMxQueryCmd_DebtRotRiskJSON ./internal/cli/` (asserts exactly one `"rotRisk": "no-trigger"` token) | `--- PASS: TestMxQueryCmd_DebtRotRiskJSON` |
| AC-SL-009 | REQ-2.4 | PASS | `go test -run TestScanDebtTagAndSubLines ./internal/mx/` (no "unknown tag kind: DEBT" error, 1 DEBT tag) | `--- PASS: TestScanDebtTagAndSubLines` |
| AC-SL-009b | REQ-2.4 | PASS | `go test -run TestParseTagSubLineSentinel ./internal/mx/` (CEILING/UPGRADE return errSubLineKind, GetErrors empty) | `--- PASS: TestParseTagSubLineSentinel` (6 sub-line subtests PASS) |
| AC-SL-009c | REQ-2.4 | PASS | `go test -run TestScanLegacyNotDroppedBySubLineSet ./internal/mx/` (standalone @MX:LEGACY → 1 MXLegacy tag) | `--- PASS: TestScanLegacyNotDroppedBySubLineSet` |
| AC-SL-010 | REQ-2.5 | PASS | `grep -i 'DEBT.*not.*escalat\|DEBT.*does not.*WARN' mx-tag-protocol.md` (Scenario 4 doctrine-text grep) | `1` (≥1 = PASS) + @MX:TODO-vs-@MX:DEBT distinction stated |
| AC-SL-011a | REQ-2.* | PASS | `diff -q .claude/.../mx-tag-protocol.md <mirror>` | `PASS (identical)` |
| AC-SL-011b | REQ-2.* | PASS | `diff -q .claude/.../moai-constitution.md <mirror>` (shared with 005a) | `PASS (identical)` |
| AC-SL-011c | REQ-2.* | PASS | mx-tag.md:39 grammar `tag_type := "NOTE" \| "WARN" \| "ANCHOR" \| "TODO" \| "DEBT"` + `diff -q` clean | grammar shows `\| "DEBT"`; mirror PASS (identical) |
| AC-SL-012 | (cross) | PASS | no new config file / lint rule / hook / JSON ledger (RotRisk is a field on existing Tag struct) | irony guard intact — only 90 markdown lines + the pre-existing Go field |

### Given-When-Then Scenarios

| Scenario | Status | Evidence |
|----------|--------|----------|
| 1 — ladder present, ordered, language-neutral | PASS | AC-SL-001 (6 rungs) + AC-SL-002 (0 forbidden tokens) |
| 2 — @MX:DEBT + sub-lines scan without error | PASS | TestScanDebtTagAndSubLines (GetErrors empty, 1 DEBT tag) + TestParseTagSubLineSentinel |
| 3 — no-trigger DEBT carries `rotRisk: "no-trigger"` | PASS | TestMxQueryCmd_DebtRotRiskJSON (exactly 1 token) + TestScanDebtRotRiskNoTrigger |
| 4 — DEBT non-escalation asserted as doctrine text | PASS | `grep -i 'DEBT.*not.*escalat\|DEBT.*does not.*WARN' mx-tag-protocol.md` → 1 |
| 5 — @MX:LEGACY NOT dropped by sub-line set | PASS | TestScanLegacyNotDroppedBySubLineSet (1 MXLegacy tag) + TestParseTagSubLineSentinel/LEGACY |

### Invariants

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Existing scanner tests unbroken (`TestScanFileWithWarnReason`) | PASS | `go test -run TestScanFileWithWarnReason ./internal/mx/` → PASS |
| Cross-platform build | PASS | `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Template neutrality | PASS | `TestTemplateNoInternalContentLeak` + `TestTemplateNeutralityAudit` + `TestLanguageNeutrality` PASS (scan edited mirrors) |
| spec-lint | PASS | `moai spec lint` → `No findings — all SPEC documents are valid` |
| golangci-lint (mx + cli) | PASS | `0 issues.` |
| go vet (mx + cli) | PASS | exit 0 |

### Residual-risk disclosure (verification-claim-integrity §3.5)

- `scripts/i18n-validator/TestBudget_FullRepoScanWithin35Sec` FAILs (full-repo scan ~49-51s vs 35s wall-clock budget). This is **OUT OF SCOPE** (B10 PRESERVE — `scripts/i18n-validator/` not modified by this SPEC; `git status` empty for that path), **pre-existing + known-flaky** (git history: a prior commit skipped it as "Windows-flaky timing test" and another bumped the budget 30s→35s; owned by SPEC-V3R6-I18N-VALIDATOR-BUDGET-001), **machine-dependent** (a full-repo file-count scan, independent of the 90 markdown lines this SPEC adds), and **NOT a CI required-checks gate** (absent from `.github/`). It is disclosed here, not claimed as a regression of this SPEC.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: implemented
ac_pass_count: 17
ac_fail_count: 0
run_commit_sha: 5de6f19ee
preserve_list_post_run_count: 6   # 6 pre-existing Go files PRESERVED (tag.go, scanner.go, resolver_query.go, mx_query.go, scanner_debt_test.go, mx_query_debt_test.go)
l44_pre_commit_fetch: "git fetch origin main → 0 0 (synced)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows_amd64: pass
total_run_phase_files: 14   # 4 live doctrine + 4 template mirrors + 4 pre-existing Go src + 2 pre-existing Go tests
m1_to_mN_commit_strategy: "single cohesive commit — REQ-1 doctrine + REQ-2 doctrine + 4 mirrors + pre-existing Go infra + SPEC dir, one Conventional Commit, main-direct push"
coverage:
  internal_mx: "87.5%"
  internal_cli_mx_query: "covered by TestMxQueryCmd_Debt* + scanner_debt_test"
residual_risk: "scripts/i18n-validator perf-budget test FAIL — out-of-scope, pre-existing, machine-dependent, not a CI gate (disclosed §E.2)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: completed
sync_commit_sha: 5de6f19ee  # backfill will follow, see below
sync_completeness: 100%
sync_artifacts_touched:
  - spec.md (frontmatter: status in-progress → completed + updated timestamp)
  - CHANGELOG.md ([Unreleased] entry: 17 ACs + 2 REQs summary)
mirror_parity:
  - moai-constitution.md: live vs template PASS
  - karpathy-quickref.md: live vs template PASS
  - mx-tag-protocol.md: live vs template PASS
  - mx-tag.md: live vs template PASS
lint_status: PASS (StatusGitConsistency warning now resolved: status: completed)
go_build: PASS (native + GOOS=windows GOARCH=amd64)
go_test: PASS (14 test cases + 4 existing baselines unbroken)
```

