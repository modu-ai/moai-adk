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

### M2 — A1: the literal pre-filter derived from the error-severity rules

Baseline attribution for every row below: measured in this run, against this tree, with the
pre-M2 tree at `dad1f5bb4` on branch `WT-security-scan-surface`. Redirected output is under
`.moai/state/verify/t217/`.

**RED evidence.** The extractor was written before its tests, so RED was demonstrated by removing
the implementation and running the M2 tests against the tree without it — the pre-implementation
state each acceptance criterion records as "does not compile. RED." This is a genuine measurement
of these tests against a tree lacking the unit, not a strict test-first ordering; it is recorded
as such rather than claimed as test-first.

`.moai/state/verify/t217/red-security.txt` — `go test ./internal/hook/security/ -run 'TestPrefilter' -count=1`:

```
internal/hook/security/prefilter_test.go:18:38: undefined: PrefilterSet
internal/hook/security/prefilter_test.go:25:9: undefined: DerivePrefilters
internal/hook/security/prefilter_test.go:83:15: undefined: deriveRuleTokens
internal/hook/security/prefilter_test.go:237:11: undefined: derivePrefilters
FAIL	github.com/modu-ai/moai-adk/internal/hook/security [build failed]
exit=1
```

`.moai/state/verify/t217/red-hook.txt` — `go test ./internal/hook/ -run 'TestScanWriteContentPrefilterSkip|TestScanWriteContentUnderivableEscalates|TestScanWriteContentDifferential' -count=1`:

```
internal/hook/pre_tool.go:697:14: undefined: security.DerivePrefilters
internal/hook/pre_tool_scan_differential_test.go:271:25: undefined: security.DerivePrefilters
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
exit=1
```

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-SSS-001 (assertion ii) | PASS | `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v` | `--- PASS: TestScanWriteContentDifferential (0.46s)` with `assertion (ii): 5 denying fixtures, none skippable by the pre-filter`. PASS (not SKIP) is what proves assertion (i) also ran — the corpus-validity gate `t.Skip`s instead of passing when no fixture denies. The recorded corpus rows, `wantDeny` values, and assertion (i) are unmodified; M2 appended assertion (ii) only |
| AC-SSS-008 | PASS | `go test ./internal/hook/security/ -run TestPrefilterDerivationSeverityScope -count=1 -v` | `--- PASS: TestPrefilterDerivationSeverityScope (0.05s)`, logging `non-error-only tokens excluded: 10`. Both halves asserted: no token contributed **only** by a non-error rule reaches the set (the exclusion set is computed from the ruleset, not hardcoded, so a token shared with an error rule is not falsely flagged), and every derivable error-severity rule contributes at least one token. The `10` is the count of genuinely warning-only tokens — without it the exclusion would observe nothing, so the test fails if it reaches 0 |
| AC-SSS-009 | PASS | `go test ./internal/hook/security/ -run TestPrefilterKindPlusRegexAlternation -count=1 -v` | `--- PASS: TestPrefilterKindPlusRegexAlternation (0.02s)` — `sec-hardcoded-credential` derives exactly `{AIza, AKIA, ghp_, sk-, xox}` in **all four** covered languages (go's `kind: interpreted_string_literal` + js/ts/python's `kind: string`), none marked underivable. The test also asserts all four documents were found, so a ruleset that stopped shipping one language fails rather than silently observing three |
| AC-SSS-010 (derivation) | PASS | `go test ./internal/hook/security/ -run TestPrefilterUnderivableShapes -count=1 -v` | `--- PASS` all 7 arms: the three the criterion names (`regex with no literal anchor`, `any with one tokenless branch`, `inside composite`) plus `kind alone`, `regex with an inline case-insensitive flag`, `regex whose leading literal is optional`, and a control proving a recognized shape in the same position IS derivable. Each arm additionally asserts `CanSkip` returns false for that language |
| AC-SSS-010 (end-to-end) | PASS | `go test ./internal/hook/ -run TestScanWriteContentUnderivableEscalates -count=1 -v` | `--- PASS` both arms: a javascript ruleset carrying one derivable rule **and** one `kind:`-only rule records **1** `ScanFile` call for a payload with no dangerous construct (fail-open escalation); the same payload against the same ruleset **minus** the underivable rule records **0**. The pair is what shows the 1 is the underivability escalating, not the language never being skippable |
| AC-SSS-011 | PASS | `go test ./internal/hook/ -run TestScanWriteContentPrefilterSkip -count=1 -v` | `--- PASS` both arms: a `.js` payload carrying none of the javascript tokens records **0** `ScanFile` calls and decision allow; the control — same language, same ruleset, a payload carrying `cp.exec` — records **1**. Pre-implementation measurement was 1 for the skip arm |

