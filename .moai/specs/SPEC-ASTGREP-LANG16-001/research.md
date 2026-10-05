# SPEC-ASTGREP-LANG16-001 — Research

> Codebase and tool analysis behind the SPEC. Measurement tree per `spec.md` §A.0:
> worktree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`, base `294b4b6ab`,
> tool `ast-grep 0.40.5` (`/opt/homebrew/bin/sg`).
>
> The primary measured record is `.moai/reports/t228/plan-measurements.md` (M1-M5). This
> document records what was found beyond it, and the two corrections to it.

## §1. Provenance

`.moai/reports/t228/plan-measurements.md` supplied M1 (parser probe), M2 (rule inventory),
M3 (config mode works), M4 (`sg test` catches the inert-rule mutant), M5 (corpus availability).
M1-M4 are not re-derived here.

Three findings from this plan-phase pass amend it, and **M5 is superseded outright**: it reported
the differential corpus as reachable only from an unmerged lane. The corpus is committed on
`origin/main` (§7, `spec.md` §A.6); §8 records the measurement error that produced the M5 claim.

## §2. Correction — F8 severity

`plan-measurements.md` M2 lists `sec-template-injection-html` at `warning`.

Measured (`$T/security/web.yml`, `awk` over id/language/severity triples):

```
security/web.yml   sec-template-injection-html    go   error
security/web.yml   sec-log-injection-unsanitized  go   warning
security/web.yml   sec-csrf-no-token-check        go   warning
```

The tree says `error`. The measured value governs; `spec.md` §A.3 carries the superseding note
and AC-A16-018 checks the two agree at close.

The full triple enumeration also confirms M2's totals: 26 rules, of which 12 non-security Go
rules all sit at `warning`, and 14 security cells of which 12 are `error` and 2 are `warning`.

## §3. New finding — the severity-to-deny wiring

The t227 severity principle ("`error` means a write is refused") had not been located in code.
It is in `internal/hook/pre_tool.go`, on the write-scan path:

```
result, err := h.scanner.ScanFile(ctx, tmpFile.Name(), h.projectRoot())
...
if h.scanner.ShouldAlert(result) {
    report := h.scanner.GetReport(result, filePath)
    reason := fmt.Sprintf("Security vulnerabilities detected in %s:\n%s", ...)
    return DecisionDeny, reason
}
// warning-count-only results fall through to allow
```

Cited by the `ShouldAlert` → `DecisionDeny` token pair rather than a line number, which drifts
between trees.

Consequence for this SPEC: breadth and precision are one question, not two. Every rule promoted
to `error` refuses writes on everything it matches, which is why `spec.md` REQ-A16-012 carries a
precision clause — clause 2, the benign same-shape `valid` case — rather than a bare
"security ⇒ error" rule.

## §4. New finding — `sec-csrf-no-token-check` is a shape matcher

```yaml
id: sec-csrf-no-token-check
language: go
severity: warning
rule:
  pattern: "func $HANDLER(w http.ResponseWriter, r *http.Request) {\n  $$$BODY\n}\n"
