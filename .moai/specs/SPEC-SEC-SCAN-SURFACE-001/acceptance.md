# Acceptance Criteria — SPEC-SEC-SCAN-SURFACE-001

Every criterion below is closed by **running something** and reading its output. No criterion is
closed by a grep over source text. Every criterion states the value measured on the
**pre-implementation tree** next to its PASS value; where the two are equal the criterion
observes nothing and does not belong here.

No criterion asserts a latency figure. Cost is expressed as a count of scans dispatched and temp
files created.

**The instrument.** Counts are taken on a fake scanner injected through the interface M0 step 1
introduces, and by snapshotting the process temp directory. `ScanFile` is the only route from
this path to an `sg` spawn, so a `ScanFile` count of 0 proves a spawn count of 0 — exact for
every skip case asserted below. `astGrepScanner.scanFunc` is **not** the instrument: it is
consulted only inside `ScanMultiple` (`ast_grep.go:199-200`) and the single-file `Scan` execs
`sg` directly at `:137`.

---

## §A Definition of Done

- AC-SSS-001 through AC-SSS-016 all pass, each with the command run and its verbatim output cited.
- `go test ./internal/hook/...` and `go test ./internal/cli/...` pass on the branch.
- `go vet ./...` is clean; `golangci-lint run` reports no new finding.
- `~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md` reports 0 errors.
  (The linter takes files, not directories — a directory argument fails with
  `ParseFailure ... is a directory`, measured.)
- CI is green on the pull request. CI, not a local run, is the verdict for the full suite.

---

## §B Criteria

### B.1 The invariant

**AC-SSS-001 — the deny verdict is unchanged, and the pre-filter suppresses no deny.**
Given the differential fixture corpus of plan §F M0, recorded against the **unmodified** gate,
When the corpus is replayed through `scanWriteContent` after M1, M2, and M3 have landed,
Then (i) every fixture yields the identical `(decision, reason-nonempty)` pair it yielded before,
and (ii) for every fixture that denies, the derived pre-filter would not have skipped it.
Command: `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v`.
Pre-implementation measurement: assertion (i)'s expectations are generated from the untouched
tree in M0 and committed before any behaviour change; assertion (ii) does not compile before M2.
Corpus validity gate: at least one fixture per covered language must **deny** on the
pre-implementation tree, else the corpus observes nothing and is rejected.

---

### B.2 Item A2 — no rules config, no scan

**AC-SSS-002 — no config resolves ⇒ no scan dispatched.**
Given a project root containing no `sgconfig.yml`, no `.ast-grep/`, and no
`.moai/config/astgrep-rules/`,
When a `.go` Write payload is passed to `scanWriteContent` with a counting fake scanner,
Then the fake records **0** `ScanFile` calls and the decision is allow.
Command: `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoScan -count=1 -v`.
Pre-implementation measurement: **1** `ScanFile` call. RED before M1.

**AC-SSS-003 — no config resolves ⇒ no temp file.**
Given the same no-config project root,
When the same payload is scanned,
Then no file matching `moai-security-scan-*` appears in the process temp directory during or
after the call (snapshot before, snapshot after, compare the matching sets).
Command: `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoTempFile -count=1 -v`.
Pre-implementation measurement: exactly **1** such file is created (and then removed by the
deferred cleanup, which is why the check is a during-call snapshot). RED before M1.

**AC-SSS-004 — the scanner performs no second config resolution.**
Given a `RuleManager` stub counting `FindRulesConfig` calls, wired into the scanner, and a
caller-side resolution counter,
When one Write payload is processed end to end with a resolvable config,
Then the **scanner-side** counter reads **0** and the caller-side counter reads **1**.
Command: `go test ./internal/hook/ -run TestConfigResolvedByCallerNotScanner -count=1 -v`.
Pre-implementation measurement: scanner-side **1**, caller-side **0** — measured directly:
`grep -rn "FindRulesConfig" internal/ | grep -v _test` shows the only pre-write-path call at
`internal/hook/security/scanner.go:84`, inside `ScanFile`. The two counters invert, so the
criterion cannot pass on the untouched tree.