**Derived token sets (shipped ruleset, this tree).** Measured via `DerivePrefilters` against a
temp project root carrying `internal/template/templates/.moai/config/astgrep-rules/`:

```
go         derivable=true  ["AIza" "AKIA" "SignedString" "const" "exec.Command" "ghp_" "md5.New" "sk-" "template.HTML" "xox"]
python     derivable=true  ["AIza" "AKIA" "ghp_" "os.system" "sk-" "subprocess.Popen" "subprocess.call" "subprocess.run" "xox"]
javascript derivable=true  ["AIza" "AKIA" "child_process.exec" "cp.exec" "ghp_" "sk-" "xox"]
typescript derivable=true  ["AIza" "AKIA" "child_process.exec" "cp.exec" "ghp_" "sk-" "xox"]
```

Enumerated **rule by rule**, not sampled, reading each rule file whole rather than through a
bounded grep window: go carries 6 error rules (`sec-hardcoded-credential`, `sec-weak-hash-md5`,
`sec-command-injection-shell`, `sec-hardcoded-api-key`, `sec-hardcoded-jwt-signing-key`,
`sec-template-injection-html`); python 2 (`sec-hardcoded-credential`,
`sec-command-injection-shell` — whose `any:` has **four** branches, all four present above);
javascript and typescript 2 each (`sec-hardcoded-credential`, `sec-command-injection-exec` —
`any:` with **two** branches, both present). Every other rule in the shipped set is
`warning`-severity and contributes nothing.

**Divergence from `spec.md` §C.3's go enumeration (recorded, not silently absorbed).** §C.3 lists
go's `go-error-ignored-blank` (`,` `=`) and `go-defer-in-loop` (`for` `defer`) as error-severity
contributors. In this tree both rules are `severity: warning` (they were demoted in M0/M1 with the
reason recorded inline in `go/error-handling.yml` and `go/concurrency.yml`: error severity on the
pre-write gate refuses ordinary Go edits). So go carries 6 error rules here, not 8, and its token
set does not include those four tokens. The consequence is that go's measured skip rate would now
be **higher** than §C.3's 0.9%, not lower — the direction that makes A1 less degenerate for Go, not
more. No re-measurement of §C.3's rates is claimed.

**Deliverables.**

- `internal/hook/security/prefilter.go` (new) — `Prefilter` / `PrefilterSet` / `CanSkip`,
  `DerivePrefilters(configPath)` (I/O wrapper over the already-resolved config path the caller
  holds), and the pure `derivePrefilters(docs)` + `deriveRuleTokens(node)` extractors implementing
  `spec.md` §C.2 row for row: `pattern:` (longest identifier chain of the literal runs), `all:`
  (union — a conjunction, so any derivable conjunct's tokens are mandatory), `any:` (union, and
  underivable if **any** branch is tokenless), `regex:` with a top-level alternation (union of the
  per-branch mandatory literal prefixes), `regex:` without one (its mandatory literal prefix),
  `kind:` + `regex:` (tokens from the regex conjunct), and underivable for `kind:` alone,
  `inside:` / `has:` / `follows:` / any unrecognized key, an inline `(?i)`/`(?s)`/`(?m)` flag, an
  optional or quantified leading literal, and a pattern that is entirely metavariables.
- `internal/hook/pre_tool.go` — the M2 slot filled: `security.DerivePrefilters(coverage.ConfigPath).CanSkip(language, content)`
  after the covered-language check and before the temp file. The commented slot is gone.
- `internal/hook/security/prefilter_test.go` (new) — the four criterion tests plus a row-by-row
  extraction table over `regexTokens` / `patternToken` / the composites, a fail-open suite over the
  I/O paths, and a property arm asserting each admitted regex token really is a substring of a
  string the expression matches.
- `internal/hook/pre_tool_scan_prefilter_test.go` (new) — the two end-to-end criterion tests, each
  with a control arm.
- `internal/hook/pre_tool_scan_differential_test.go` — assertion (ii) appended. Corpus rows,
  `wantDeny` values, the corpus-validity gate, and assertion (i) unchanged.
- `internal/hook/pre_tool_scan_config_test.go` — **one constant changed**: `goPayloadContent` now
  carries `const`, a go pre-filter token. Three M1 controls assert "a covered language still
  scans"; once the pre-filter is live a token-free payload is skipped by the pre-filter rather than
  reaching the coverage check those controls observe, so they measured 0 instead of 1. The payload
  remains clean and no recorded decision changed. Observed before the fix:
  `TestScanWriteContentNoConfigNoTempFile/control` , `TestScanWriteContentUncoveredLanguage/control_sample.go`,
  `TestScanWriteContentCoveredLanguageFollowsConfig/shipped_ruleset_scans_go` — each
  `expected 1 ScanFile call, got 0`.

