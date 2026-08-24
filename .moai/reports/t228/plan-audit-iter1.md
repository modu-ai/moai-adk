# SPEC Review Report: SPEC-ASTGREP-LANG16-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.68** (Tier L PASS threshold: 0.85)

Auditor: plan-auditor. Bias-prevention M1-M6 active.

**M1 Context Isolation statement.** No author reasoning, prior draft, or conversation history was
consulted. The operator decisions D1 (corpus in scope) and D2 (widest breadth) were taken as
binding scope constraints and audited only for whether the SPEC *serves* them. Every finding below
is derived from the six SPEC artifacts, the three lane evidence reports, and my own measurements
against the tree.

## Measurement attribution

| field | value |
|---|---|
| worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab3e9797d7a928e98a6241e9c44e5b445` |
| `git status --porcelain` | 2 untracked entries only (`.moai/reports/t228/`, `.moai/specs/SPEC-ASTGREP-LANG16-001/`) — no tracked modifications, so every citation below is against committed content |
| tool | `ast-grep 0.40.5` (`sg --version` → `ast-grep 0.40.5`) |
| probe scratch | `/tmp/t228sgt`, `/tmp/t228audit_*.go` (outside the tree; nothing written into the repo) |

Every measurement below was run with cwd = worktree root, per the R9 hygiene rule the SPEC itself
records.

---

## Must-Pass Results

- **[FAIL] MP-1 REQ number consistency.** The requirement sequence is not contiguous.
  `grep -o "REQ-A16-[0-9a-z]*" spec.md | sort -u` returns 26 tokens spanning `001`…`025` **plus**
  `005a`, and `REQ-A16-017` appears **only** inside the HISTORY row at `spec.md:26`
  ("REQ-A16-017 merged into REQ-A16-016 … its number is retired rather than reused") — there is no
  `REQ-A16-017` definition anywhere in §C. So the layer carries one gap (`017`) and one
  non-sequential insertion (`005a`). The retirement is documented and deliberate, and the fix is a
  renumber to a contiguous `001`…`025`; MP-1 is nonetheless mechanical and it fails.
- **[FAIL] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer in
  `spec.md` §C only.** (The `AC-A16-*` entries in `acceptance.md` are Given-When-Then by design and
  were graded under Group 4, not here.) 24 of 25 requirement entries match a GEARS pattern.
  **`REQ-A16-011` (`spec.md:287-289`) does not.** Verbatim: *"**Every** promotion to severity
  `error` — without exception, for every rule, in every family, in every language — requires
  **both**: 1. … 2. …"*. It is labelled `(Where)` but carries no `Where …` clause, no named
  `<subject>`, and no `shall` — it is a predicate definition, not a requirement. It is also the
  single most load-bearing requirement in the document (the severity gate). Remediable by
  rewording, e.g. *"Where a rule is promoted to severity `error`, the promotion shall satisfy
  both: …"*.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present with correct types
  (`spec.md:1-16`): `id` / `title` (quoted) / `version` `"0.3.0"` (quoted semver) / `status: draft`
  (enum) / `created: 2026-08-24` / `updated: 2026-08-24` (ISO) / `author` / `priority: P2` (enum) /
  `phase: "v3.2.0 target"` (a release target, not a prohibited lifecycle stage) / `module` /
  `lifecycle: spec-anchored` (enum) / `tags` (comma-separated string). No rejected snake_case alias
  (`created_at` / `updated_at` / `labels` / `spec_id`) is present. The id `SPEC-ASTGREP-LANG16-001`
  is multi-segment; I verified it against the **implemented** validator rather than the prose copy:
  `internal/spec/lint.go:715` → `specIDPattern = regexp.MustCompile(`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)`,
  which admits it. (Observation, not a defect of this SPEC: the regex quoted in
  `spec-frontmatter-schema.md` is the single-segment form and is stale against that code.)
  Optional `tier: L` present. `related_specs:` is not a documented optional field — see D17.
- **[PASS] MP-4 language neutrality.** The SPEC is multi-language by subject and enumerates all 16
  with equal weight (`spec.md:54-87`, REQ-A16-001/002/003). The two exclusions are version-scoped
  tool facts, and I re-measured them rather than trusting the record:
  `echo foo | sg run -p foo -l r --stdin` → `error: invalid value 'r' … r is not supported!` (exit 2);
  `-l dart` → `error: invalid value 'dart' … dart is not supported!` (exit 2). No language is
  labelled primary, planned, or unsupported anywhere in the SPEC's own prose. (The *acceptance
  criterion* that guards this at close is defective — see D9 — but the SPEC itself does not create a
  neutrality violation.)
