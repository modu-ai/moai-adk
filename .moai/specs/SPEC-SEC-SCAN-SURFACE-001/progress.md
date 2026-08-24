# Progress — SPEC-SEC-SCAN-SURFACE-001

Card: t217 · Branch: `security-scan-surface` · Worktree: `.claude/worktrees/t217`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (spec.md + plan.md + acceptance.md + progress.md). 16 REQ / 16 AC — at the Tier M
  ceiling, verified by count.
- Class: C — a design change spanning `internal/hook`, `internal/cli`, and the distributed
  template; carries a decision (`spec.md` §C) that a reviewer may disagree with.
- Evidence base: `.moai/reports/t217/investigation.md`, measured on tree `c4e90cd58`.
  Authoring-time measurements are tabulated in `spec.md` §A.1.
- Two card premises were found inaccurate and corrected in `spec.md` §A.2 rather than inherited.
- Status: `draft`. Awaiting re-audit (iteration 2).

### Audit iteration 1 remediation (FAIL 0.65 → resubmitted)

Verdict read from `.moai/reports/t217/plan-audit.md`. All six blocking findings were
independently re-verified before editing; none was auditor error.

| Finding | Disposition |
|---|---|
| D1 `tags` sequence ⇒ frontmatter ParseFailure | Fixed — `tags:` is now a quoted comma-separated string per the schema SSOT (`spec-frontmatter-schema.md` §Field Reference, `tags` = string). Lint clean, output below |
| D2 pre-filter may be a no-op | Resolved by measurement; exit **(a)** taken — see below |
| D3 AC-SSS-004 vacuous | Fixed — the criterion now counts scanner-side (PASS 0, today 1) *and* caller-side (PASS 1, today 0) resolutions; the two invert, so it cannot pass on the untouched tree |
| D4 drift command fails today | Fixed — `[ -f "$b" ]` guard restored (measured: guarded 0 lines, unguarded 31), and the pair check re-scoped to files this SPEC's diff changes. Pre-existing `handle-pre-tool.sh` comment drift recorded as out of scope |
| D5 measurement seam does not exist | Fixed — M0 step 1 now owns creating it. Also corrects the auditor's proposed remedy: a handler-level stub is *not* possible today either, because `preToolHandler.scanner` is a concrete `*security.SecurityScanner` (`pre_tool.go:325`) |
| D6 20 REQ > Tier M ceiling 16 | Fixed by consolidation to exactly 16, no content dropped (verified: `grep -c "^- \*\*REQ-SSS-"` → 16) |
| D7 short-circuit order dependency | Adopted — reachability clause added to REQ-SSS-012, evidence recorded in `spec.md` §A.3, asserted by AC-SSS-012 |
| D8 empty derived language set | Adopted — folded into REQ-SSS-006, asserted by AC-SSS-007 |

**D2 exit taken: (a) — specify the extractor, keep A1.**

Measurement (`python3 .moai/reports/t217/skiprate.py .`, worktree `c4e90cd58`):

```
go files=2438 wouldSKIP=22 rate=0.9%
js files=81  wouldSKIP=78 rate=96.3%
py files=14  wouldSKIP=12 rate=85.7%
```

(Figures corrected in iteration 2 — see below. The iteration-1 figures were go 1.2% / py 92.9%,
inflated by an incomplete token set.)

The auditor's reading was correct that the shipped `sec-hardcoded-credential` is `kind:` +
`regex:` in all four covered languages, and that v0.1.0's §C.2 would have classified it
underivable — making A1 a no-op everywhere. §C.2 now carries a full extraction table with two
added rows: regex top-level alternation (union of per-branch mandatory prefixes, sound by the
same disjunction logic as `any:`) and `kind:` + `regex:` (derive from the regex conjunct, sound
because `kind` narrows and never widens). Under that table the rule is derivable with tokens
`{sk-, AKIA, ghp_, xox, AIza}`, and the measured skip rate is 96.3% / 85.7% for js-ts / python
against 0.9% for Go. A1 is kept because the saving is decisively non-trivial for three of the
four covered languages, and cutting it on the strength of the 0.9% Go figure would be deciding a
16-language tool's behaviour from the fact that this repository happens to be written in Go
(`CLAUDE.local.md` §15). The Go degeneracy and the measurement's gaps are now stated in §C.3
rather than asserted away.

