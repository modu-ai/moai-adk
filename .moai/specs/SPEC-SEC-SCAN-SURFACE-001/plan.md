# Implementation Plan — SPEC-SEC-SCAN-SURFACE-001

Ordering principle: the decisions most likely to change come first. M0 pins the invariant and
builds the measurement instrument the rest of the plan depends on. M1 fixes the shape of the gate
and of the config-derived language set. M2 carries the one genuinely uncertain piece (pre-filter
derivation). M3 is the mechanical merge, deferred to last because its design is already settled
by an existing in-tree merge helper.

---

## §A Context

Evidence base: `.moai/reports/t217/investigation.md`, plus the measurements in `spec.md` §A.1.
Requirements: `spec.md` §B. The pre-filter-derivation decision, its extraction table, its
measured skip rates, and its rejected alternative are recorded in `spec.md` §C and are not
re-opened during implementation.

Two card premises were corrected during authoring (`spec.md` §A.2). One auditor-suggested remedy
was also corrected — see B-5 below.

---

## §B Known issues in the current surface

| # | Issue | Location | Consequence for the plan |
|---|---|---|---|
| B-1 | Config is resolved inside `SecurityScanner.ScanFile`, not in the caller, so the caller cannot short-circuit on an empty result. Measured: exactly one resolution today, at `scanner.go:84`. | `internal/hook/security/scanner.go:84` | M1 moves it to the caller and removes the scanner-side call |
| B-2 | The temp file is created **before** anything checks whether a scan can produce a finding. | `internal/hook/pre_tool.go` `scanWriteContent` | M1 reorders |
| B-3 | `IsSupportedExtension` is the only gate between a Write and an `sg` spawn; it answers "can ast-grep parse this?", not "do we have rules for this?". | `internal/hook/security/ast_grep.go` | M1 adds the covered-language gate |
| B-4 | `slog` output on the `moai hook` path is discarded (`internal/cli/logging.go`), so the existing `slog.Warn` / `slog.Info` lines in `scanWriteContent` are not an observation channel. | `internal/cli/logging.go` | No acceptance criterion may depend on log output |
| B-5 | **No injection seam exists on the single-file scan path.** `astGrepScanner.scanFunc` is consulted only inside `ScanMultiple` (`ast_grep.go:199-200`); the single-file `Scan` execs `sg` directly at `:137`. The plan audit proposed injecting a stub at the handler level instead — but `preToolHandler.scanner` is a **concrete** `*security.SecurityScanner` (`pre_tool.go:325`), so that is not possible either without a type change. | `internal/hook/security/ast_grep.go`, `internal/hook/pre_tool.go:325` | **M0 owns creating the seam** — see M0 step 1 |

---

## §C Pre-flight

- Confirm the tree still matches the premises: `.moai/config/astgrep-rules` absent in this
  worktree, `sg` on PATH.
- Confirm `mergeHandlerOutput` still accumulates `additionalContext` and `systemMessage`
  (`internal/hook/registry.go:180`), and that `TestDispatch_MergeAccumulatesAdditionalContext`
  passes on the untouched tree. M3's design depends on it.
- Copy the shipped ruleset (`internal/template/templates/.moai/config/astgrep-rules/`) into a
  `t.TempDir()` project root for every test that needs a resolvable config. Never point a test
  at the developer's real `.moai/config/astgrep-rules/`, which is local-only and dogfood-grade.

---

## §D Constraints

- **The deny verdict is the invariant.** No milestone may land without M0's differential test
  passing, including its second assertion (no denying fixture is suppressed by the pre-filter).
- **No latency assertion anywhere.** Cost is measured as counts. The instrument is defined in M0
  and is the **number of `ScanFile` calls reaching the scanner**, plus the presence or absence of
  a temp file. `ScanFile` is the sole route from this path to an `sg` spawn, so zero `ScanFile`
  calls proves zero spawns; the proxy is exact for every skip case an acceptance criterion
  asserts.
- **Fail-open on every uncertainty.** Unreadable config, unparseable rule, unrecognized rule
  shape, empty derived language set ⇒ escalate to `sg`. Never skip.
- **Template-First.** Every `.claude/` edit lands with its `internal/template/templates/.claude/`
  mirror in the same commit. This SPEC changes `settings.json` only; it touches no hook wrapper.