---

### B.3 Item A3 — no rules for this language, no scan

**AC-SSS-005 — an uncovered language dispatches no scan.**
Given a project root carrying the shipped ruleset (four covered languages),
When a payload whose extension maps to an uncovered but ast-grep-supported language is scanned
(one case per uncovered language: `.rs`, `.java`, `.kt`, `.c`, `.cpp`, `.rb`, `.php`, `.swift`,
`.cs`, `.ex`, `.scala`),
Then the fake records **0** `ScanFile` calls for every case, and a `.go` control case in the same
test records **1**.
Command: `go test ./internal/hook/ -run TestScanWriteContentUncoveredLanguage -count=1 -v`.
Pre-implementation measurement: **11** calls. The `.go` control proves the test is not passing by
scanning nothing at all.

**AC-SSS-006 — the covered-language set comes from the config, not from a list in the code.**
Given a temp project root carrying a **modified** copy of the shipped ruleset whose `ruleDirs` no
longer names the directory holding the `language: go` rules,
When a `.go` payload is scanned,
Then the fake records **0** `ScanFile` calls — the covered set followed the config.
And given the unmodified ruleset, the same payload records **1**.
Command: `go test ./internal/hook/security/ -run TestCoveredLanguagesFollowConfig -count=1 -v`.
Pre-implementation measurement: **1** in both arms. No hardcoded language list can produce the
`0`/`1` split, which is what makes this an observation of REQ-SSS-005 rather than a restatement.

**AC-SSS-007 — an unreadable or empty result escalates rather than skips.**
Given three project roots — one whose resolved `sgconfig.yml` is malformed YAML, one whose
`ruleDirs` names a directory that does not exist, and one whose rule files declare no `language:`
at all,
When a `.go` payload is scanned against each,
Then the fake records **1** `ScanFile` call in every case (fail-open), not 0.
Command: `go test ./internal/hook/security/ -run TestUnreadableOrEmptyConfigEscalates -count=1 -v`.
Pre-implementation measurement: **1** in all three arms — this criterion guards against a
regression M1 could introduce, and is the one criterion whose PASS value equals today's value by
design. It is retained because M1's skip logic is what could break it; it is a
behaviour-preservation criterion, not a change-detection one, and is labelled as such.

---

### B.4 Item A1 — the derived pre-filter

**AC-SSS-008 — the pre-filter is derived only from `error`-severity rules.**
Given the shipped ruleset parsed in memory,
When the derivation function is called,
Then no token contributed exclusively by a rule at `warning`, `info`, or `hint` severity appears
in the result, and at least one token from each derivable `error`-severity rule does.
Command: `go test ./internal/hook/security/ -run TestPrefilterDerivationSeverityScope -count=1 -v`.
Pre-implementation measurement: the function does not exist; the test does not compile. RED.

**AC-SSS-009 — the shipped ruleset's dominant shape is derivable, with the right tokens.**
Given `sec-hardcoded-credential` as shipped — `kind:` + `regex:` with no `pattern:`, in all four
covered languages,
When the derivation function is called,
Then none of the four languages is marked underivable **by this rule**, and the tokens derived
from it are exactly `{sk-, AKIA, ghp_, xox, AIza}` — the mandatory literal prefix of each branch
of the regex's top-level alternation.
Command: `go test ./internal/hook/security/ -run TestPrefilterKindPlusRegexAlternation -count=1 -v`.
Pre-implementation measurement: does not compile. This is the criterion the plan audit's D2
finding produced: under the v0.1.0 rules this rule read as `kind:`-only, all four languages went
underivable, and item A1 skipped nothing anywhere.