**Cross-cutting verification.**

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Vet | `go vet ./internal/hook/...` | exit 0, no output |
| Tests (hook tree) | `go test ./internal/hook/... -count=1 -timeout 900s` | all 11 packages `ok` (`internal/hook` 153s). A later re-run of the same batch saw one unrelated failure — see the flake row below |
| Tests (`internal/hook` re-run) | `go test ./internal/hook/ -count=1 -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/hook  121.013s`, exit 0 |
| Lint | `golangci-lint run ./internal/hook/... --timeout=5m` | `0 issues.` |
| gofmt (M2 files) | `gofmt -l` over the six touched files | only `pre_tool_scan_config_test.go`, which was **already** unformatted at `HEAD` (verified by running `gofmt -l` on `git show HEAD:` output). Not introduced here and not fixed here — out of M2 scope |
| Coverage `internal/hook/security` | `go test ./internal/hook/security/ -count=1 -cover` | **90.2%**, against M1's recorded 90.3% — a 0.1 pp dip. `prefilter.go`'s remaining uncovered statements are I/O and malformed-node error branches |
| Coverage `internal/hook` | `go test ./internal/hook/ -count=1 -cover` | **84.8%**, unchanged from M1's 84.8%. Below the 85% target; the shortfall is pre-existing |
| Subagent boundary | `grep -rn 'AskUserQuestion\|mcp__askuser'` over the three new/changed M2 source files | exit 1, no matches |

**Observed flake (unrelated to M2).** One run of the full hook-tree batch failed
`TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory` with
`Handle blocked 1.22377525s; expected return near the 250ms bound`. It is a wall-clock-bound
assertion in `session_start_parallel_test.go`, a file M2 does not touch, on a code path M2 does not
touch; it passed 3/3 in isolation (`-count=3`) and the `internal/hook` package passed clean on
re-run. Recorded as an observation, not repaired — repairing it would be out of scope.

**Scope.** M3 (the PostToolUse guardian merge) and M4 (template mirroring / PR) are untouched. No
`.claude/` file and no template file is in this diff; the shipped ruleset under
`internal/template/templates/.moai/config/astgrep-rules/` is **read** by the tests, never modified.

**Gaps (explicitly NOT observed).**

- `go test ./internal/cli/...` was NOT run. Plan §D names it alongside the hook suite; the M2
  dispatch scoped verification to `./internal/hook/...`. M2 adds one exported function and changes
  no existing signature, and `go build ./...` (which compiles `internal/cli`) is green — but the
  cli **tests** were not executed in this run. CI on the pull request is the verdict.
- The full repository suite was not run locally (`plan.md` §D).
- AC-SSS-012 through AC-SSS-016 belong to M3-M4 and are not claimed.
- No latency figure is claimed. The skip is measured as a `ScanFile` count, never as elapsed time.
- `spec.md` §C.3's measured skip rates (go 0.9% / js 96.3% / py 85.7%) were NOT re-measured
  against this tree's token sets. The severity divergence noted above means go's real rate here is
  higher than 0.9%; the figure stands as the plan-time measurement it was recorded as.

**Residual risk.**

- **The extractor is the soundness-critical unit and it guards a deny.** Its correctness rests on
  every admitted token being mandatory. Three things bound that: the underivable-escalates
  fallback, the extraction-table tests (including the property arm that checks each admitted regex
  token against strings the expression actually matches), and the differential test's assertion
  (ii). None of the three is a proof.
- **`pattern:` tokens assume a match reproduces the pattern's literal text.** ast-grep matches
  ASTs, not text, so source written with unusual whitespace inside an identifier chain
  (`exec . Command(...)` — legal Go, rejected by gofmt) would not contain the token `exec.Command`
  and could be skipped. `spec.md` §C.2 asserts this reproduction property as the basis for the
  `pattern:` row; the risk is inherited from that decision, not introduced here.
- **The regex analyzer is hand-written, not `regexp/syntax`-backed.** It is deliberately
  conservative — every construct it does not model returns "" or underivable — but a construct it
  mis-models as literal would yield a token that never appears in real code, and a token that never
  appears is a token that always permits a skip. The metacharacter set is handled explicitly and
  the property arm samples the shipped shapes; unusual regexes outside the shipped ruleset are
  untested.
- **No caching.** `DerivePrefilters` re-reads and re-parses the rule files on every
  `scanWriteContent` call. The hook process is short-lived (one Write per invocation), so this is
  one parse per invocation — but a long-lived caller would pay it per call.
- **A language whose rules are all `warning`-severity is derivable with an empty token set, so it
  is always skipped.** That is correct for the deny (nothing can deny), and the only thing lost is
  the warning-count `slog.Info` line — which B-4 records as discarded on the `moai hook` path. If a
  future change makes warnings observable on this gate, this decision must be revisited.

### M3 — B: fold the PostToolUse guardian into the post-tool handler

