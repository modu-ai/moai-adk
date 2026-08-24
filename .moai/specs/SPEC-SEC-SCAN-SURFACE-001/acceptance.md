# Acceptance Criteria — SPEC-SEC-SCAN-SURFACE-001

Every criterion below is closed by **running something** and reading its output. No criterion is
closed by a grep over source text. Where a criterion is measured by a count, its
pre-implementation baseline is stated so the measurement is known to observe something.

No criterion asserts a latency figure. Cost is expressed as a count of temporary files created
and `sg` processes spawned, both observable through the scanner's injectable scan function.

---

## §A Definition of Done

- AC-SSS-001 through AC-SSS-016 all pass, each with the command run and its verbatim output cited.
- `go test ./internal/hook/...` and `go test ./internal/cli/...` pass on the branch.
- `go vet ./...` is clean; `golangci-lint run` reports no new finding.
- CI is green on the pull request. CI, not a local run, is the verdict for the full suite.

---

## §B Criteria

### B.1 The invariant

**AC-SSS-001 — the deny verdict is unchanged as a function.**
Given the differential fixture corpus of plan §F M0, recorded against the **unmodified** gate,
When the corpus is replayed through `scanWriteContent` after M1, M2, and M3 have landed,
Then every fixture yields the identical `(decision, reason-nonempty)` pair it yielded before.
Command: `go test ./internal/hook/ -run TestScanWriteContentDifferential -count=1 -v`.
Baseline: the recorded expectations are produced from the pre-implementation tree in M0 and are
committed before any behaviour change; a corpus that produced no deny at all would observe
nothing, so the corpus is rejected unless at least one fixture per covered language denies on the
pre-implementation tree.

---

### B.2 Item A2 — no rules config, no scan

**AC-SSS-002 — no config resolves ⇒ no `sg` spawn.**
Given a project root containing no `sgconfig.yml`, no `.ast-grep/`, and no
`.moai/config/astgrep-rules/`,
When a `.go` Write payload is passed to `scanWriteContent` with a scanner whose scan function is
a counting stub,
Then the counting stub records **0** invocations and the decision is allow.
Command: `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoSpawn -count=1 -v`.
Baseline: on the pre-implementation tree the same test records **1** invocation — the test is a
RED test before M1 and must be shown failing first.

**AC-SSS-003 — no config resolves ⇒ no temp file.**
Given the same no-config project root,
When the same payload is scanned,
Then no file matching `moai-security-scan-*` exists in the process temp directory during or after
the call.
Command: `go test ./internal/hook/ -run TestScanWriteContentNoConfigNoTempFile -count=1 -v`
(the test snapshots the temp dir before and after and asserts equality of the matching set).
Baseline: fails on the pre-implementation tree, which creates the file before scanning.

**AC-SSS-004 — the config is resolved once.**
Given a `RuleManager` stub that counts `FindRulesConfig` calls,
When one Write payload is processed end to end,
Then the counter reads exactly **1**.
Command: `go test ./internal/hook/ -run TestScanWriteContentResolvesConfigOnce -count=1 -v`.
Baseline: **2** on the pre-implementation tree (once in the caller after M1's reorder, once
inside `ScanFile`) — this criterion is measured after M1's reorder introduces the caller-side
resolution and must show that the scanner-side resolution was removed, not duplicated.

---

### B.3 Item A3 — no rules for this language, no scan

**AC-SSS-005 — an uncovered language does not spawn `sg`.**
Given a project root carrying the shipped ruleset (four covered languages),
When a payload whose extension maps to an uncovered but ast-grep-supported language is scanned
(one case per uncovered language: `.rs`, `.java`, `.kt`, `.c`, `.cpp`, `.rb`, `.php`, `.swift`,
`.cs`, `.ex`, `.scala`),
Then the counting stub records **0** invocations for every case, and a `.go` control case in the
same test records **1**.
Command: `go test ./internal/hook/ -run TestScanWriteContentUncoveredLanguage -count=1 -v`.
Baseline: **11** invocations on the pre-implementation tree; the `.go` control proves the test is
not passing by scanning nothing at all.