- **[FAIL] MP-5 D7 cross-SPEC reconciliation — BLOCKING finding, unresolved.** Extraction:
  `grep -rEo "SPEC-([A-Z][A-Z0-9]+-)+[0-9]+" spec.md | sort -u` → `SPEC-ASTG-UPGRADE-001`,
  `SPEC-ASTGREP-DOGFOOD-CLEANUP-001`, `SPEC-ASTGREP-MULTILANG-001` (plus self). Statuses measured:
  MULTILANG-001 `completed`, DOGFOOD-CLEANUP-001 `completed`, **ASTG-UPGRADE-001 `archived`**.
  `SPEC-ASTG-UPGRADE-001` is referenced in three places (`spec.md:15` frontmatter `related_specs`,
  `spec.md:473` "sibling ast-grep SPECs", `research.md:106` "ast-grep version work") and **no
  reconciliation** appears near any of them — no "supersede", "absorb", "carve-out", "reversal", and
  no acknowledgement that it is archived. This is not merely formal: R6 (version drift, `spec.md:452`)
  and the entire `0.40.5` pinning depend on ast-grep version work, and the SPEC names an **archived**
  SPEC as that work's home, leaving R6's re-probe obligation without an owner. Folded into
  `## Defects Found` at severity=critical as **D3**.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -rn "syscall" .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1, no match. D8-4 auto-PASS. No BLOCKING finding.
- **[PASS] MP-7 clarification gate.** `grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1, no match, across all six artifacts including `plan.md` and `research.md`.

**Three must-pass failures (MP-1, MP-2, MP-5). Any one forces FAIL; the aggregate score is
independently below the Tier L threshold.**

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | between 0.50 and 0.75, nearer 0.75 | Prose is unusually precise and every figure names its tree (`spec.md` §A.0). Against that: `REQ-A16-011` is not a requirement in form (MP-2); the checker's cell-identity semantics are unspecified (D7) — two engineers would build different checkers; `REQ-A16-022` (`spec.md:359-368`) fuses a deliverable with an unobservable process obligation (D15); `AC-A16-015`'s scan target collides with existing correct wording (D9). Four interpretation-divergent items exceeds the 0.75 anchor's "one or two". |
| Completeness | 0.75 | 0.75 | All required sections present: HISTORY (`spec.md:20-26`), §A Context, §B Scope, §C Requirements, §D Exclusions with three `### Out of Scope — <topic>` H3 sub-headings each carrying `-` bullets (`spec.md:391-429`), §E Success criteria. Frontmatter complete (MP-3). Tier L 5-artifact set present and all five substantive. Deductions: `REQ-A16-025` has no acceptance criterion and is explicitly excluded from verification (D6); `REQ-A16-017` retired leaving a numbering hole (D4); the `make build` artifact is never named (D12). |
| Testability | 0.60 | between 0.50 and 0.75 | Most criteria are binary and command-shaped. Against that: `AC-A16-011` is tautological — I measured `security/web.yml` `sec-template-injection-html` → `severity: error`, and `spec.md` §A.3 already asserts `error`, so the criterion holds on the untouched tree with zero work (D11). Five further criteria (AC-015/018/019/020/021) also hold today, measured (D11). `AC-A16-025`'s second clause ("never cited as evidence") is a discipline statement with no observable check. `AC-A16-023`'s detection path depends on a human reading a `t.Skip` message that `go test` prints only under `-v` (D1). `AC-A16-006/007` can only verify the *form* of a citation, never its truth. `AC-A16-002` is satisfied by a transient observation that leaves no artifact in the final tree (acknowledged in its own text, but it means the verifier must trust a transcript). |
| Traceability | 0.70 | between 0.50 and 0.75, nearer 0.75 | 24 of 25 requirements map to at least one criterion. **`REQ-A16-025` (the twelve fixtures shall not be modified or renamed) has none**, and `acceptance.md` §G *affirmatively excludes* it from verification (D6). One dangling reference in the reverse direction: `plan.md:255` cites `REQ-A16-022a`, which does not exist — `grep -rn "REQ-A16-022a"` returns exactly that one line and no definition (D10). Plus the `017` gap (D4). |

Aggregate = harmonic mean of (0.70, 0.75, 0.60, 0.70) = **0.68**. Tier L threshold 0.85.

---

## What I verified independently (and what it changed)

Recorded because several of these *confirmed* the SPEC and two *refuted* it. The confirmations are
as load-bearing as the refutations.