Baseline attribution for every row below: measured in this run, against this tree, with the pre-M3
tree at `7d4805049` on branch `WT-security-scan-surface`. Redirected output is under
`.moai/state/verify/t217/`.

**RED evidence.** The three merge tests and the render test were written before any implementation
and run against the untouched tree.

`.moai/state/verify/t217/red-hook.txt` — `go test ./internal/hook/ -run 'TestPostToolGuardianMergeKeepsBothAdvisories|TestMergedGuardianTextMatchesStandalone|TestMergedGuardianNeverBlocks' -count=1`:

```
internal/hook/post_tool_guardian_test.go:66:16: undefined: NewPostToolGuardianHandler
internal/hook/post_tool_guardian_test.go:96:16: undefined: NewPostToolGuardianHandler
internal/hook/post_tool_guardian_test.go:118:14: undefined: NewPostToolGuardianHandler
internal/hook/post_tool_guardian_test.go:143:16: undefined: NewPostToolGuardianHandler
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
exit=1
```

`.moai/state/verify/t217/red-template.txt` — `go test ./internal/template/ -run TestRenderedSettingsHasNoSecurityScanPostToolEntry -count=1 -v`:

```
    settings_security_scan_entry_test.go:52: rendered settings still registers handle-security-scan on Write|Edit|MultiEdit (1 arg matches)
--- FAIL: TestRenderedSettingsHasNoSecurityScanPostToolEntry (0.00s)
    --- FAIL: TestRenderedSettingsHasNoSecurityScanPostToolEntry/darwin (0.00s)
    --- FAIL: TestRenderedSettingsHasNoSecurityScanPostToolEntry/linux (0.00s)
    --- FAIL: TestRenderedSettingsHasNoSecurityScanPostToolEntry/windows (0.00s)
exit=1
```

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-SSS-012 | PASS | `go test ./internal/hook/ -run TestPostToolGuardianMergeKeepsBothAdvisories -count=1 -v` | `--- PASS: TestPostToolGuardianMergeKeepsBothAdvisories (0.00s)` with both arms passing: `both_advisories_survive_the_merge` and `a_preceding_block_does_not_short-circuit_the_guardian`. The first arm asserts all four halves the criterion names — `systemMessage` equals the post-tool stub text exactly, `additionalContext` is non-empty and carries the `hardcoded-secret` finding, and neither field contains the other's content. The second arm registers a handler ahead of the guardian that returns `decision: block` (the dispatch log line `handler blocked action ... reason="blocked by an earlier handler"` records that the short-circuit really fired) and still observes a non-empty `additionalContext`, which is only reachable if the guardian's scan ran |
| AC-SSS-013 | PASS | `go test ./internal/hook/ -run TestMergedGuardianTextMatchesStandalone -count=1 -v` | `--- PASS: TestMergedGuardianTextMatchesStandalone (0.00s)`. The test compares the handler's `additionalContext` against the `additionalContext` `security.HandleSecurityScan` writes to stdout for the same payload, with `!=` on the raw strings — byte-identity is measured, not assumed. It is also structural: both surfaces now call the single producer `security.ScanBufferAdvisory`, so the banner and finding format cannot drift between them |
| AC-SSS-014 | PASS | `go test ./internal/hook/ -run TestMergedGuardianNeverBlocks -count=1 -v` | `--- PASS` all three arms (`critical_finding`, `clean_payload`, `empty_content`): nil error, `Decision` empty, `Continue` nil, `ExitCode` 0, `hookSpecificOutput.permissionDecision` empty and `hookSpecificOutput.decision` nil. The `critical_finding` arm is the load-bearing one — its payload trips the `hardcoded-secret` class, whose severity is `critical`, so the "even a critical finding does not block" half is actually observed rather than vacuously true on a clean buffer |
| AC-SSS-015 (settings.json) | PASS | `jq '[.hooks.PostToolUse[] \| select(.matcher=="Write\|Edit\|MultiEdit") \| .hooks[].args[-1]] \| map(select(test("handle-security-scan"))) \| length' .claude/settings.json` | `0` (measured `1` on the pre-M3 tree) |
| AC-SSS-015 (template) | PASS | `go test ./internal/template/ -run TestRenderedSettingsHasNoSecurityScanPostToolEntry -count=1 -v` | `--- PASS` for `darwin`, `linux`, and `windows`. The `.tmpl` is Go-templated and is not valid JSON, so the assertion runs against the rendered output, never against the raw file |
| AC-SSS-015 (subcommand retained) | PASS | `printf '{}' \| go run ./cmd/moai hook security-scan; echo "exit=$?"` | `exit=0` with empty stdout — the REQ-SSS-013 behaviour-preservation check. Run through `go run` against this tree rather than the installed `~/go/bin/moai`, so the result is attributable to this tree; re-installing the binary belongs to M4 |
| AC-SSS-001 (unchanged) | PASS | `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v` | `--- PASS: TestScanWriteContentDifferential (5.28s)` with `assertion (ii): 5 denying fixtures, none skippable by the pre-filter`. The M0 file was not edited in M3; `grep -c SKIP` over the verbose output returns `0`, so both assertions ran rather than one being skipped |

