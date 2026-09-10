# SPEC-ASTGREP-LANG16-001 — Design

> Design decisions and artifact schemas. Requirements live in `spec.md`; verification in
> `acceptance.md`. Measurement tree per `spec.md` §A.0.

## §1. The coverage matrix

### 1.1 Why a document rather than a generated report

The matrix could be derived from the rule files: walk the tree, collect (id, language) pairs,
print what is missing. That derivation answers "which cells have rules" but cannot answer "which
missing cells are deliberate", because the deliberateness lives nowhere in the rule files. The
matrix is therefore an authored document — the rationale for each absence is content a human
wrote, and the checker's job is to verify the document is complete and evidenced, not to
generate it.

The consequence worth stating: the matrix can drift from the rule tree. `sg test` cannot close
that — it never reads the matrix. The checker's fourth failure class (§5, dangling rule id) is
what closes it, by resolving every cell's named id against the shipped ruleset.

### 1.2 Axes

**Family axis (8, fixed).** F1 command injection, F2 hardcoded credential, F3 weak hash,
F4 hardcoded API key, F5 hardcoded JWT signing key, F6 CSRF token absence, F7 log injection,
F8 template injection / XSS. Fixed so the matrix has a stable shape; new families are out of
scope per `spec.md` §D.

**Language axis (14 parseable).** The four already covered — go, javascript, python, typescript —
plus the ten reachable under ast-grep 0.40.5: rust, java, kotlin, csharp, ruby, php, elixir,
cpp, scala, swift.

The four already-covered languages are **on the axis**, not exempt from it. Their gaps are real
gaps: javascript has no weak-hash rule today, python has no JWT rule, and neither absence is
recorded anywhere. Placing them on the axis means those 18 pre-existing gaps get the same
treatment as the 80 new cells — implemented, or exempted with evidence. Excluding them would
have made the matrix cheaper and dishonest.

`r` and `flutter` are off the language axis entirely and live in a separate excluded-languages
record. They contribute no cells: a cell for a language with no parser could never be
implemented, so carrying 16 exemptions for them would inflate the exemption count with a fact
already stated once.

### 1.3 Cell schema

```markdown
| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F2 | rust    | IMPLEMENTED | sec-hardcoded-credential-rust | `$R`/rust/credentials/ |
| F6 | cpp     | EXEMPT      | no web-framework request surface in scope | probe: `sg run -p '...' -l cpp --stdin` → no match |
| F7 | elixir  | EXEMPT      | Logger macros take no user-format string | cite: Elixir Logger docs, "Message" section |
```

`state` is one of exactly `IMPLEMENTED` or `EXEMPT`. The `evidence` column is required for
`EXEMPT` and holds the rule-test path for `IMPLEMENTED`.

### 1.4 Why exemptions owe evidence

An `IMPLEMENTED` cell already carries its evidence structurally: the rule exists, and `sg test`
proves it fires on the dangerous construct and stays silent on the benign one. An `EXEMPT` cell,
without an evidence requirement, carries nothing — the rationale is a sentence asserting that a
construct does not exist.

That asymmetry is the hazard. A 112-cell matrix can reach full resolution by writing 98
plausible-sounding sentences, and the resulting document reports itself complete while
establishing almost nothing. The failure is quieter than a missing cell, because a missing cell
at least looks missing.

Requiring a citation or a probe puts both halves of the matrix on one standard: the presence
claim is proven by a test, the absence claim by a reference or a probe that returned nothing.
The checker enforces this as a **distinct failure class** (AC-A16-011) rather than folding it
into "unresolved cell", so a reviewer reading checker output can tell an unevidenced exemption
apart from a blank.

---

## §2. Severity predicate

### 2.1 The two clauses

A rule carries `severity: error` when both hold:

1. it belongs to a security family, and
2. its pattern names a **specific dangerous construct** — evidenced by a `valid` test case
   containing a benign construct of the same shape, on which the rule produces zero findings.

Everything else is `warning`.

### 2.2 Why clause 2 exists