**AC-SSS-006 — the covered-language set comes from the config, not from a list in the code.**
Given a temp project root carrying a **modified** copy of the shipped ruleset in which every
`language: go` rule file has been removed from `ruleDirs`,
When a `.go` payload is scanned,
Then the counting stub records **0** invocations — the covered set followed the config.
And given the unmodified ruleset, the same payload records **1**.
Command: `go test ./internal/hook/security/ -run TestCoveredLanguagesFollowConfig -count=1 -v`.
Baseline: this test cannot pass against any hardcoded language list, which is what makes it an
observation of REQ-SSS-006 rather than a restatement of it.

**AC-SSS-007 — an unreadable config escalates rather than skips.**
Given a project root whose resolved `sgconfig.yml` is present but malformed YAML,
When a `.go` payload is scanned,
Then the counting stub records **1** invocation (fail-open to `sg`), not 0.
Command: `go test ./internal/hook/security/ -run TestUnreadableConfigEscalates -count=1 -v`.

---

### B.4 Item A1 — the derived pre-filter

**AC-SSS-008 — the pre-filter is derived only from `error`-severity rules.**
Given the shipped ruleset parsed in memory,
When the derivation function is called,
Then no token contributed by a rule whose severity is `warning`, `info`, or `hint` appears in the
result, and at least one token from each derivable `error`-severity rule does.
Command: `go test ./internal/hook/security/ -run TestPrefilterDerivationSeverityScope -count=1 -v`.

**AC-SSS-009 — an unrecognized rule shape marks the language underivable.**
Given synthetic rule sets containing, respectively, a `regex:` constraint with no literal anchor,
an `any:` with one tokenless branch, and an `inside:` composite,
When the derivation function is called for each,
Then each returns the `underivable` marker for that rule's language.
Command: `go test ./internal/hook/security/ -run TestPrefilterUnderivableShapes -count=1 -v`.

**AC-SSS-010 — an underivable language always escalates.**
Given a project root whose ruleset makes a covered language underivable,
When a payload in that language with no dangerous construct is scanned,
Then the counting stub records **1** invocation — the pre-filter never suppresses an
underivable language.
Command: `go test ./internal/hook/ -run TestScanWriteContentUnderivableEscalates -count=1 -v`.

**AC-SSS-011 — the pre-filter skips a payload no error rule can match.**
Given the shipped ruleset and a `.js` payload containing none of the tokens derived for
javascript,
When it is scanned,
Then the counting stub records **0** invocations and the decision is allow.
Command: `go test ./internal/hook/ -run TestScanWriteContentPrefilterSkip -count=1 -v`.
Baseline: **1** on the pre-implementation tree. The language is javascript rather than go
deliberately — `spec.md` §C.2 records that the Go pre-filter rarely skips while
`go-error-ignored-blank` is an `error` rule, so a Go-based version of this criterion would be
observing the wrong thing.

---

### B.5 Item B — the merge preserves both advisories

