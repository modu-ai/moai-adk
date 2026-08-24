# Implementation Plan — SPEC-SEC-SCAN-SURFACE-001

Ordering principle: the decisions most likely to change come first. M1 fixes the shape of the
gate and the shape of the config-derived language set — the two decisions a reviewer is most
likely to push back on. M2 carries the one genuinely uncertain piece (pre-filter derivation).
M3 is the mechanical merge, deferred to last because its design is already settled by an
existing in-tree merge helper.

---

## §A Context

Evidence base: `.moai/reports/t217/investigation.md`. Requirements: `spec.md` §B. The
pre-filter-derivation decision and its rejected alternative are recorded in `spec.md` §C and are
not re-opened during implementation.

Two card premises were corrected during authoring (`spec.md` §A.1): the advisory collision is
`systemMessage` against `additionalContext`, not `additionalContext` against itself; and
`moai hook security-scan` has no third-party caller, so its retention is a compatibility choice.

---

## §B Known issues in the current surface

| # | Issue | Location |
|---|---|---|
| B-1 | Config is resolved twice per Write — once implicitly inside `SecurityScanner.ScanFile`, and never in the caller, so the caller cannot short-circuit on an empty result. | `internal/hook/security/scanner.go` `ScanFile` |
| B-2 | The temp file is created **before** anything checks whether a scan can produce a finding. | `internal/hook/pre_tool.go` `scanWriteContent` |
| B-3 | `IsSupportedExtension` is the only gate between a Write and an `sg` spawn; it answers "can ast-grep parse this?", not "do we have rules for this?". | `internal/hook/security/ast_grep.go` |
| B-4 | `slog` output on the `moai hook` path is discarded (`internal/cli/logging.go`), so the existing `slog.Warn`/`slog.Info` lines in `scanWriteContent` are not an observation channel. Do not add evidence obligations that depend on them. | `internal/cli/logging.go` |

---

## §C Pre-flight

- Confirm the tree still matches the investigation's premises: `.moai/config/astgrep-rules` is
  absent in this worktree, and `sg` is on PATH.
- Confirm `mergeHandlerOutput` still accumulates `additionalContext` and `systemMessage`
  (`internal/hook/registry.go`), and that
  `TestDispatch_MergeAccumulatesAdditionalContext` still passes on the untouched tree. M3's
  design depends on it.
- Copy the shipped ruleset (`internal/template/templates/.moai/config/astgrep-rules/`) into a
  `t.TempDir()` project root for every test that needs a resolvable config. Never point a test
  at the developer's real `.moai/config/astgrep-rules/`, which is local-only and dogfood-grade.

---

## §D Constraints

- **The deny verdict is the invariant.** No milestone may land without the differential test of
  M0 passing (§F).
- **No latency assertion anywhere.** Cost is measured as counts: temp files created, `sg`
  processes spawned. The scanner's `sg` call is injectable (`astGrepScanner.scanFunc`), so the
  count is observable without timing anything.
- **Fail-open on every uncertainty.** Unreadable config, unparseable rule, unrecognized rule
  shape ⇒ escalate to `sg`. Never skip.
- **Template-First.** Every `.claude/` edit lands with its `internal/template/templates/.claude/`
  mirror in the same commit; hook wrappers move as `.sh` + `.sh.tmpl` pairs.
- **Test scope.** Run `go test ./internal/hook/...` and `go test ./internal/cli/...`; do not run
  the full suite locally. CI on the pull request is the verdict.

---

## §E Self-verification

Before declaring any milestone complete, the implementer states the command run and its verbatim
output for each acceptance criterion the milestone closes (`acceptance.md`). A criterion with no
cited command output is a gap, not a pass.

---

## §F Milestones

### M0 — Priority High — Pin the invariant before changing anything

The differential harness is written **first**, against the unmodified gate, so it records real
behaviour rather than the behaviour the refactor produces.

1. Build a fixture corpus under `internal/hook/security/testdata/`: for each of the four covered
   languages, at least one payload that trips an `error`-severity rule (deny expected) and one
   clean payload (allow expected); plus payloads for two uncovered languages and one payload in
   a covered language whose content trips only a `warning` rule.
2. Add a table test that runs each fixture through `scanWriteContent` against a temp project root
   carrying the shipped ruleset, and records `(decision, reason-nonempty)`.
3. Land the recorded expectations as the test's assertions. This test must pass unmodified after
   M1, M2, and M3.

Deliverable: `internal/hook/pre_tool_scan_differential_test.go`. Closes AC-SSS-001.

### M1 — Priority High — A2 + A3: resolve once, skip when there is nothing to find