**Deliverables.**

- `internal/hook/post_tool_guardian.go` (new) — `postToolGuardianHandler`, registered on `EventPostToolUse`. Gated to `Write` / `Edit` / `MultiEdit`, fail-open on every path (a payload it cannot read yields a silent nil, never an error), and it carries its advisory on `hookSpecificOutput.additionalContext` so `mergeHandlerOutput` accumulates it beside the post-tool handler's `systemMessage`.
- `internal/hook/security/guardian.go` — two exported seams, both extracted from code `HandleSecurityScan` already ran rather than newly written: `ScanBufferAdvisory` (the single producer of the advisory text; `HandleSecurityScan` now calls it) and `ExtractToolInputContent` (the MultiEdit-aware extraction, now callable on an already-unwrapped `tool_input`; `extractWrittenContent` unwraps then delegates to it).
- `internal/hook/registry.go` — `alwaysRunHandler` marker interface plus `runAlwaysRunTail`, called on both short-circuit paths (`isBlockDecision` and `ExitCode == 2`). This is what makes the reachability invariant real rather than a registration-order convention: registering the guardian first would satisfy "nothing precedes it" today, but AC-SSS-012 asks for the guardian to survive a handler registered *before* it, which registration order cannot deliver.
- `internal/cli/deps.go` — the guardian registered alongside the post-tool handler.
- `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` — the `handle-security-scan.sh` entry removed from the `Write|Edit|MultiEdit` matcher on both surfaces, in this commit. The `handle-security-scan.sh` / `.sh.tmpl` wrapper pair is untouched (both files still present), because the subcommand it fronts is retained and a user project whose settings still name it must not hit the missing-hook log path.
- `internal/hook/post_tool_guardian_test.go`, `internal/template/settings_security_scan_entry_test.go` (new tests).
- `internal/cli/hook_e2e_test.go` — `TestHookDepsWiring_HandlerCounts` updated from `want 1` to `want 2` PostToolUse handlers. That assertion is the wiring guard for the deps.go registration, so M3 necessarily moves it; it was left as a real assertion on an exact count rather than relaxed to a lower bound.

**Suite results.**

- `go test ./internal/hook/... -count=1 -timeout 3600s` → exit 0, all 12 packages `ok` (`internal/hook` itself 144.994s). `.moai/state/verify/t217/test-hook-retry.txt`.
- `go test ./internal/cli/... -count=1 -timeout 1800s` → exit 0, all 17 packages `ok` (`internal/cli` itself 341.320s). `.moai/state/verify/t217/test-cli-retry.txt`.
- `go build ./...` → exit 0. `go vet ./internal/hook/... ./internal/cli/...` → exit 0.

**Gaps — what was NOT observed.**

- **The first `./internal/hook/...` run failed on a timeout, and that failure was environmental.** `go test ./internal/hook/... -count=1 -timeout 900s` panicked with `test timed out after 15m0s`, naming `TestScanWriteContentUnreadableConfigEscalates` as the running test. Measured cause: the machine was carrying load average **90.59** from parallel lanes at the time. Run alone, that test passes in **328.49s** (`.moai/state/verify/t217/isolated-unreadable.txt`), and the whole package then completes in 144.994s. The test is an M1/M2 deliverable that M3 does not touch, and nothing in the M3 diff runs on its path. The cost is real regardless of who owns it: three `sg`-spawning subtests take over five minutes together on a loaded machine, close enough to CI's budget to be worth watching.
- **The first `./internal/cli/...` run failed on `TestHookDepsWiring_HandlerCounts` (`got 2 handlers, want 1`)** — the expected consequence of registering a second PostToolUse handler. The assertion was updated and the whole suite re-run green; the first run is recorded at `.moai/state/verify/t217/test-cli.txt`.
- The full repository suite was not run locally (`plan.md` §D). CI on the pull request is the verdict.
- `golangci-lint` was not run in this milestone.
- M4 was not performed: no `make build` re-embed, no repository-wide mirror sweep, no template neutrality check, no pull request. AC-SSS-016 is not claimed.
- No latency or cost figure is claimed for the merge. The saving is one subprocess spawn per Write/Edit/MultiEdit event; it was not measured, and no acceptance criterion asks for it.
- The guardian was not exercised end-to-end through a real Claude Code PostToolUse invocation — only through the registry and the handler directly.

**Residual risk.**