- **Test scope.** Run `go test ./internal/hook/...` and `go test ./internal/cli/...`; do not run
  the full suite locally. CI on the pull request is the verdict.

---

## §E Self-verification

Before declaring any milestone complete, the implementer states the command run and its verbatim
output for each acceptance criterion the milestone closes (`acceptance.md`). A criterion with no
cited command output is a gap, not a pass.

---

## §F Milestones

### M0 — Priority High — Build the instrument, then pin the invariant

Nothing below can be measured until the seam exists (B-5), and nothing can be safely changed
until the invariant is recorded against the untouched gate.

1. **Create the measurement seam.** Introduce a narrow interface in `internal/hook` describing
   only what `scanWriteContent` uses of the scanner (`IsAvailable`, `ScanFile`, `ShouldAlert`,
   `GetReport`) and change `preToolHandler.scanner` from `*security.SecurityScanner` to that
   interface. `NewPreToolHandlerWithScanner` keeps its concrete parameter type so no caller
   changes. This is a type-narrowing refactor with no behaviour change, and it makes a counting
   fake injectable.
2. Build the fixture corpus under `internal/hook/security/testdata/`: for each of the four
   covered languages, at least one payload that trips an `error`-severity rule (deny expected)
   and one clean payload (allow expected); plus payloads for two uncovered languages and one
   payload in a covered language whose content trips only a `warning` rule.
3. Add the differential table test running each fixture through `scanWriteContent` against a
   temp project root carrying the shipped ruleset, recording `(decision, reason-nonempty)`.
   **The corpus is rejected unless at least one fixture per covered language denies on the
   pre-implementation tree** — a corpus that denies nothing observes nothing.
4. Land the recorded expectations as assertions. This test must pass unmodified after M1, M2,
   and M3, and gains its second assertion in M2.

Deliverable: `internal/hook/pre_tool_scan_differential_test.go`. Closes AC-SSS-001.

### M1 — Priority High — A2 + A3: resolve once, skip when there is nothing to find

Reversible decisions here: where config resolution moves to, and what shape the config-derived
language set takes. Both are visible in the public surface of `internal/hook/security`.

1. Give `RuleManager` a method returning, for a project root, the resolved config path **and**
   the set of languages for which the resolved configuration declares at least one rule — walking
   `ruleDirs` and reading each rule file's `language:` field. Return a distinguishable "unknown"
   when anything fails to read or parse, **and when the derived set is empty**.