**AC-SSS-010 — an unrecognized rule shape marks the language underivable, and underivable always
escalates.**
Given synthetic rule sets containing, respectively, a `regex:` with no literal anchor, an `any:`
with one tokenless branch, and an `inside:` composite,
When the derivation function is called for each,
Then each returns the `underivable` marker for that rule's language.
And given a project root whose ruleset makes a covered language underivable, when a payload in
that language containing no dangerous construct is scanned, then the fake records **1**
`ScanFile` call.
Commands: `go test ./internal/hook/security/ -run TestPrefilterUnderivableShapes -count=1 -v`
and `go test ./internal/hook/ -run TestScanWriteContentUnderivableEscalates -count=1 -v`.
Pre-implementation measurement: neither compiles. RED.

**AC-SSS-011 — the pre-filter skips a payload no error rule can match.**
Given the shipped ruleset and a `.js` payload containing none of the tokens derived for
javascript,
When it is scanned,
Then the fake records **0** `ScanFile` calls and the decision is allow.
Command: `go test ./internal/hook/ -run TestScanWriteContentPrefilterSkip -count=1 -v`.
Pre-implementation measurement: **1**. The language is javascript rather than go deliberately —
`spec.md` §C.3 measured the Go skip rate at 1.2% against javascript's 96.3%, so a Go-based
version of this criterion would be asserting the thing that measurement says will not happen.

---

### B.5 Item B — the merge preserves both advisories

**AC-SSS-012 — both payloads survive the merge, and the guardian is reachable.**
Given a registry with both the post-tool handler and the guardian handler registered, the
post-tool handler producing a non-empty `systemMessage` and the guardian producing a non-empty
`hookSpecificOutput.additionalContext` for the same `Write` event,
When the event is dispatched once,
Then the single emitted output carries the post-tool text in `systemMessage` **and** the guardian
banner text in `hookSpecificOutput.additionalContext`, with neither field empty and neither
containing only the other's content.
And when a handler registered **before** the guardian returns a block decision, then the guardian
scan is still reached (`spec.md` §A.3 — `registry.go:142` short-circuits on a block).
Command: `go test ./internal/hook/ -run TestPostToolGuardianMergeKeepsBothAdvisories -count=1 -v`.
Pre-implementation measurement: does not compile — no guardian PostToolUse handler exists to
register. RED in M3.

**AC-SSS-013 — the guardian's advisory text is unchanged by the merge.**
Given a buffer that produces a known guardian finding,
When it is processed through the merged handler and, separately, through `HandleSecurityScan`
directly,
Then the `additionalContext` strings are byte-identical.
Command: `go test ./internal/hook/ -run TestMergedGuardianTextMatchesStandalone -count=1 -v`.
Pre-implementation measurement: does not compile. RED.

**AC-SSS-014 — the merge introduces no block.**
Given any buffer, including one producing a critical-severity guardian finding,
When it is processed through the merged handler,
Then the emitted output carries no `decision`, no `permissionDecision`, and no `continue: false`,
and the handler returns a nil error.
Command: `go test ./internal/hook/ -run TestMergedGuardianNeverBlocks -count=1 -v`.
Pre-implementation measurement: does not compile. RED.

**AC-SSS-015 — the settings entry is gone from both surfaces, and the subcommand still answers.**
Given the shipped settings and the rendered template,
When the `Write|Edit|MultiEdit` matcher's hook list is inspected in each,
Then neither contains a `handle-security-scan.sh` entry; and the retained subcommand still
answers.
Commands and their pre-implementation measurements:

| Command | PASS | Measured today |
|---|---|---|
| `jq '[.hooks.PostToolUse[] \| select(.matcher=="Write\|Edit\|MultiEdit") \| .hooks[].args[-1]] \| map(select(test("handle-security-scan"))) \| length' .claude/settings.json` | `0` | `1` |
| `go test ./internal/template/ -run TestRenderedSettingsHasNoSecurityScanPostToolEntry -count=1 -v` | pass | the rendered output contains 1 such entry (the `.tmpl` is Go-templated and is **not** valid JSON, so it is asserted through a render, never through `jq`) |
| `printf '{}' \| ./bin/moai hook security-scan; echo "exit=$?"` | `exit=0` | `exit=0` — a behaviour-preservation check for REQ-SSS-013, labelled as such |