**Confirmed — the severity→deny claim is correct, and if anything understated (audit point 4).**
`internal/hook/pre_tool.go:671` calls `h.scanner.ShouldAlert(result)` and returns `DecisionDeny`
with the finding report as reason; `internal/hook/security/scanner.go:166` delegates to
`reporter.ShouldExitWithError`, and `internal/hook/security/reporter.go:112-117` is
`return result.ErrorCount > 0`. Warning-only results fall through to the allow path
(`pre_tool.go:679-686`). The scanned config resolves to the project's
`.moai/config/astgrep-rules/sgconfig.yml` (`internal/hook/security/rules.go:40`), which is exactly
where the template deploys — so a template severity promotion refuses writes in every user project.
I then tested the specific claim about `sec-csrf-no-token-check`. Its pattern
(`security/web.yml`) is `func $HANDLER(w http.ResponseWriter, r *http.Request) {\n  $$$BODY\n}\n`.
Run against a benign, CSRF-irrelevant handler I wrote — and against a **gofmt tab-indented** one, to
test whether the literal two-space indent in the pattern narrows it — it matched both:

```
$ sg run -p 'func $HANDLER(w http.ResponseWriter, r *http.Request) {\n  $$$BODY\n}\n' -l go /tmp/t228audit_tab.go
/tmp/t228audit_tab.go:5:func Handler(w http.ResponseWriter, r *http.Request) {
```

The pattern is AST-matched, not text-matched, so formatting does not narrow it. The SPEC's central
severity argument stands, and the only correction is in the *understating* direction: the match is
formatting-independent, so "every Go HTTP handler" is literally true for any handler with a
non-empty body. (An empty-bodied handler did not match — immaterial.)

**Confirmed — both `sg test` mutants are real (audit points 2 and 3).** I did not take M4 on trust.
Matches-nothing: rule `neverEverCalled($X)` against an invalid case containing no such call →
`[Missing] Expect rule probe-nothing to report issues, but none found in:` … `FAIL`, exit **4**.
Matches-everything: rule `func $F($$$P) { $$$B }` with a benign `valid` case →
`[Noisy] Expect probe-everything to report no issue, but some issues found in:` … `FAIL`, exit **4**.
So the `valid` half is genuinely enforced by the tool, and `REQ-A16-008` / `AC-A16-009` /
`AC-A16-012` rest on a capability that exists. The mirror hazard is asserted as a **pair** in the
criteria and not merely in prose — `AC-A16-012` requires both cases for *every* implemented rule
(warning-severity included), and `AC-A16-022` states outright that "The clean half is not optional".
That is correct and I could not break it.

**Confirmed — the corpus exists and currently runs, not skips.**
`go test ./internal/hook/ -run TestScanWriteContentDifferential -v` → `--- PASS` in 1.13s with four
`security scan blocked write operation` lines for `sample.go`, `digest.go`, `run.js`, `run.ts`,
`run.py`. The 12 fixtures are tracked and clean. `plan-measurements.md` M5 is indeed superseded, and
the v2 remeasure is the correct record. I found no text anywhere in the six artifacts asserting the
withdrawn v1 conclusion.

**Confirmed — the inventory arithmetic.** `grep -rc "^id:"` over the two rule directories sums to
**26** (go dir 12 + security 14); go-language security rules = 8, so go = 20; js/py/ts = 2 each.
8 × 14 = 112; 8 × 10 = 80; 32 − 14 = 18 pre-existing gaps for the four covered languages
(`design.md:33`). Every number in §A.1 / §A.3 / §D.1.2 checks out.

**Refuted — see D1 and D2 below.** Two of the SPEC's stated defenses do not hold.

---

## Defects Found (structured defect-list)