- **`runAlwaysRunTail` changes `Dispatch` for every event, not just PostToolUse.** The blast radius is bounded by the marker: only handlers implementing `AlwaysRun() bool` run past a short-circuit, and the guardian is the only one in-tree. The decided verdict is preserved because `mergeHandlerOutput` keeps the first explicit halt and resolves permission decisions on a deny > ask > allow ladder, so an advisory tail cannot downgrade a block. A future handler that opts into the marker *and* returns a decision would break that reasoning.
- **The guardian now shares the post-tool handler's process and its hook timeout.** `ScanBuffer` is pure regex over a bounded buffer, so this is cheap — but it is no longer isolated in its own async subprocess, and a pathological payload that made the regex engine slow would consume the post-tool handler's budget instead of its own.
- **A user project that has not re-run `moai update` keeps the old settings entry**, so the guardian runs twice for a while: once in-process and once via the wrapper. Both paths are advisory and produce identical text, so the visible effect is a duplicated advisory, not a wrong one.
- **AC-SSS-013's byte-identity holds by construction only as long as both surfaces keep calling `ScanBufferAdvisory`.** A future edit that re-inlines the banner on either side would restore the drift the extraction removed; the test catches it, which is why it compares raw strings rather than asserting a substring.

### M4 — Distribution and disclosure (steps 1-3; the pull request is excluded)

Baseline attribution for every row below: measured in this run, against this tree, at HEAD
`6a56ae86a` on branch `WT-security-scan-surface`, merge base `a9eb896ce`. Redirected output is
under `.moai/state/verify/t217/`. M4 step 4 (opening the pull request) is **not** performed here
and is the orchestrator's to dispatch; AC-SSS-016's fifth row is therefore still open.

**Step 1 — mirror and pair verification, scoped to the diff.**

The audit reads `git diff --name-only a9eb896ce..HEAD` and is scoped to that list, never to the
repository (`plan.md` §F M4.1, `spec.md` §D).

```
$ git diff --name-only a9eb896ce..HEAD -- .claude/
.claude/settings.json

$ git diff --name-only a9eb896ce..HEAD -- internal/template/templates/.claude/
internal/template/templates/.claude/settings.json.tmpl

$ git diff --name-only a9eb896ce..HEAD -- '.claude/hooks/**' 'internal/template/templates/.claude/hooks/**'
(no output)
```

One `.claude/` path changed and its template mirror changed in the same branch — the mirror
obligation is satisfied. No hook wrapper appears in the diff, so the `.sh` / `.sh.tmpl` pair
obligation is vacuously satisfied rather than assumed: the pair check has nothing in scope to
check. No mirror was missing, so none was added.

**Step 1 — the two `driftcheck.sh` axes.**

```
$ bash .moai/reports/t217/driftcheck.sh pairaxis          → 1 line, exit 0
DRIFT .claude/hooks/moai/handle-pre-tool.sh

$ bash .moai/reports/t217/driftcheck.sh guarded            → 0 lines
(no output)
```

The single `pairaxis` line is the pre-existing, out-of-scope drift `acceptance.md` AC-SSS-016
predicted and `spec.md` §D records: `handle-pre-tool.sh` legitimately differs from its template in
comments. That it is pre-existing is measured, not assumed —
`git diff --stat a9eb896ce..HEAD -- .claude/hooks/moai/handle-pre-tool.sh internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl`
returns empty output, so this branch did not touch either file. **Scoped to this SPEC's diff, both
axes return 0.** Evidence: `.moai/state/verify/t217/driftcheck-pairaxis.txt`,
`.moai/state/verify/t217/driftcheck-guarded.txt`.

Two measurement notes, so a later reader does not mis-read the raw numbers: the `guarded` mode
exits 1 while printing nothing, because its last `[ -f "$b" ]` test is false — the check result is
the empty output, not the exit code. The `unguarded` mode prints 31 lines, which is the documented
false-positive count (31 of 35 `.tmpl` files have no `.sh` sibling inside `templates/`); it is
recorded for context only and is not a gate.

**Step 2 — `make build`.**

```
$ make build   → exit 0
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "... -X ...version.Version=v3.1.2 -X ...version.Commit=6a56ae86a
  -X ...version.Date=2026-08-24T13:50:49Z" -o bin/moai ./cmd/moai

$ ./bin/moai version
 v3.1.2   6a56ae86a   built 2026-08-24T13:50:49Z
```

The binary rebuilt and reports this tree's HEAD, so the re-embed is attributable to this commit
rather than to a stale artifact. `make build` regenerated `internal/template/catalog.yaml` and its
agent hashes, and the result is byte-identical to the committed file —
`git status --porcelain internal/template/` returns empty. **No version-controlled artifact
changed**, so the M4 commit carries no re-embedded template file. Evidence:
`.moai/state/verify/t217/make-build.txt`.

**Step 3 — template neutrality self-check over the changed template file.**