**D1 closure evidence** — `~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md`:

```
✓ No findings — all SPEC documents are valid
rc=0
```

(Before the fix, the same command returned
`ERROR ParseFailure ... line 13: cannot unmarshal !!seq into string`, `rc=1`.)

### Audit iteration 2 remediation (PASS 0.925 — two blocking findings closed)

Verdict read from `.moai/reports/t217/plan-audit-2.md`. Both findings verified independently
before editing.

**E1 — the measurement script did not implement the rule it claimed to. Confirmed and fixed.**
`sec-command-injection-shell` (python, error) is a **four**-branch `any:` — `subprocess.call`,
`subprocess.run`, `subprocess.Popen`, `os.system` — and the script listed two. §C.2's `any:` row
requires the full union, so the omission inflated the reported saving. Root cause of my own
error: the branches were read through `grep -B4 -A8 "severity: error"`, whose 8-line window cut
the tail of the block — a bounded-output read presented as a complete one.

Auditing the other two token sets while fixing it found a **second** incomplete set the audit did
not flag: **go** carried only `,` and `=` (the first rule's tokens) on the argument that they
bound the rate. The argument is sound — a union can only lower a skip rate — but it made the
figure an upper bound, not a measurement. Both are now enumerated rule-by-rule.

- **js / ts — audited, already complete.** The `any:` has exactly two branches
  (`child_process.exec`, `cp.exec`) and both were present; the credential prefixes were present.
  Rate unchanged at 96.3%.
- **python — 2 tokens added** (`subprocess.run`, `subprocess.Popen`). 92.9% → **85.7%**, matching
  the auditor's corrected figure exactly.
- **go — 7 tokens added** (`for`, `defer`, `const`, `SignedString`, `exec.Command`,
  `template.HTML`, `md5.New`, plus the credential prefixes). 1.2% → **0.9%**.

Both corrections move in the deflating direction. The conclusion is unchanged: 0.9% for Go
against 85.7-96.3% elsewhere still carries the exit-(a) decision.

**E2 — soundness claim made honest.** §C.2's underivable row now names two further regex forms:
an inline flag (`(?i)`, `(?s)`, `(?m)`) and an optional or quantified leading literal. Both are
absent from the shipped ruleset, so the row is a soundness closure rather than machinery — the
row states that explicitly so no one builds an extractor for a case that does not exist.

**Optional items, all three taken.** AC-SSS-016's pair check now runs
`driftcheck.sh pairaxis` (the deployed ↔ template axis REQ-SSS-015 actually speaks to) as the
primary check, with the templates-internal `guarded` mode retained as a secondary signal;
AC-SSS-015 cites `~/go/bin/moai` instead of the untracked `./bin/moai` (verified: `exit=0`); and
§C.3's saving claim is softened to describe this sample rather than project a rate.

**Open gap left open, marked.** The `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/`
row in AC-SSS-016 is now labelled **UNVERIFIED** rather than "pass" — the run exceeded 120 s with
empty output, so its value is a Gap, not a Claim (`verification-claim-integrity.md` §2). M4 must
measure it with an explicit longer timeout before citing it.

## §E.2 Run-phase Evidence

### M1 — A2 + A3: resolve once, skip when there is nothing to find

Baseline attribution for every row below: measured in this run, against this tree, with the
pre-M1 tree at `053fb1f25` on branch `WT-security-scan-surface`. Redirected output is under
`.moai/state/verify/t217/`.

**RED evidence (captured before any implementation, `.moai/state/verify/t217/red.txt`)** — the
new criteria could not compile against the untouched tree, which is what makes the M1 seams
falsifiably test-first:

```
internal/hook/security/coverage_test.go:88:27: NewRuleManager().ResolveCoverage undefined (type RuleManager has no field or method ResolveCoverage)
internal/hook/pre_tool_scan_config_test.go:139:75: undefined: security.LanguageCoverage
internal/hook/pre_tool_scan_config_test.go:164:75: unknown field Rules in struct literal of type security.ScannerConfig
internal/hook/pre_tool_scan_config_test.go:165:41: unknown field rules in struct literal of type preToolHandler
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
FAIL	github.com/modu-ai/moai-adk/internal/hook/security [build failed]
```

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-SSS-001 | PASS | `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v` | `--- PASS: TestScanWriteContentDifferential (0.18s)` — ran, not skipped: all five denying fixtures still deny (`error_count=1` for `sample.go`, `digest.go`, `run.js`, `run.ts`, `run.py`), the warning-only fixture still allows. Assertion (ii) does not compile before M2 and is not claimed. Test file unmodified — absent from `git status --short` |
| AC-SSS-002 | PASS | `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoScan -count=1 -v` | `--- PASS: TestScanWriteContentNoConfigNoScan (0.00s)` — fake records 0 `ScanFile` calls, decision allow |
| AC-SSS-003 | PASS | `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoTempFile -count=1 -v` | `--- PASS` both arms: `no_config_creates_no_temp_file` (no new `moai-security-scan-*`) and the control `resolvable_config_creates_exactly_one_temp_file` (exactly 1, snapshotted **during** the call from inside the fake, since the deferred cleanup removes it before return) |
| AC-SSS-004 | PASS | `go test ./internal/hook/ -run TestConfigResolvedByCallerNotScanner -count=1 -v` | `--- PASS: TestConfigResolvedByCallerNotScanner (0.04s)` — scanner-side `FindRulesConfig` counter 0, caller-side resolution counter 1. Both halves counted; the counters invert against the pre-implementation tree (scanner-side 1, caller-side 0). Corroborating grep of the non-test tree: the only remaining `FindRulesConfig` call sites are `coverage.go:74` (the caller-side resolver) and `scanner.go:123`, which is inside `ScanFiles` (**plural**) — not the pre-write path. `ScanFile` performs zero resolutions |
| AC-SSS-005 | PASS | `go test ./internal/hook/ -run TestScanWriteContentUncoveredLanguage -count=1 -v` | `--- PASS` for all 11 uncovered extensions (`.rs .java .kt .c .cpp .rb .php .swift .cs .ex .scala`), 0 `ScanFile` calls each, plus `control_sample.go` recording 1 |
| AC-SSS-006 | PASS | `go test ./internal/hook/security/ -run TestCoveredLanguagesFollowConfig -count=1 -v` **and** `go test ./internal/hook/ -run TestScanWriteContentCoveredLanguageFollowsConfig -count=1 -v` | `--- PASS` both. Derivation half: the shipped ruleset covers `go`; a copy whose `ruleDirs` names only a python-rule directory does not. Scan-count half: 1 call vs 0 across the same two arms — the split no hardcoded language list can produce. See the note below on the criterion's cited package |
| AC-SSS-007 | PASS | `go test ./internal/hook/security/ -run TestUnreadableOrEmptyConfigEscalates -count=1 -v` **and** `go test ./internal/hook/ -run TestScanWriteContentUnreadableConfigEscalates -count=1 -v` | `--- PASS` all three arms in both (malformed YAML / missing `ruleDir` / rule files declaring no `language:`): derivation reports UNKNOWN, and the gate dispatches **1** `ScanFile` call in every case. Behaviour-preservation criterion — PASS value equals today's value by design |

**Note on AC-SSS-006 / AC-SSS-007's cited command.** Both criteria name
`go test ./internal/hook/security/`, but their stated instrument (the counting fake + a
`scanWriteContent` call) lives in `internal/hook`; the two cannot be satisfied by one package.
Each is therefore closed by a pair: the derivation assertion in the package that owns it, and the
scan-count assertion in the package that owns the gate. Both commands above run and pass. This is
a criterion-authoring inconsistency, recorded rather than silently resolved.

**Deliverables.**

- `internal/hook/security/coverage.go` (new) — `LanguageCoverage` + `ruleManager.ResolveCoverage`.
  Three distinguishable states: no config resolved / resolved-but-unknown / resolved-and-known.
  Every failure path (unreadable file, unparseable YAML, missing or unwalkable `ruleDir`,
  undecodable rule document, empty derived set) returns `Known=false`, so an unknown derivation
  escalates and never skips. Derivation is cached per `RuleManager` for its lifetime — one load
  per hook invocation, no invalidation machinery.
- `internal/hook/security/types.go` — `ResolveCoverage` added to the `RuleManager` interface.
- `internal/hook/security/scanner.go` — `ScanFile`'s third parameter is now an already-resolved
  **config path**, not a project dir, and the scanner-side resolution on that path is removed;
  `ScannerConfig.Rules` added as the seam AC-SSS-004 injects its counting manager through.
- `internal/hook/pre_tool.go` — `scanWriteContent` reordered to extension check → config
  resolution → covered-language check → (empty M2 pre-filter slot) → temp file → scan, returning
  `("", "")` before the temp file on every skip; `preToolHandler` gained a lazily-constructed
  `rules` field.
- `internal/hook/security/scanner_test.go` — the four `ScanFile` call sites updated to the new
  config-path contract (they passed a project dir with no config, so they now pass `""`).
- New tests: `internal/hook/pre_tool_scan_config_test.go`,
  `internal/hook/security/coverage_test.go`.

**Cross-cutting verification.**

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Vet | `go vet ./internal/hook/...` | exit 0, no output |
| Tests (hook) | `go test ./internal/hook/... -count=1 -timeout 600s` | all 11 packages `ok` |
| Tests (cli, per plan §D) | `go test ./internal/cli/... -count=1 -timeout 900s` | exit 0, all 17 packages `ok` (`internal/cli` 466 s) |
| Lint | `golangci-lint run ./internal/hook/... --timeout=5m` | `0 issues.` |
| Coverage `internal/hook` | `go test ./internal/hook/ -count=1 -cover` | **84.8%**, against a measured pre-change baseline of **84.5%** (baseline taken by extracting `git archive HEAD` to a scratch tree and running the same command there). Below the 85% target, but the shortfall is pre-existing and M1 moved it **up** 0.3 pp |
| Coverage `internal/hook/security` | `go test ./internal/hook/security/ -count=1 -cover` | **90.3%**, baseline **91.0%** — a 0.7 pp dip from `coverage.go`'s untaken error branches; still well above the 85% target |
| Subagent boundary | `grep -rn 'AskUserQuestion' internal/hook internal/hook/security` (non-test, non-comment) | no matches |

**Scope.** M2 (the literal pre-filter), M3 (the PostToolUse guardian merge), and M4 (template
mirroring / PR) are untouched. The pre-filter exists only as a commented ordering slot in
`scanWriteContent`. No `.claude/` file and no template file is in this diff.

**Gaps (explicitly NOT observed).** The full repository suite was not run locally — CI on the pull
request is the verdict for it (`plan.md` §D). AC-SSS-008 through AC-SSS-016 belong to M2-M4 and
are not claimed. AC-SSS-001's assertion (ii) does not compile before M2 and is not claimed. No
latency figure is claimed anywhere.

**Residual risk.** The covered-language derivation reads the configuration's own `ruleDirs`; a
project relying on ast-grep's implicit rule discovery outside `ruleDirs` would derive a narrower
set than `sg` actually applies, and a language could be skipped that `sg` would have scanned. The
shipped ruleset declares its directories explicitly, and the differential corpus covers all four
of its languages, so the risk is bounded to configurations unlike the shipped one. The
per-manager cache assumes the hook process is short-lived; a long-lived process editing its own
rules mid-run would keep the stale set.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