2. Reorder `scanWriteContent`: extension check → config resolution → covered-language check →
   (M2's pre-filter) → temp file → scan. Return `("", "")` before the temp file on every skip.
3. Pass the already-resolved config path into the scanner and **remove the scanner-side
   resolution on this path** (B-1), so `ScanFile` performs zero resolutions when given a path.
4. Cache the derived language set for the process lifetime. The hook process is short-lived, so
   this is a single load per invocation, not a long-lived cache with invalidation concerns.

Closes AC-SSS-002, AC-SSS-003, AC-SSS-004, AC-SSS-005, AC-SSS-006, AC-SSS-007.

### M2 — Priority Medium — A1: literal pre-filter derived from the error-severity rules

The one genuinely uncertain milestone. Its correctness rests on the extractor being conservative,
and `spec.md` §C.2's table is its specification — implement that table, not an interpretation.

1. New `internal/hook/security/prefilter.go`: a pure function from a parsed rule set to
   `map[language]prefilter`, where a `prefilter` is either a set of literal tokens or the
   `underivable` marker.
2. Implement every row of the §C.2 extraction table, including the two rows the plan audit added:
   regex top-level alternation (union of per-branch mandatory prefixes) and `kind:` + `regex:`
   (derive from the regex conjunct). The shipped ruleset's dominant rule
   (`sec-hardcoded-credential`, present in all four covered languages) exercises both.
3. Wire it into `scanWriteContent` after the covered-language check.
4. Unit-test the extractor against the shipped ruleset's real rule shapes, plus the negative
   cases: a `regex:` with no literal anchor, an `any:` with one tokenless branch, and an
   `inside:` composite must each mark the language underivable.
5. Add the second assertion to M0's differential test: for every fixture that denies, assert the
   derived pre-filter would **not** have skipped it.

Closes AC-SSS-008, AC-SSS-009, AC-SSS-010, AC-SSS-011.

### M3 — Priority Medium — B: fold the guardian into the post-tool handler

Design already settled by `mergeHandlerOutput`; this is the mechanical milestone.

1. Add a PostToolUse handler that extracts the written content from `HookInput.ToolInput`, runs
   `ScanBuffer`, and returns a `HookOutput` carrying the guardian banner + findings on
   `hookSpecificOutput.additionalContext` — the same field and the same text the standalone
   handler emits today. Register it alongside the existing post-tool handler in
   `internal/cli/deps.go` so the registry's existing accumulation does the merge.
2. Regression test: dispatch one `Write` event through the registry with both handlers
   registered, the post-tool handler producing a non-empty `systemMessage` and the guardian a
   non-empty `additionalContext`; assert **both** strings survive. This is the test the card's
   [HARD] constraint names.
3. Reachability: register the guardian handler such that a preceding handler's decision cannot
   short-circuit it (`spec.md` §A.3), and assert reachability in the same test by having the
   first-registered handler return a block decision.
4. Remove the `handle-security-scan.sh` entry from the `Write|Edit|MultiEdit` matcher in
   `.claude/settings.json` **and** `internal/template/templates/.claude/settings.json.tmpl`.
   Leave the `handle-security-scan.sh` / `.sh.tmpl` wrapper pair in place — the subcommand it
   fronts is retained (REQ-SSS-013), and a user project whose settings still reference it must
   not hit the missing-hook log path.
5. Confirm the guardian handler introduces no deny and no non-zero exit.

Closes AC-SSS-012, AC-SSS-013, AC-SSS-014, AC-SSS-015.

### M4 — Priority Low — Distribution and disclosure

1. Verify every `.claude/` file in this branch's diff has its template mirror in the same commit,
   and that any hook wrapper in the diff moved as a `.sh` / `.sh.tmpl` pair. Scope the pair check
   to the diff, not to the repository: `handle-pre-tool.sh` already differs from its template in
   comments for a legitimate reason (`spec.md` §D), so a repository-wide byte-equality sweep is
   not a valid gate.
2. `make build` to re-embed templates.
3. Template neutrality self-check over the changed template files.
4. Open the pull request with the security-surface declaration in the title and the first body
   sentence (REQ-SSS-016).

Closes AC-SSS-016.

---

## §G Anti-patterns

- **Deriving the pre-filter from `GuardianPatterns()`** because it is already compiled. Rejected
  with named consequences in `spec.md` §C.1.
- **Reading `kind:` + `regex:` as "kind-only, therefore underivable"**, which silently reduces
  item A1 to a no-op in all four covered languages. `spec.md` §C.2 resolves this explicitly.
- **Treating an unreadable or empty config as "no rules"** and skipping. Absence of evidence is
  not evidence of absence; REQ-SSS-006 forbids it.
- **Asserting a latency improvement.** The available numbers are load-contaminated ratios. Assert
  counts.
- **Measuring cost through a seam that does not exist.** `scanFunc` is `ScanMultiple`-only
  (B-5); M0 step 1 creates the seam before any acceptance criterion relies on it.
- **Folding the guardian text into `systemMessage`** because it is the field the post-tool
  handler already uses. That is exactly the drop the card's [HARD] constraint exists to prevent;
  the guardian's carrier field is `additionalContext` and stays that way.
- **Deleting the `handle-security-scan.sh` wrapper** along with the settings entry. The
  subcommand is retained; an orphaned settings reference in a user project must still find a
  wrapper.
- **Extending scope to the post-write ast-grep pass** because it has the same waste. Out of
  scope by `spec.md` §D.

---

## §H Cross-references

- `.moai/reports/t217/investigation.md` — the evidence base for `spec.md` §A.
- `.moai/reports/t217/plan-audit.md` — audit iteration 1; D1-D8 remediation is recorded in
  `spec.md` HISTORY and in the corrected criteria of `acceptance.md`.
- `.moai/reports/t217/skiprate.py` — the §C.3 skip-rate measurement, re-runnable.
- `internal/hook/registry.go` § `mergeHandlerOutput` — the merge semantics M3 relies on.
- `internal/hook/security/guardian.go` § `HandleSecurityScan` — the advisory contract M3 preserves.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the evidence obligation §E encodes.