The changed template surface is exactly one file, `internal/template/templates/.claude/settings.json.tmpl`,
and its diff is a **pure deletion** — the `handle-security-scan.sh` PostToolUse entry removed by M3,
with no line added. A deletion cannot introduce forbidden content, but the five checks were run
against the post-change file rather than inferred from that reasoning:

| Item | Pattern | Count |
|---|---|---|
| C1 SPEC ID | `SPEC-[A-Z0-9]+-` | `0` |
| C2 REQ / AC token | `REQ-[A-Z]{2,}-[0-9]{3}\|AC-[A-Z]{2,}-[0-9]{3}` | `0` |
| C3 audit citation | `Audit [0-9]\|Finding A[0-9]\|spec\.md` | `0` |
| C4 date / short-sha | `20[2-9][0-9]-[01][0-9]-[0-3][0-9]\|[0-9a-f]{7,8}` | `0` |
| C5 maintainer path | `CLAUDE\.local\|/Users/goos\|\.moai/backups` | `0` |

Each row is `grep -cE '<pattern>' internal/template/templates/.claude/settings.json.tmpl`,
which printed `0` and exited 1 (no match) on all five. The counts are read from `grep -c` directly
rather than through a pipe, so a `0` here is an observation and not a swallowed failure.

**The previously-UNVERIFIED template-leak row is now measured.**

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1 -timeout 900s   → exit 0
ok  	github.com/modu-ai/moai-adk/internal/template	28.425s
```

The authoring-time Gap was a measurement failure, not a test failure: the run exceeded a 120 s
Bash-call budget and returned empty output, which is silence rather than a verdict. Re-run with an
explicit `-timeout 900s` and a 900 s call budget, the test completes in **28.4 s** and passes. The
`acceptance.md` AC-SSS-016 row has been updated from `UNVERIFIED` to the measured result. Evidence:
`.moai/state/verify/t217/template-leak-strict.txt`.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-SSS-016 (mirror + pair, diff-scoped) | PASS | `git diff --name-only a9eb896ce..HEAD -- .claude/` and the mirror / wrapper variants above | one `.claude/` path (`settings.json`), its mirror (`settings.json.tmpl`) changed in the same branch, zero hook wrappers in the diff |
| AC-SSS-016 (`pairaxis`) | PASS | `bash .moai/reports/t217/driftcheck.sh pairaxis` | `DRIFT .claude/hooks/moai/handle-pre-tool.sh` — 1 line repo-wide, **0 lines scoped to this diff**, and measured pre-existing (the branch touches neither file) |
| AC-SSS-016 (`guarded`) | PASS | `bash .moai/reports/t217/driftcheck.sh guarded` | no output (0 lines) |
| AC-SSS-016 (template leak, strict) | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1 -timeout 900s` | `ok github.com/modu-ai/moai-adk/internal/template 28.425s`, exit 0 — the authoring-time Gap is closed |
| AC-SSS-016 (PR declaration) | **OPEN** | `gh pr view --json title,body ...` | not measured — M4 step 4 is excluded from this delegation. No pull request exists on this branch |
| AC-SSS-001 (unchanged, re-run) | PASS | `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v -timeout 900s` | `--- PASS: TestScanWriteContentDifferential (0.09s)`, with `assertion (ii): 5 denying fixtures, none skippable by the pre-filter` logged and five `security scan blocked write operation` lines from assertion (i)'s corpus loop. `grep -c 'assertion'` returns 1 — assertion (i) carries no log line by construction; it is the loop that precedes (ii) in the same function, and (ii) is only reached if (i) ran. The M0 file was not edited |

**Suite results (this tree, after `make build`).**

- `go build ./...` → exit 0. `.moai/state/verify/t217/go-build.txt`
- `go vet ./internal/hook/... ./internal/cli/... ./internal/template/...` → exit 0. `.moai/state/verify/t217/go-vet.txt`
- `GOOS=windows GOARCH=amd64 go vet` (same packages) → exit 0. `.moai/state/verify/t217/vet-windows.txt`
- `GOOS=linux GOARCH=amd64 go vet` (same packages) → exit 0. `.moai/state/verify/t217/vet-linux.txt`
- `go test ./internal/template/... -count=1 -timeout 900s` → exit 0, all packages `ok` (`internal/template` 25.533s). `.moai/state/verify/t217/template-test.txt`
- `go test ./internal/hook/... -count=1 -timeout 900s` → exit 0, all 11 packages `ok` (`internal/hook` 49.970s). `.moai/state/verify/t217/hook-test.txt`

**Working-tree hygiene.** `go test ./internal/hook/...` regenerates
`.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md` — those files carry benchmark
timings that the `internal/hook/perf` suite rewrites on every run. They belong to a different,
already-closed SPEC and are outside this SPEC's scope, so they were restored with `git restore`
and are not staged. `git status --porcelain` is empty apart from this milestone's own edits.