`error` is not a label. On the measured tree, `internal/hook/pre_tool.go` calls
`scanner.ShouldAlert(result)` on the write-scan path and returns `DecisionDeny` when it is true;
warning-only results fall through to allow. A rule at `error` refuses `Write` and `Edit` on any
file it matches.

`sec-csrf-no-token-check` shows what clause 1 alone would do. Its pattern is
`func $HANDLER(w http.ResponseWriter, r *http.Request) { $$$BODY }` — the shape of every Go HTTP
handler, CSRF-protected or not. Promoted to `error` on the strength of "it is a security rule",
it would refuse every write to every Go HTTP handler in the project. A gate that refuses correct
code gets disabled, and then it protects nothing.

Clause 2 makes the distinction mechanical rather than a case-by-case judgement: if you cannot
write a benign same-shape example the rule stays silent on, the rule is matching shape, not
danger.

### 2.3 Reclassification moves in both directions

This is not a relaxation mechanism. A rule leaves `error` when it lacks the evidence and enters
`error` when it acquires it; the total may rise or fall. What is fixed is the predicate. A
security rule demoted under clause 2 keeps its finding — it still reports — it simply stops
refusing the write, and the matrix records why (`spec.md` REQ-A16-015) so the demotion is
visible rather than inferred from a severity field.

---

## §3. `sg test` harness

### 3.1 Wiring

`testConfigs` in the template `sgconfig.yml` points at the **repo-side rule-test root, outside the
distributed template tree** (REQ-A16-019, and §4 below for why). The `testDir` value is therefore a
relative path that traverses upward out of `$T`; fixtures and snapshots live under that repo-side
root, never under `$T/rule-tests/`. Snapshots are generated and committed so a later regression is
a diff rather than a re-derivation.

**The placement is measured feasible, not assumed.** Plan-audit reproduced it: `sg test` resolves a
`testConfigs` entry of the form `testDir: ../outside-tests` relative to the `sgconfig.yml` that
declares it, traverses upward out of the config's own directory, and exits 0. So the upward path is
a working configuration rather than a hoped-for one — which matters because the alternative
(fixtures inside `$T`) is not merely untidy: it ships ~90 deliberately-dangerous constructs into
every user project, inside the exact tree the deployed ruleset scans.

The residual risk of that choice is the deployed config carrying a `testDir` the user does not
have; AC-A16-006 is the criterion that holds it.

### 3.2 The mutant this catches

Measured in the M4 prototype: a rule rewritten to `NeverMatchesAnything::zzz("sh")` — still valid
YAML, still a valid pattern, matching nothing — produced
`[Missing] Expect rule probe-rust to report issues, but none found in: ...` and exited FAIL.

This is the failure mode that scales badly. Eighty rules written across ten unfamiliar languages,
verified only by "the file parses", would produce a ruleset that reports clean because it finds
nothing. `sg test` is the tool that distinguishes an inert rule from an absent one, and every
`IMPLEMENTED` cell depends on it.

### 3.3 Settled — rule-id keying (R1). Measured M1 (card t228), ast-grep 0.40.5

**Measured verdict: `sg test` keys each test case by the case `id` ALONE, globally — one snapshot
file `<id>-snapshot.yml` shared by every case carrying that id, independent of testDir or file
name. Two further engine facts shape the same record:**

1. A case whose id does not equal an existing rule configuration is silently dropped ("Configuration not found!", exit still 0).
2. Inline `sg test` snippets route through a single language selection: python-branch rules never observe their inline snippets even when their pattern matches byte-for-byte on real files (`sg scan --rule <doc> py_deny_os_system.py` fires; the identical snippet inside a case is `[Missing]`).

**Commands and observed outputs (scratch ruleset `dup.yml`: two docs, one id, languages go + python):**