D1. **corpus-gate-skips-not-fails** — `spec.md:186-195` (§A.7), `spec.md:454` (R8),
`design.md:227-247` (§6.3, §6.4), `plan.md:264-268` (M8) — **The covered-language validity gate
`t.Skip`s; it does not fail.** The SPEC's stated forcing function does not exist. Measured:
`internal/hook/pre_tool_scan_differential_test.go:242` is `t.Skip(b.String())` — reached whenever a
language in `coveredCorpusLanguages` has no denying fixture. A skip makes `go test` report `ok`, and
because the skip fires *after* the observation loop but *before* the assertion loop
(`:244-259`), it silently disables **all twelve** recorded differential assertions at once. The SPEC
states the opposite in four places: "the gate **will fail** until the rows are promoted … it is what
keeps the existing test passing" (§A.7); "The existing test **failing** is the detection mechanism,
so this risk is **self-announcing rather than silent**" (R8); "the validity gate fails **loudly**
elsewhere … the suite is **red** between the rule landing and the promotion landing" (§6.3).
Consequences, separated because they differ by milestone:
  - **M8 (java, rust) survives by accident, via a different mechanism.** I read the two placeholder
    fixtures: `java_uncovered.java` and `rs_uncovered.rs` both contain an API-key-shaped literal
    (`"sk-abcdefghijklmnopqrstuvwx"`). Once F2/F4 gain java/rust variants at `error`, those rows flip
    to deny against `wantDeny: false` → `t.Errorf` at `:247-250` → genuinely red. So the red exists,
    but it is the **per-row differential assertion**, not the validity gate the SPEC credits.
  - **M9 (the other eight languages) has no forcing function at all, and has a green escape hatch.**
    An implementer who cannot produce a deny fixture for, say, swift can add swift to
    `coveredCorpusLanguages`, watch the entire differential test turn into a skip, and read `ok` from
    CI. `plan.md:397` forbids exactly this move — but the enforcement it names does not exist.
    `AC-A16-025` is the only guard and it is human discipline, not a mechanical check.
  - `spec.md` §D forbids "Restructuring `pre_tool_scan_differential_test.go` beyond adding rows and
    extending `coveredCorpusLanguages`", so the SPEC **forbids converting the skip to a failure**
    while resting its risk register on a failure that does not occur.
  — Severity: **critical** — Class: **blocking** — Required fix: correct §A.7, R8, design §6.3/§6.4
  and plan M8 to state the measured behaviour (`t.Skip`, `:242`), and name the per-row `wantDeny`
  mismatch as the actual M8 forcing function. Then close the M9 hole with one of: (a) narrow the §D
  exclusion to permit changing the validity gate from `t.Skip` to `t.Fatalf` and add a requirement +
  criterion for it; or (b) add an acceptance criterion that mechanically asserts the differential
  test did not skip — e.g. `go test ./internal/hook/ -run TestScanWriteContentDifferential -v` output
  contains `--- PASS` and contains neither `--- SKIP` nor `corpus rejected:`. Option (b) requires no
  harness change and stays inside the current §D exclusion.

D2. **contrived-rule-passes-every-gate** — `spec.md:448` (R2), `spec.md:287-306` (REQ-A16-011),
`plan.md:349-350`, `plan.md:382-384` — **R2's mitigation is empirically false, and I falsified it in
one line.** The SPEC's guard against contrived cell-filling is: *"REQ-A16-008's benign-shape `valid`
case is the guard: a contrived pattern has no plausible benign counterpart to write."* Measured
counterexample — a rule matching an API that does not exist, with a self-consistent pair:

```yaml
id: probe-contrived
language: go
severity: error
rule:
  pattern: zzzNeverRealApi($X)
# valid:   func ok()  { zzzSafeApi(x) }
# invalid: func bad() { zzzNeverRealApi(x) }
```
```
$ sg test --config /tmp/t228sgt/sgconfig.yml
PASS probe-contrived  .U
test result: ok. 1 passed; 0 failed;
EXIT=0
```

The benign same-shape counterpart took one line to write. This rule would: pass `sg test`
(`REQ-A16-009`, `AC-A16-012`); satisfy **both** clauses of `REQ-A16-011` and therefore qualify for
`severity: error`; be recorded IMPLEMENTED in the matrix with its rule-test path as evidence
(`design.md:47`); and, if a deny fixture is authored against the same contrived pattern, pass the
corpus too — the corpus proves a rule fires on a fixture, never that the fixture is realistic code.
Nothing in the requirement or criterion set anchors a rule to a *real-world* dangerous construct.
Under D2 (widest breadth: 80 rules across ten languages the author has not worked in), this is the
highest-probability failure mode in the SPEC, and it is the exact hazard R2 claims to have closed.
`REQ-A16-008`'s "the `invalid` case shall contain the dangerous construct" is the only textual
anchor and it is an unverifiable judgement.
  — Severity: **critical** — Class: **blocking** — Required fix: withdraw R2's claim that the
  benign-counterpart requirement is sufficient, and add the missing anchor using a convention
  **already present in the tree** (so this is a precision fix, not scope expansion): every existing
  security rule carries `metadata.owasp` and `metadata.cwe` (verified in `security/web.yml`), and
  `research.md:98-99` already recommends new rules follow it. Elevate that from a recommendation to
  a requirement + criterion: every IMPLEMENTED security cell's rule shall carry `metadata.cwe`
  naming the weakness class, and its `invalid` case shall instantiate that class in idiomatic code
  for the language.