**Gaps — what was NOT observed.**

- **AC-SSS-016's fifth row (the PR title / first-body-sentence security-surface declaration,
  REQ-SSS-016) is NOT measured.** No pull request exists on this branch. AC-SSS-016 is therefore
  **not fully PASS**: four of its five rows are measured PASS and the fifth is open pending M4
  step 4. It is recorded here as open rather than claimed.
- No push was performed, so nothing on this branch is on the remote. `l44_post_push_fetch` below
  is recorded as not-applicable, not as a clean result.
- `golangci-lint` was not run in this milestone (unchanged from M3).
- The full repository suite was not run locally (`plan.md` §D). CI on the pull request is the verdict.
- `GOOS=windows` / `GOOS=linux` `go vet` proves the packages **compile** cross-platform; it does
  not prove runtime behaviour on those platforms. No cross-platform runtime claim is made.
- The neutrality self-check covers the one changed template file. It is not a repository-wide
  neutrality sweep, and none is claimed.
- The rebuilt `bin/moai` was **not** installed to `~/go/bin/moai`. Nothing in M4 was verified
  through the installed binary.

**Residual risk.**

- **The pre-existing `handle-pre-tool.sh` pair drift stays unresolved.** It is out of scope by
  `spec.md` §D, and scoping the gate to the diff is what keeps this SPEC from being blocked by it —
  but a future SPEC that does touch that wrapper inherits the drift, and the diff-scoped gate will
  then correctly refuse to ignore it.
- **The template-leak test's cost is close to a default Bash budget.** It passes in ~28 s on an
  unloaded machine, but the authoring-time run exceeded 120 s under load. A future run on a loaded
  machine can time out again and produce the same silence-mistaken-for-failure; the explicit
  `-timeout 900s` and the redirected evidence file are the mitigation, not a fix for the cost.
- **A user project that has not re-run `moai update` keeps the old settings entry** (unchanged from
  M3): the guardian runs twice for a while, once in-process and once via the wrapper. Both paths
  are advisory and produce identical text.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24T13:59:00Z
run_commit_sha: efbb21196             # the M4 commit that produced the evidence above; backfilled by the review-remediation commit, since a commit cannot name its own hash
run_status: complete-except-pr        # M1-M4 steps 1-3 landed; M4 step 4 (pull request) is the orchestrator's
ac_pass_count: 15                     # AC-SSS-001 .. AC-SSS-015
ac_fail_count: 0
ac_open_count: 1                      # AC-SSS-016 — 4 of 5 rows PASS, the PR-declaration row open pending M4 step 4
preserve_list_post_run_count: 2       # plan.md §A PRESERVE items, both intact
preserve_list_post_run_detail:
  - "internal/hook/security/guardian.go § HandleSecurityScan — present (grep -c 'func HandleSecurityScan' = 1); the advisory contract is preserved and now shared via ScanBufferAdvisory"
  - ".claude/rules/moai/core/verification-claim-integrity.md — untouched by this branch (empty git diff over the path)"
l44_pre_commit_fetch: "git merge-base origin/main HEAD = a9eb896ce; HEAD re-read immediately before commit"
l44_post_push_fetch: "pushed at the review-remediation step; branch WT-security-scan-surface -> PR #1643, head efbb21196 == remote tip == PR headRefOid (verified by ls-remote and gh pr view)"   # the authoring-time "no push performed" was true of the M4 delegation only
new_warnings_or_lints_introduced: 0   # go vet clean on darwin, windows, and linux. The authoring-time "golangci-lint not run (Gap)" is CLOSED: the orchestrator ran `golangci-lint run ./internal/hook/... ./internal/cli/... --timeout=10m` at the M4 gate and observed `0 issues.` — evidence `.moai/state/verify/t217/orch-m4-lint.log`
cross_platform_build:
  darwin_arm64_build: "`go env GOOS GOARCH` → darwin arm64 (this host), so the bare `go build ./...` → exit 0 above IS the darwin/arm64 build; no cross-compile env was needed for this row"
  darwin_vet: "go vet ./internal/hook/... ./internal/cli/... ./internal/template/... → exit 0"
  windows_amd64_vet: "GOOS=windows GOARCH=amd64 go vet (same packages) → exit 0"
  linux_amd64_vet: "GOOS=linux GOARCH=amd64 go vet (same packages) → exit 0"
  note: "vet proves cross-platform compilation, including test files; it does not prove runtime behaviour on those platforms"
total_run_phase_files: 23             # git diff --name-status a9eb896ce..HEAD, before the M4 commit
m1_to_mN_commit_strategy: "one commit per milestone on WT-security-scan-surface: M1 dad1f5bb4, M2 7d4805049, M3 6a56ae86a, M4 this commit. No squash, no amend, no force-push."
evidence_dir: .moai/state/verify/t217/
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