```

The pattern is the shape of every Go HTTP handler. It carries no CSRF-specific token, so a
handler with correct CSRF middleware matches identically to one without. At `warning` it is a
noisy-but-harmless advisory; at `error` it would refuse writes to every Go HTTP handler in the
project.

This is the concrete instance that motivates the two-clause severity predicate, and it is why a
blanket "all security families → error" reclassification would have been a defect rather than an
improvement.

## §5. Existing conventions worth preserving

**The equal-opportunity idiom.** `$T/sgconfig.yml` already carries the wording the `r` /
`flutter` exclusion must inherit:

> Language coverage is intentionally partial and treated with equal opportunity: a language that
> does not yet appear in a rule directory is an equal-priority future addition, never an
> unsupported one.

**Family-grouped security files.** `credentials.yml` / `crypto.yml` / `injection.yml` /
`secrets.yml` / `web.yml` group by family, with language variants adjacent inside one file. The
existing `injection.yml` shows the pattern: four language variants of one family, each with its
own `note:` phrased for that language's ecosystem, and a header comment stating the
equal-opportunity intent.

**Repeated ids across languages.** `sec-hardcoded-credential` appears four times with four
`language:` values. Whether `sg test` tolerates this is the open R1 question settled at M1.

**Metadata fields.** Existing security rules carry `metadata.owasp` and `metadata.cwe`. New rules
should follow, so a finding is traceable to a classification rather than only to a rule id.

## §6. Predecessor and sibling SPECs

- **SPEC-ASTGREP-MULTILANG-001** (completed, v3.0.0) — landed the 26-rule curated baseline this
  SPEC extends. Its `module:` is the same template directory. It established the
  vetted-against-positive-and-negative-fixture convention that `sg test` now mechanizes.
- **SPEC-ASTG-UPGRADE-001** — **archived** (`status: archived`, measured). It owns no active
  work, so the pinned-version question it would otherwise have carried is owned by R6 in this
  SPEC. Listed only to record that this SPEC does not depend on it.
- **SPEC-ASTGREP-DOGFOOD-CLEANUP-001** (completed) — owns `.moai/astgrep-rules/`, the local
  dogfood tree this SPEC must not mirror.

No SPEC currently owns the language-coverage gap; this one does.

## §7. Constraints found in the surrounding tooling

- **`.github/workflows/template-neutrality-check.yaml` exists** and triggers on changes under
  `internal/template/templates/**`. It is the CI enforcement behind REQ-A16-019.
- **`gate.yaml` `ast_grep_gate` ships enabled in advisory mode** (`block_on_error: false`,
  `warn_only_mode: true`), with blocking opt-in. This SPEC changes none of it — but it means a
  severity promotion affects the PreToolUse write-scan path (§3) independently of the commit-time
  gate's advisory posture. The two surfaces are separate, and only the first refuses writes.
- **`internal/hook/security/testdata/scan-corpus/` is present on `origin/main`** — 12 committed
  files landed by `a9eb896ce` (PR #1637, card t227), with the consuming harness
  `internal/hook/pre_tool_scan_differential_test.go`. Two earlier plan-phase records reported it
  absent; §8 records why, and `spec.md` §A.6 carries the four confirmations. The t217 lane
  inherits these files unchanged and modifies nothing under that path, so no merge gate and no
  path collision exist. **This SPEC does not touch the corpus at all** — it is read to state what
  the harness enforces (§9), and extended by `SPEC-ASTGREP-BREADTH-001`.

## §8. Measurement hygiene — an absence claim this SPEC got wrong twice

This is not generic advice. It is an incident from this SPEC's own authoring, recorded because it
cost a full round of scoping here and a second one in the lead session the same day.

**What happened.** The differential corpus was twice reported as existing on no ref, and this
SPEC was scoped around that absence — first deferring the corpus work to a successor card, then
planning to create the directory from scratch. Both scopings were wrong. The corpus had been
committed to `main` by `a9eb896ce` throughout.

**The mechanism.** `git ls-tree -r --name-only <ref>` scopes its listing to the **current
directory prefix**. The measuring shell's cwd had drifted into
`internal/template/templates/.moai/config/astgrep-rules/security` — a directory containing no
`internal/hook/...` — so the listing came back empty and `grep -c` printed `0` with rc=1. The
same drift emptied `git log <ref> -- '*scan-corpus*'`, because a glob pathspec also resolves
against cwd.

**Why it is dangerous.** The trap fires **only on absence claims**, and it fails **silently** —
no error, no warning, no diagnostic. A present-claim measured this way still shows the file, so
the technique looks reliable right up until the one question where it is not. And an empty
result is exactly what an absence claim is looking for, so it confirms rather than contradicts.

**The rule.** Measure ref contents from the repository root, or pass `--full-tree`. Use
repo-relative pathspecs (`-- internal/hook/security/testdata/scan-corpus`) rather than globs
(`-- '*scan-corpus*'`). Treat rc=1-with-empty-output as **unproven**, not as proof. State the
directory a measurement was run from alongside the command, exactly as `spec.md` §A.0 requires
the tree to be stated.

**What resolved it.** Two confirmations independent of the failing technique:
`git merge-base --is-ancestor a9eb896ce HEAD` (exit 0), and `git ls-files` + `git status
--porcelain` on the path showing 12 tracked, clean files. A path cannot be absent from every ref
and simultaneously tracked-and-clean in a worktree checked out from one.

## §9. Two defenses falsified during plan-audit

Recorded because both were load-bearing, and a dropped claim teaches nothing.

**The corpus validity gate is a skip, not a failure.** Three revisions of this SPEC described
`coveredCorpusLanguages` as a forcing function that turns the suite red when a covered language
lacks a denying fixture. Measured at `internal/hook/pre_tool_scan_differential_test.go:242`, the
gate calls `t.Skip` — and because it sits above the assertion loop at line 245, it disables all
twelve differential assertions at once while `go test` prints `ok`.

The skip message names the root cause: `astGrepScanner.Scan` uses `CombinedOutput`, and
`sg scan --json` writes `Error: N error(s) found in code.` to stderr on any error-severity
finding, corrupting the JSON so no error finding ever parses — *"The pre-write gate can warn but
cannot deny."* A live defect in the tree, excluded from repair by `spec.md` §D, and the reason
REQ-A16-016 requires enforcement claims to match measured behaviour.

**A contrived rule passes every gate.** R2's mitigation held that "a contrived pattern has no
plausible benign counterpart to write". A rule matching `zzzNeverRealApi($X)`, paired with a
`valid` case calling `zzzSafeApi(x)` — neither API existing — passed `sg test` at exit 0 and would
qualify for `severity: error`. The benign counterpart took one line.

Both share a shape: **a defense that reads convincingly in prose and is never executed.** That is
what the contract half of this work exists to prevent for the 80 rules to come, and it is why the
`metadata.cwe` anchor (REQ-A16-011) is stated as necessary-not-sufficient rather than as a
closure — the honest reach of an anchor a reviewer checks against a public catalogue is a raised
floor, not a proof.

## §10. Open questions carried into run-phase

| # | Question | Settled at |
|---|---|---|
| Q1 | Does `sg test` key snapshots by rule id alone or by id+language? | M1 |
| Q2 | Which (family, language) cells have no equivalent construct? | `SPEC-ASTGREP-BREADTH-001`, per cell, with evidence. Unknown at this SPEC's plan time and explicitly not estimated here — the feasibility probe covered 6 patterns of 80. |
| Q3 | Does `sec-log-injection-unsanitized` satisfy the precision predicate, or is it a second shape matcher? | M3 |
| Q4 | Does `sg test` key snapshots by rule id alone or id+language? | M1 — settled before any rule is authored, because the id-alone branch renames four shipped rules. |