D3. **archived-spec-referenced-without-reconciliation** — `spec.md:15`, `spec.md:473`,
`research.md:106` — `SPEC-ASTG-UPGRADE-001` carries `status: archived` (measured) and is cited three
times as a live sibling owning "ast-grep version work", with no reconciliation and no
acknowledgement of its status. R6 (`spec.md:452`) makes version drift a standing risk whose
re-probe it delegates to M2, but the SPEC that owned version work is archived — so nothing outside
this SPEC owns the pinned-version question. This is the MP-5 must-pass failure.
  — Severity: **critical** — Class: **blocking** — Required fix: state the archived status at each
  of the three reference sites and reconcile explicitly — either "archived; this SPEC absorbs no part
  of it and pins 0.40.5 independently (R6 owns re-probe at M2)", or drop it from `related_specs:` and
  §H if it is not actually related.

D4. **req-numbering-gap** — `spec.md:26` (HISTORY), `spec.md` §C — `REQ-A16-017` is retired and
exists only in the HISTORY row; `REQ-A16-005a` is a non-sequential insertion. MP-1 failure.
  — Severity: **major** — Class: **blocking** — Required fix: renumber §C to a contiguous
  `REQ-A16-001` … `REQ-A16-025` (which the current 25 entries exactly fill), keep the HISTORY row as
  the audit trail of the merge, and update the four `plan.md` / `design.md` cross-references.