---

### B.6 Distribution and disclosure

**AC-SSS-016 — mirrors, pairs, neutrality, and the PR declaration.**
Given this branch's diff against its merge base,
When each changed path under `.claude/` is checked for a corresponding changed path under
`internal/template/templates/.claude/`, each changed hook wrapper is checked for its `.sh` /
`.sh.tmpl` sibling, and the changed template files are checked for forbidden content,
Then every `.claude/` change has a mirror, every changed wrapper pair moved together, and no
template file introduces a SPEC ID, a REQ token, a date, a commit SHA, an absolute macOS path, or
a `CLAUDE.local` reference.

Commands and their pre-implementation measurements:

| Command | PASS | Measured today |
|---|---|---|
| `git diff --name-only <merge-base>...HEAD` — the mirror and pair audit reads this list, **scoped to the diff** | every `.claude/` entry paired | this SPEC changes `settings.json` only; it touches no hook wrapper |
| `for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do b=${f%.tmpl}; [ -f "$b" ] && { diff -q "$b" "$f" >/dev/null \|\| echo "DRIFT $b"; }; done` | prints nothing | prints nothing (**0** lines). Without the `[ -f "$b" ]` guard the same loop prints **31** lines, because 31 of 35 `.tmpl` files legitimately have no `.sh` sibling inside `templates/`. The guard is the canonical form from `CLAUDE.local.md` §2.3 |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` | pass | pass |
| `gh pr view --json title,body --jq '.title, (.body \| split("\n")[0])'` | title and first body line each name this a change to the **security scan surface**, neither presents it as a performance improvement (REQ-SSS-016) | no PR exists until M4 |

Note on scope: a repository-wide `.sh` ↔ `.sh.tmpl` byte-equality sweep is **not** a valid gate.
`.claude/hooks/moai/handle-pre-tool.sh` already differs from its template in comments today, for
a legitimate reason (template neutrality strips the SPEC ID the deployed copy carries) —
`spec.md` §D records this as out of scope. The pair check above is therefore scoped to wrappers
this SPEC's diff changes, of which there are currently none.

---

## §C Edge cases the criteria deliberately cover

| Edge case | Covered by |
|---|---|
| Config present but malformed | AC-SSS-007 |
| `ruleDirs` names a missing directory | AC-SSS-007 (observed live: `sg` reports `Cannot read rule directory`) |
| Derived covered-language set is empty | AC-SSS-007 |
| Rule shape unparseable in a covered language | AC-SSS-010 |
| The shipped ruleset's `kind:` + `regex:` dominant shape | AC-SSS-009 |
| Payload in a covered language with only `warning`-severity matches | AC-SSS-001 fixture set |
| Payload in an ast-grep-supported but rule-uncovered language | AC-SSS-005 |
| Guardian finding and LSP advisory on the same event | AC-SSS-012 |
| A preceding handler returning a block decision | AC-SSS-012 |
| A user project whose `settings.json` still names `handle-security-scan.sh` | AC-SSS-015 |

## §D Explicitly not asserted

- No criterion asserts a wall-clock duration, a percentage speedup, or a millisecond figure. The
  audit's absolute numbers were taken at load 8-10 and the investigation at load 38; they are
  ratios, not guarantees.
- No criterion asserts that the pre-write deny has ever fired for a real user. Investigation
  Claim 1 establishes zero firings **on this machine only**; distributed-user behaviour is
  outside the observation window and is not claimed in either direction.
- No criterion asserts that item A1 reduces scans for Go payloads. `spec.md` §C.3 measured that
  rate at 1.2% and states why.
- Three criteria (AC-SSS-007, and the third row of AC-SSS-015) have a PASS value equal to today's
  value. They are **behaviour-preservation** criteria — the change is what could break them — and
  are labelled as such rather than presented as change detection.