Reversible decisions here: where config resolution moves to, and what shape the config-derived
language set takes. Both are visible in the public surface of `internal/hook/security`.

1. Give `RuleManager` (or a small sibling in `rules.go`) a method that returns, for a project
   root, the resolved config path **and** the set of languages for which the resolved
   configuration declares at least one rule — walking `ruleDirs` and reading each rule file's
   `language:` field. Return a distinguishable "unknown" when anything fails to read or parse.
2. Reorder `scanWriteContent`: extension check → config resolution → covered-language check →
   (M2's pre-filter) → temp file → scan. Return `("", "")` before the temp file on every skip.
3. Pass the already-resolved config path into the scanner rather than letting `ScanFile` resolve
   it again (B-1).
4. Cache the derived language set for the process lifetime. The hook process is short-lived, so
   this is a single load per invocation, not a long-lived cache with invalidation concerns.

Closes AC-SSS-002, AC-SSS-003, AC-SSS-004, AC-SSS-005.

### M2 — Priority Medium — A1: literal pre-filter derived from the error-severity rules

The one genuinely uncertain milestone. Its correctness rests on the extractor being conservative.

1. New `internal/hook/security/prefilter.go`: a pure function from a parsed rule set to
   `map[language]prefilter`, where a `prefilter` is either a set of literal tokens or the
   `underivable` marker.
2. Extraction rules per `spec.md` §C.2 — mandatory literal runs outside metavariables, mandatory
   literal prefixes of `regex:` constraints, `all:` union, `any:` only when every branch yields a
   token, and `underivable` for every shape not confidently understood.
3. Wire it into `scanWriteContent` after the covered-language check: no token present in the
   payload ⇒ allow without temp file or spawn.
4. Unit-test the extractor directly against the shipped ruleset's rule shapes, including the
   negative cases: a `regex:`-only rule with no literal anchor, an `any:` with one tokenless
   branch, and an `inside:` rule must each mark the language underivable.

Closes AC-SSS-006, AC-SSS-007, AC-SSS-008, AC-SSS-009.

### M3 — Priority Medium — B: fold the guardian into the post-tool handler

Design already settled by `mergeHandlerOutput`; this is the mechanical milestone.

1. Add a PostToolUse handler that extracts the written content from `HookInput.ToolInput`, runs
   `ScanBuffer`, and returns a `HookOutput` carrying the guardian banner + findings on
   `hookSpecificOutput.additionalContext` — the same field and the same text the standalone
   handler emits today. Register it alongside the existing post-tool handler in
   `internal/cli/deps.go` so the registry's existing accumulation does the merge.
2. Regression test: dispatch one `Write` event through the registry with both handlers
   registered, with the post-tool handler producing a non-empty `systemMessage` and the guardian
   producing a non-empty `additionalContext`; assert **both** strings survive in the single
   emitted output. This is the test the card's [HARD] constraint names.
3. Remove the `handle-security-scan.sh` entry from the `Write|Edit|MultiEdit` matcher in
   `.claude/settings.json` **and** `internal/template/templates/.claude/settings.json.tmpl`.
   Leave the `handle-security-scan.sh` / `.sh.tmpl` wrapper pair in place — the subcommand it
   fronts is retained (REQ-SSS-016), and a user project whose settings still reference it must
   not hit the missing-hook log path.
4. Confirm the guardian handler introduces no deny and no non-zero exit.

Closes AC-SSS-010, AC-SSS-011, AC-SSS-012, AC-SSS-013.

### M4 — Priority Low — Distribution and disclosure

1. Verify every touched `.claude/` file has its template mirror changed in the same commit, and
   that no touched hook wrapper's `.sh` / `.sh.tmpl` pair has drifted.
2. `make build` to re-embed templates.
3. Template neutrality self-check over the changed template files.
4. Open the pull request with the security-surface declaration in the title and the first body
   sentence (REQ-SSS-020).

Closes AC-SSS-014, AC-SSS-015, AC-SSS-016.

---

## §G Anti-patterns

- **Deriving the pre-filter from `GuardianPatterns()`** because it is already compiled. Rejected
  with named consequences in `spec.md` §C.1.
- **Treating an unreadable config as "no rules"** and skipping. Absence of evidence is not
  evidence of absence; `spec.md` REQ-SSS-007 forbids it.
- **Asserting a latency improvement.** The available numbers are load-contaminated ratios. Assert
  counts.
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

- `.moai/reports/t217/investigation.md` — the evidence base for every claim in §A of `spec.md`.
- `internal/hook/registry.go` § `mergeHandlerOutput` — the merge semantics M3 relies on.
- `internal/hook/security/guardian.go` § `HandleSecurityScan` — the advisory contract M3 preserves.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the evidence obligation §E encodes.