D5. **req-011-not-a-requirement** — `spec.md:287-296` — the severity gate, the most load-bearing
requirement in the document, is written as a predicate definition ("**Every** promotion … requires
**both**") with no `Where` clause, no subject, and no `shall`. MP-2 failure.
  — Severity: **major** — Class: **blocking** — Required fix: reword to
  "**Where** a rule is promoted to severity `error`, the promotion **shall** satisfy both: (1) …
  (2) …", preserving the existing clause text and the `sec-csrf-no-token-check` counterexample note
  verbatim.

D6. **req-025-unverified-and-excluded** — `spec.md:379-382`, `acceptance.md:209-211` (§G) — the
requirement protecting the recorded corpus baseline ("The twelve fixtures … shall not be modified or
renamed, and their `scanCorpus` rows shall not change verdict") has **no acceptance criterion**, and
§G affirmatively excludes it: *"The recorded verdicts of the twelve pre-existing fixtures — they are
a baseline this SPEC preserves rather than re-verifies."* `plan.md:304-306` (M10 step 3) asks the
implementer to confirm it but cites `AC-A16-021`, which is about the dogfood tree, not the corpus.
So the single guard against "make the gate pass by weakening a recorded verdict" — named as an
anti-pattern at `plan.md:391-393` — is unverified. This matters more given D1: with the validity
gate silent, editing a `wantDeny` column is a live way to turn red green.
  — Severity: **major** — Class: **blocking** — Required fix: replace §G's first exclusion with an
  acceptance criterion. It is one command: `git diff --stat origin/main -- internal/hook/security/testdata/scan-corpus/`
  shows changes only to the two files `REQ-A16-024` sanctions, and `git diff origin/main -- internal/hook/pre_tool_scan_differential_test.go`
  shows no `wantDeny` value changed on a pre-existing row.

D7. **checker-checks-cardinality-not-identity** — `design.md:175-193` (§5), `spec.md:253-258`
(REQ-A16-006), `acceptance.md:42-64` (AC-004/005/007) — the checker's three failure classes are
*count mismatch*, *unresolved cell*, and *unevidenced exemption*. Two holes follow:
  - **The pair set is never checked.** `REQ-A16-004` demands "exactly one cell for every (security
    family, parseable language) pair", but the checker's class 1 is "cell count ≠ 8 × 14" — pure
    cardinality. A matrix with `F1/rust` duplicated and `F1/java` missing passes class 1, passes
    class 2 (both present cells resolve), and passes `AC-A16-016` (whose row check is also by count:
    "it has exactly 14 cells"). Answering the audit's question directly: states (c) unevidenced
    exemption and (d) silently missing *are* separated by the taxonomy — but (d) is only detectable
    through a count, so a **substituted** cell collapses into "no failure at all", which is worse
    than the two collapsing into one.
  - **Matrix→ruleset drift is not closed.** `design.md:17-18` states the risk and claims
    "AC-A16-012 closes that by requiring every cell's named rule to exist and pass `sg test`". As
    written, `AC-A16-012` runs `sg test` over the ruleset; `sg test` never reads the matrix, and no
    checker class covers "cell names a rule id absent from the ruleset". A cell naming a rule that
    does not exist passes every stated check.
  — Severity: **major** — Class: **blocking** — Required fix: change class 1 from a count to a set
  comparison (the cell key set equals the Cartesian product of the two axes — catches missing,
  duplicate, and substituted cells in one class), and add a fourth class "named rule id not found in
  the ruleset", with `AC-A16-007` extended to four synthetic defects.

D8. **at-ceiling-with-unestimated-breadth — split required** — `plan.md:65-91` (§B.0),
`progress.md:8-17`, `acceptance.md:6` — Answering audit point 6 directly: **I believe this must
split, and the seam the SPEC itself names is the right one.** Three measured reasons, not a
preference:
  - **Zero headroom.** 25 requirements and 25 criteria against a Tier L ceiling of 25/25 — and the
    count only reaches 25 by way of a sub-numbered `005a` and a retired `017` (D4). That numbering
    shape is the signature of scope compressed to fit rather than sized. The concrete consequence:
    the corrections this audit requires (D1 needs one criterion, D2 needs one requirement plus one
    criterion, D6 needs one criterion, D7 needs one criterion) **cannot be made without breaching the
    ceiling**. A SPEC that cannot absorb its own audit findings is over-budget by definition.
  - **M4-M7 are unestimated, and the SPEC says so.** `pattern-feasibility-probe.md:38` states
    plainly that the probe is "6개 언어 × 1패턴 표본이며, 10개 언어 × 8패밀리 전수의 근거가 아닙니다"
    — a 6-of-80 sample. `research.md:166` (Q2) defers "which cells have no equivalent construct" to
    M4-M7 "per cell, with evidence". So the implement-vs-exempt split across 80 cells — the single
    largest driver of effort — is unknown at plan time. `plan.md`'s "up to 24 cells" per milestone is
    an upper bound with no lower bound.
  - **10 milestones against a 5-6 norm**, disclosed honestly at §B.0 but not acted on, with
    ~90 `sg test` pairs, up to 80 rules, 20 fixture pairs and a 112-cell document in one card.
  — Severity: **major** — Class: **blocking** — Required fix: take the seam `plan.md:88-91` already
  identifies. **SPEC A = M1-M3** (harness + id keying + matrix/checker + severity reclassification
  over the existing 26) — this is the contract, it is where every reviewable decision lives, it fits
  a Tier L budget with room for D1/D2/D6/D7's added requirements, and it delivers standalone value
  (the existing 26 rules gain tests and a correct severity assignment). **SPEC B = M4-M10** (breadth
  + corpus + close), consuming A's fixed contract, with its own budget for the 80-cell axis. This
  split does not trim scope: both halves remain, and D1/D2 make A's contract stronger before 80 rules
  are written against it.

D9. **ac-015-collides-with-existing-correct-wording** — `acceptance.md:118-123` — the criterion
scans every file under `internal/template/templates/**` for `primary`, `planned`, `unsupported` "or
an equivalent applied to a language name", and requires no match outside the excluded-languages
record. Measured on the untouched tree:

```
$ grep -rniE "primary|planned|unsupported" internal/template/templates/.moai/config/astgrep-rules/
sgconfig.yml:10:# future addition, never an unsupported one. New languages are added by shipping
```

That hit is the equal-opportunity idiom the SPEC elsewhere requires the exclusion record to
**inherit** (`spec.md:85-87`, `research.md:82-87`). So `AC-A16-015` as written fails on correct
content today, and the excluded-languages record it exempts lives under `.moai/specs/`, not under the
template tree it scans. The predictable implementer response under deadline is to weaken the grep
until it matches nothing — which is the failure this criterion exists to prevent.
  — Severity: **major** — Class: **blocking** — Required fix: restate the criterion as a scan for
  ranking *applied to a language name* with the `sgconfig.yml` equal-opportunity paragraph named as a
  sanctioned exemption, and state the exact command and its expected output so the check is
  reproducible rather than re-derived.

D10. **dangling-req-reference** — `plan.md:255` — cites `REQ-A16-022a`, which does not exist;
`grep -rn "REQ-A16-022a"` returns that one line and no definition. The intended target is
`REQ-A16-022`'s second sentence.
  — Severity: **minor** — Class: **blocking** — Required fix: change the citation to `REQ-A16-022`.

D11. **criteria-that-pass-on-the-untouched-tree** — `acceptance.md` — six of 25 criteria hold today
with zero implementation. Measured:
  - `AC-A16-011` (F8 severity agrees with §A.3) — `security/web.yml` `sec-template-injection-html`
    → `severity: error`, and `spec.md:102` already asserts `error`. **Tautological**: it verifies the
    SPEC agrees with a tree neither is changing. It implements nothing.
  - `AC-A16-018` (no SPEC-ID in the template ruleset) — `grep -rlniE "SPEC-[A-Z]"` over the rules dir
    → 0 hits today.
  - `AC-A16-021` (dogfood not mirrored) — true today by construction.
  - `AC-A16-019` (`sg scan --config` completes) — already measured working (`plan-measurements.md` M3).
  - `AC-A16-020` (`make build` leaves no further diff) — true on any untouched tree.
  - `AC-A16-015` — would "pass" only if its grep is weakened; see D9.
  Four of these (018/019/020/021) are legitimate regression guards and should stay. `AC-A16-011` is
  not a guard of anything and consumes one of 25 ceiling slots that D1/D6/D7 need.
  — Severity: **minor** — Class: **blocking** — Required fix: retire `AC-A16-011` (fold the §A.3
  correction into the HISTORY record where it already lives) and reuse the slot for the D1
  no-skip criterion.

D12. **make-build-artifact-never-named** — `spec.md:338-340` (REQ-A16-016), `acceptance.md:149-153`
(AC-A16-020), `plan.md:301-303` — both require the commit to "carry the regenerated embedded
artifacts produced by `make build`", but never say what that artifact is. It is not obvious:
`internal/template/embed.go` uses `//go:embed all:templates`, which generates **no** file, so a
reader can reasonably conclude there is nothing to carry. The actual artifact is
`internal/template/catalog.yaml`, regenerated by `Makefile:24`
(`go run ./internal/template/scripts/gen-catalog-hashes.go --all`). Omitting the catalog is a
recorded recurring CI-parity failure in this repo.
  — Severity: **minor** — Class: **optional** — Required fix: name `internal/template/catalog.yaml`
  explicitly in REQ-A16-016 and AC-A16-020.

D13. **rule-tests-ship-to-every-user-project** — `plan.md:23-29` (§A.2), `design.md:151-158` (§4) —
`rule-tests/` and its snapshots are placed under `internal/template/templates/.moai/config/astgrep-rules/`,
which is a template-managed root: everything under it is deployed into every user project's
`.moai/config/astgrep-rules/`. At the SPEC's own volume estimate (`plan.md:79`, "up to ~90" case
pairs) that ships ~90 fixture files plus snapshots into every user project, and the deployed
`sgconfig.yml` carries a `testConfigs.testDir` pointing at them. A second interaction I did **not**
assess and am recording as a gap rather than a claim: those fixtures contain deliberately dangerous
constructs by construction (`REQ-A16-008`), and they land inside the user's project tree where the
same ruleset scans. The SPEC does not consider whether test assets belong in the distributed surface
at all.
  — Severity: **minor** — Class: **optional** — Required fix: state the decision explicitly in §B
  Scope — either "rule-tests ship with the ruleset, and here is why that is acceptable", or place
  `rule-tests/` outside the template tree with `testConfigs` wired repo-side.