```
$ sg test --config sgconfig.yml            # two case files, both id: dup-cred
[Wrong] No dup-cred baseline found.
...
FAIL dup-cred .W / FAIL dup-cred .W ; "test failed. 0 passed; 2 failed;"

$ sg test -U --config sgconfig.yml && cat tests/__snapshots__/dup-cred-snapshot.yml
...single file written...
id: dup-cred
snapshots:
  ? |
    y := "sk-proj-demo"
  : ...       # only ONE variant's entry survives; second write clobbers the first

$ sg test --config sgconfig.yml            # re-run after update
[Missing] Expect rule dup-cred ... "No dup-cred baseline found."  → 1 passed; 1 failed; rc=4

# distinct ids rename probe on one file:
Configuration not found! dup-cred-py   → runner drops the case silently, rc=0
```

**Decision: the shipped convention stays UNCHANGED — no renames.** Two reasons:

- The collision lives entirely in the CASE-id layer, which this SPEC authors. One union case per
  DISTINCT rule id covers every language branch that the snippet pipeline can reach; three shared
  families (`sec-hardcoded-credential`, `sec-command-injection-shell/exec`) collapse 26 rules into
  21 unique ids, so the harness reports **21 passing tests for 26 rules** — the derivable number
  under this keying, recorded here for AC-A16-001 rather than declared.
- The plan's first branch (per-language suffixed rule ids) was measured to be FORBIDDEN, not
  merely unnecessary: `internal/hook/security/prefilter_test.go` (PRESERVE-listed,
  `TestPrefilterKindPlusRegexAlternation`) pins the literal `sec-hardcoded-credential` document in
  all four covered languages; renaming any variant turns that guard red.

**Rule defects surfaced while making cases pass (all four were inert under 0.40.5; each fixed with
a structural pattern of the vetted kind+regex family, metadata untouched):
`sec-template-injection-html`, `go-defer-in-loop`, `sec-hardcoded-api-key`,
`sec-hardcoded-jwt-signing-key`.** Their harness coverage previously did not exist, which is why
no earlier gate ever saw these go silent.

---

## §4. Directory layout

```
$T/
├── sgconfig.yml       # ruleDirs: [go, security, idioms]; testConfigs testDir points OUTSIDE $T
├── go/                # existing — severity pass only
├── security/          # 8 family files, each gaining language variants
└── (idioms/ and per-language rules are the successor SPEC's)

<repo-side rule-tests root>    # NOT under the template tree — see below
└── <lang>/<family>/           # valid/invalid fixtures + snapshots
```

**Rule-test assets live outside the distributed tree (REQ-A16-019).** Everything under
`internal/template/templates/.moai/config/astgrep-rules/` is deployed into every user project. At
the successor's volume that would ship ~90 fixture files to every user — and by REQ-A16-008 each
fixture contains a deliberately dangerous construct, landing inside the very tree the deployed
ruleset scans, so the user's own pre-write gate would fire on our test assets. `testConfigs` is
therefore wired repo-side, and the deployed `sgconfig.yml` must stay valid where that path is
absent.

Security rules stay grouped by **family** (the existing `credentials.yml` / `crypto.yml` /
`injection.yml` / `secrets.yml` / `web.yml` split) rather than being regrouped by language.
Family grouping keeps the variants of one rule adjacent, which is where they are compared and
kept consistent; language grouping would scatter each family across ten files and make a
family-level change a ten-file edit.

`idioms/` is grouped by language instead, because an idiom rule is language-specific by nature
and has no cross-language family to sit in.

`ruleDirs` gains `idioms` only once that directory ships a passing rule (REQ-A16-023) — an empty
registered directory is a claim of coverage with nothing behind it.

---

## §5. Checker contract

Input: the matrix document. Output: exit code plus a per-failure report.

| exit | condition |
|---|---|
| 0 | key set equals the Cartesian product, every cell resolved, every exemption evidenced, every named rule id present |
| non-zero | any of the four classes below, each named separately in the output |

Failure classes:

1. **key-set mismatch** — the set of cell keys ≠ the Cartesian product of the two axes. A set
   comparison, deliberately not a count: a matrix with `F1/rust` duplicated and `F1/java` deleted
   has cardinality 112 and is missing a language. A substituted cell is worse than a missing one
   because it presents as complete, so a count-based class reports no failure at all.