**AC-SSS-012 — both payloads survive the merge.** *(the card's [HARD] constraint)*
Given a registry with both the post-tool handler and the guardian handler registered, the
post-tool handler producing a non-empty `systemMessage` and the guardian producing a non-empty
`hookSpecificOutput.additionalContext` for the same `Write` event,
When the event is dispatched once,
Then the single emitted output carries the post-tool text in `systemMessage` **and** the guardian
banner text in `hookSpecificOutput.additionalContext`, with neither field empty and neither
containing only the other's content.
Command: `go test ./internal/hook/ -run TestPostToolGuardianMergeKeepsBothAdvisories -count=1 -v`.
Baseline: the test does not compile against the pre-implementation tree, because no guardian
PostToolUse handler exists to register. It is written RED in M3 and must be shown failing before
the handler is added.

**AC-SSS-013 — the guardian's advisory text is unchanged by the merge.**
Given a buffer that produces a known guardian finding,
When it is processed through the merged handler and, separately, through
`HandleSecurityScan` directly,
Then the `additionalContext` strings are byte-identical.
Command: `go test ./internal/hook/ -run TestMergedGuardianTextMatchesStandalone -count=1 -v`.

**AC-SSS-014 — the merge introduces no block.**
Given any buffer, including one producing a critical-severity guardian finding,
When it is processed through the merged handler,
Then the emitted output carries no `decision`, no `permissionDecision`, and no `continue: false`,
and the handler returns a nil error.
Command: `go test ./internal/hook/ -run TestMergedGuardianNeverBlocks -count=1 -v`.

**AC-SSS-015 — the settings entry is gone from both surfaces, and the subcommand still answers.**
Given the shipped settings and the rendered template,
When the `Write|Edit|MultiEdit` matcher's hook list is inspected in each,
Then neither contains a `handle-security-scan.sh` entry.
Commands:
`jq '[.hooks.PostToolUse[] | select(.matcher=="Write|Edit|MultiEdit") | .hooks[].args[-1]] | map(select(test("handle-security-scan"))) | length' .claude/settings.json` → must print `0`
(measured on the pre-implementation tree: prints `1`);
`go test ./internal/template/ -run TestRenderedSettingsHasNoSecurityScanPostToolEntry -count=1 -v`
— the template is Go-templated and is **not** valid JSON, so it is asserted through a render, not
through `jq` (measured on the pre-implementation tree: the rendered output contains 1 such entry);
and `printf '{}' | ./bin/moai hook security-scan; echo "exit=$?"` → must print `exit=0`,
demonstrating the retained subcommand still answers (REQ-SSS-016).

---

### B.6 Distribution and disclosure

**AC-SSS-016 — mirrors, pairs, neutrality, and the PR declaration.**
Given the branch's diff against its merge base,
When each changed path under `.claude/` is checked for a corresponding changed path under
`internal/template/templates/.claude/`, each changed hook wrapper is checked for its `.sh` /
`.sh.tmpl` sibling, and the changed template files are checked for forbidden content,
Then every `.claude/` change has a mirror, every wrapper pair moved together, and no template
file introduces a SPEC ID, a REQ token, a date, a commit SHA, an absolute macOS path, or a
`CLAUDE.local` reference.
Commands: `git diff --name-only <merge-base>...HEAD` (the mirror and pair audit reads this list);
`for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do b=${f%.tmpl}; diff -q "$b" "$f" >/dev/null || echo "DRIFT $b"; done` → must print nothing;
`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1`.
And: `gh pr view --json title,body --jq '.title, (.body | split("\n")[0])'` → the title and the
first body line must each name this as a change to the **security scan surface**, and neither may
present the change as a performance improvement (REQ-SSS-020). This is checked by reading the
command's output, not by a source grep, and is the last criterion closed because the PR does not
exist until M4.

---

## §C Edge cases the criteria deliberately cover

| Edge case | Covered by |
|---|---|
| Config present but malformed | AC-SSS-007 (escalate, never skip) |
| Language covered by config but rule shape unparseable | AC-SSS-009, AC-SSS-010 |
| Payload in a covered language with only `warning`-severity matches | AC-SSS-001 fixture set (allow before and after) |
| Payload in an ast-grep-supported but rule-uncovered language | AC-SSS-005 |
| Guardian finding and LSP advisory on the same event | AC-SSS-012 |
| A user project whose `settings.json` still names `handle-security-scan.sh` | AC-SSS-015 (subcommand retained, wrapper retained) |

## §D Explicitly not asserted

- No criterion asserts a wall-clock duration, a percentage speedup, or a millisecond figure. The
  audit's absolute numbers were taken at load 8-10 and the investigation at load 38; they are
  ratios, not guarantees.
- No criterion asserts that the pre-write deny has ever fired for a real user. Investigation
  Claim 1 establishes zero firings **on this machine only**; distributed-user behaviour is
  outside the observation window and is not claimed in either direction.
- No criterion asserts that the pre-filter reduces `sg` spawns for Go payloads. `spec.md` §C.2
  records why it usually will not.