D14. **neutrality-guard-scoped-to-message-fields-only** — `spec.md:346-347` (REQ-A16-019),
`acceptance.md:141-147` (AC-A16-018) — the English-only obligation binds "a non-English **rule
message**", and the criterion reads "every `message:` / `note:` field in the ruleset". The SPEC
simultaneously creates a large new body of hand-authored **fixture** files under the template tree
(D13), whose comments and identifiers are outside both scopes. Given that this repo's own template
tree has previously carried mixed-locale content, the gap is not theoretical.
  — Severity: **minor** — Class: **optional** — Required fix: widen REQ-A16-019 from "rule message"
  to "any human-language text", and widen AC-A16-018's scan from `message:`/`note:` fields to all
  files under the ruleset directory including `rule-tests/`.

D15. **unobservable-process-obligation** — `spec.md:364-366` — "Before authoring any new fixture,
the implementer **shall read** the twelve existing fixtures and match their content shape". Reading
cannot be observed; only the resulting shape can. This is a HOW obligation in a requirement layer
(RQ-3) and it is untestable as written.
  — Severity: **minor** — Class: **optional** — Required fix: state the observable half only — new
  fixtures shall match the existing minimal-example shape (size, structure) — and move the "read them
  first" instruction to `plan.md` M8, where it already appears.

D16. **cwd-drift residue inside the template write surface** — measured, not cited by any artifact:
`internal/template/templates/.moai/config/astgrep-rules/security/.claude/agent-memory/manager-spec/`
exists as an empty directory tree. It is untracked (empty dirs are invisible to git, which is why
`git status --porcelain` is clean), and it is a physical artifact of the same cwd-drift incident
`research.md` §8 records — independent corroboration of that incident, sitting inside the exact
directory this SPEC will write ~90 files into. If any file is ever written there it ships to users.
  — Severity: **minor** — Class: **optional** — Required fix: `rmdir` the stray tree before
  run-phase, and add it to the `plan.md` §A.3 PRESERVE-adjacent hygiene note.