2. **unresolved cell** — a cell carrying neither a rule id nor a rationale
3. **unevidenced exemption** — an `EXEMPT` cell whose evidence column carries neither a citation
   nor a probe record
4. **dangling rule id** — a cell naming a rule id absent from the shipped ruleset. Nothing else
   reads both the matrix and the ruleset, so without this class a cell naming a renamed or
   never-written rule passes every gate in the SPEC.

The four classes are reported separately rather than collapsed, so the output distinguishes
"someone forgot a row" from "someone substituted one" from "someone asserted an absence without
checking it" from "someone named a rule that does not exist". Those need different responses, and
a single generic error would hide the later ones inside the first.

The checker runs in the Go test suite rather than as a standalone script, so it executes in CI
on every push rather than when someone remembers.

---

## §6. Differential corpus — read, not extended, by this SPEC

This SPEC touches no corpus file. The section is retained because the SPEC must state accurately
what the corpus enforces (REQ-A16-016), and because `SPEC-ASTGREP-BREADTH-001` inherits that
statement as contract.

### 6.1 What exists

`internal/hook/security/testdata/scan-corpus/` holds twelve fixtures committed by `a9eb896ce`
(PR #1637, card t227), consumed by `internal/hook/pre_tool_scan_differential_test.go`. Two harness
structures matter:

- **`scanCorpus`** — `{name, file, virtualPath, coveredLanguage, wantDeny}` per fixture. Note
  `virtualPath`: the Write payload's claimed path selects the language by extension, not the
  fixture's own filename.
- **`coveredCorpusLanguages`** — currently `{go, javascript, typescript, python}`.

### 6.2 What it enforces, and what it does not

[CORRECTED — three prior revisions of this SPEC stated the opposite.]

| Mechanism | Behaviour | Real? |
|---|---|---|
| Per-row `wantDeny` mismatch | `t.Errorf` | **Yes** |
| Deny row with empty reason | `t.Errorf` | **Yes** |
| Allow row with non-empty reason | `t.Errorf` | **Yes** |
| Covered language with no denying fixture | **`t.Skip`** at line 242 | **No — it is a skip** |

The skip sits *after* the observation loop and *before* the assertion loop at line 245, so it
disables all twelve assertions at once and `go test` reports `ok`. `coveredCorpusLanguages` is
therefore an escape hatch, not a gate: adding a language to it without a denying fixture turns the
whole differential test green-by-skip.

### 6.3 Why both halves of a fixture pair must assert

A deny fixture alone proves a rule fires; it does not prove the rule discriminates. A pattern
matching every file satisfies the deny half perfectly — the same defect `sec-csrf-no-token-check`
exhibits (§2.2) and the same shape as the `sg test` `[Noisy]` mutant (§3.2), approached from the
other side.

The harness already encodes the pairing: `wantDeny: false` rows assert an allow **and** an empty
reason. So the successor needs no new mechanism, only the discipline of adding both rows — plus
REQ-A16-017's no-skip assertion, without which a green run proves nothing.

### 6.4 The `_uncovered` placeholders (successor's problem, stated here)

`java_uncovered.java` and `rs_uncovered.rs` sit at `coveredLanguage: ""`, `wantDeny: false`. Both
contain an API-key-shaped literal, so once java/rust gain `error`-severity rules in the successor
SPEC, those rows flip to deny against `wantDeny: false` and fail via `t.Errorf`.

That failure is real — but it comes from the **per-row differential assertion**, not from the
validity gate this SPEC previously credited. The distinction matters for the successor's other
eight languages, which have no such accidental literal and therefore no forcing function at all
beyond REQ-A16-017's explicit no-skip check.

## §7. Cross-references

- `spec.md` §A.5 — the measured `ShouldAlert` → `DecisionDeny` wiring behind §2
- `spec.md` §C.2 — the matrix requirements this schema implements
- `.moai/reports/t228/plan-measurements.md` M4 — the mutant prototype behind §3.2
- `plan.md` M1-M3 — where each open decision above is settled