D17. **undocumented frontmatter field** — `spec.md:15` — `related_specs:` is not among the documented
optional fields in `spec-frontmatter-schema.md` (which lists `issue_number`, `depends_on`,
`lint.skip`, `bc_id`, `amendment_of`, `tier`). It is harmless to the 12-field schema check (MP-3
passes) but it is not a recognized field, and `depends_on` is the documented one.
  — Severity: **minor** — Class: **optional** — Required fix: either drop it or use `depends_on` if a
  dependency is actually meant — noting that after D3, one of its three entries is archived.

---

## Regression Check

Not applicable — iteration 1.

---

## Recommendation

FAIL. Three must-pass failures (MP-1, MP-2, MP-5) and an aggregate of 0.68 against a Tier L
threshold of 0.85. Fix in this order — the first two are the ones that change what gets built.

1. **D1 — correct the corpus-gate claim and close the M9 hole.** Four passages state the validity
   gate fails; `pre_tool_scan_differential_test.go:242` is `t.Skip`. Restate them against the measured
   behaviour, credit the per-row `wantDeny` mismatch as M8's real forcing function, and add a
   criterion that mechanically asserts the differential test did not skip (option (b) in D1 needs no
   harness change and stays inside the current §D exclusion).
2. **D2 — withdraw R2's mitigation and add the missing anchor.** The benign-counterpart requirement
   does not stop a contrived rule; I built one that passes `sg test` at exit 0 and qualifies for
   `error`. Bind rules to a real weakness class using the `metadata.cwe` convention the tree already
   carries and `research.md:98` already recommends.
3. **D3 — reconcile `SPEC-ASTG-UPGRADE-001` (archived) at all three reference sites.** This clears
   MP-5 and, more importantly, resolves who owns the pinned-version question that R6 depends on.
4. **D8 — split at the M3/M4 seam the SPEC already names.** SPEC A = M1-M3 (contract), SPEC B =
   M4-M10 (breadth + corpus). The decisive argument is not the milestone count: it is that the
   requirement and criterion budgets are exactly saturated, so the fixes in items 1, 2, 5 and 6
   **cannot be added without breaching the Tier L ceiling** — and that the implement-vs-exempt split
   across 80 cells is explicitly unknown at plan time (`pattern-feasibility-probe.md:38`,
   `research.md` Q2). The split is not a trim; both halves survive, and A's contract gets stronger
   before 80 rules are written against it.
5. **D7 — make the checker compare the cell key set against the Cartesian product**, not a count, and
   add a fourth failure class for a cell naming a rule id absent from the ruleset (the drift
   `design.md:17` names and does not close).
6. **D6 — give `REQ-A16-025` a criterion** and delete the §G exclusion that currently forbids
   verifying it. One `git diff --stat` closes it. This matters more once D1 is understood.
7. **D9 — repair `AC-A16-015`**, which fails today on the equal-opportunity wording the SPEC
   elsewhere requires to be inherited. Name the sanctioned exemption and the exact command.
8. **D4, D5, D10, D11** — mechanical: renumber `REQ-A16-001`…`025` contiguously; reword
   `REQ-A16-011` into `Where … shall …` form; fix `plan.md:255`'s `REQ-A16-022a`; retire the
   tautological `AC-A16-011` and reuse the slot.
9. **D12-D17 (optional class)** — surfaced for the orchestrator's discretion, not required before a
   re-audit verdict. D12 (name `catalog.yaml`) and D13 (decide whether `rule-tests/` ships to users)
   are the two worth taking; the rest are hygiene.

**On what the SPEC gets right**, since a FAIL should not obscure it. The measurement-citation
convention (§A.0) is the reason I could re-verify everything cheaply, and it caught its own stale
base SHA. The §A.6 corpus re-measurement and the `research.md` §8 cwd-drift incident record are
exemplary — I confirmed the corpus independently and found no residue of the withdrawn v1 conclusion
anywhere in the six artifacts. The central severity argument is not overstated: I reproduced the
`sec-csrf-no-token-check` shape match against a benign gofmt-formatted handler, traced
`ShouldAlert` → `ErrorCount > 0` → `DecisionDeny` in code, and confirmed the scanned config is the
one the template deploys. Both `sg test` mutants are real (`[Missing]` and `[Noisy]`, exit 4), and
the matches-everything mirror hazard — the one this audit expected to find dropped — is asserted as
a pair in `AC-A16-012` and `AC-A16-022` rather than left in prose. Every inventory number in §A
checks out. The defects above are in the defenses' *reach*, not in the SPEC's honesty.
